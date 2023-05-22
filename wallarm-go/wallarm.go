// Package wallarm implements the Wallarm v2 API.
package wallarm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// ErrExistingResource is returned when resource was created other than Terrafom ways - directly via the API.
var ErrExistingResource = errors.New("This resource has already been created earlier")

// ErrInvalidCredentials is raised when not all the credentials are presented.
var ErrInvalidCredentials = errors.New("Credentials are not set. Specify Token or Pair of Secret and UUID")

// New creates a new Wallarm API client.
func New(opts ...Option) (API, error) {

	api, err := newClient(opts...)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func newClient(opts ...Option) (API, error) {
	silentLogger := log.New(ioutil.Discard, "", log.LstdFlags)
	defaultUserAgent := "Wallarm-go/" + Version

	api := &api{
		baseURL: apiURL,
		headers: make(http.Header),
		retryPolicy: RetryPolicy{
			MaxRetries:    3,
			MinRetryDelay: time.Duration(1) * time.Second,
			MaxRetryDelay: time.Duration(30) * time.Second,
		},
		logger: silentLogger,
		Mutex:  &sync.Mutex{},
	}

	if err := api.parseOptions(opts...); err != nil {
		return nil, errors.Wrap(err, "options parsing failed")
	}

	// Fall back to http.DefaultClient if the package user does not provide
	// their own.
	if api.httpClient == nil {
		api.httpClient = http.DefaultClient
		api.UserAgent = defaultUserAgent
	}

	return api, nil
}

// makeRequest makes an HTTP request and returns the body as a byte slice,
// closing it before returning. params will be serialized to JSON or string for GET query.
func (api *api) makeRequest(method, uri, reqType string, params interface{}) ([]byte, error) {
	return api.makeRequestContext(context.TODO(), method, uri, reqType, params)
}

func (api *api) makeRequestContext(ctx context.Context, method, uri, reqType string, params interface{}) ([]byte, error) {
	// Replace nil with a JSON object if needed
	var (
		jsonBody []byte
		err      error
		resp     *http.Response
		respErr  error
		reqBody  io.Reader
		respBody []byte
	)

	if params != nil {
		if _, ok := params.(string); ok {
			jsonBody = nil
		} else if paramBytes, ok := params.([]byte); ok {
			jsonBody = paramBytes
		} else {
			jsonBody, err = json.Marshal(params)
			if err != nil {
				return nil, err
			}
		}
	} else {
		jsonBody = nil
	}

	for i := 0; i <= api.retryPolicy.MaxRetries; i++ {
		if jsonBody != nil {
			reqBody = bytes.NewReader(jsonBody)
		}

		if i > 0 {
			// expect the backoff introduced here on errored requests to dominate the effect of rate limiting
			// don't need a random component here as the rate limiter should do something similar
			// nb time duration could truncate an arbitrary float. Since our inputs are all ints, we should be ok
			sleepDuration := time.Duration(math.Pow(2, float64(i-1)) * float64(api.retryPolicy.MinRetryDelay))

			if sleepDuration > api.retryPolicy.MaxRetryDelay {
				sleepDuration = api.retryPolicy.MaxRetryDelay
			}
			// useful to do some simple logging here, maybe introduce levels later
			api.logger.Printf("Sleeping %s before retry attempt number %d for request %s %s", sleepDuration.String(), i, method, uri)
			time.Sleep(sleepDuration)

		}

		if query, ok := params.(string); ok {
			q := strings.NewReader(query)
			resp, err = api.request(ctx, method, uri, reqType, reqBody, q)
			respErr = errors.Wrap(err, "could not make a request with get query")
		} else {
			resp, err = api.request(ctx, method, uri, reqType, reqBody, nil)
			respErr = errors.Wrap(err, "could not make a request with JSON body")
		}

		// retry if the server is rate limiting us or if it failed
		// assumes server operations are rolled back on failure
		if respErr != nil || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if respErr == nil {
				respBody, err = ioutil.ReadAll(resp.Body)
				resp.Body.Close()

				respErr = errors.Wrap(err, "could not read response body")

				api.logger.Printf("Request: %s %s got an error response %d: %s\n", method, uri, resp.StatusCode,
					strings.Replace(strings.Replace(string(respBody), "\n", "", -1), "\t", "", -1))
			} else {
				api.logger.Printf("Error performing request: %s %s : %s \n", method, uri, respErr.Error())
			}
			continue
		} else {
			respBody, err = ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			if err != nil {
				return nil, errors.Wrap(err, "could not read response body")
			}
			break
		}
	}

	if respErr != nil {
		return nil, respErr
	}

	specificResourceProcessing := []string{"scanner", "user"}

	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, errors.Errorf("HTTP Status: %d, Body: %s", resp.StatusCode, respBody)
	case resp.StatusCode == http.StatusForbidden:
		return nil, errors.Errorf("HTTP Status: %d, Body: %s", resp.StatusCode, respBody)
	case resp.StatusCode == http.StatusServiceUnavailable,
		resp.StatusCode == http.StatusBadGateway,
		resp.StatusCode == http.StatusGatewayTimeout,
		resp.StatusCode == 522,
		resp.StatusCode == 523,
		resp.StatusCode == 524:
		return nil, errors.Errorf("HTTP Status: %d, Body: %s", resp.StatusCode, respBody)
	case resp.StatusCode == http.StatusBadRequest && (reqType == "node" || reqType == "app") && string(respBody) == `{"status":400,"body":"Already exists"}`:
		return nil, errors.Wrap(ErrExistingResource, fmt.Sprintf("HTTP Status: %[1]v Body: %[2]s", resp.StatusCode, string(respBody)))
	case resp.StatusCode == http.StatusConflict && Contains(specificResourceProcessing, reqType):
		return nil, errors.Wrap(ErrExistingResource, fmt.Sprintf("HTTP Status: %[1]v Body: %[2]s", resp.StatusCode, string(respBody)))
	default:
		return nil, errors.Errorf("HTTP Status: %d, Body: %s", resp.StatusCode, respBody)
	}

	return respBody, nil
}

func (api *api) request(ctx context.Context, method, uri, reqType string, reqBody, query io.Reader) (*http.Response, error) {
	api.Lock()
	defer api.Unlock()

	req, err := http.NewRequestWithContext(ctx, method, api.baseURL+uri, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request creation failed")
	}

	req.Header = api.headers
	if api.UserAgent != "" {
		req.Header.Set("User-Agent", api.UserAgent)
	}

	if req.Header.Get("Content-Type") == "" &&
		(Contains([]string{"POST", "PUT"}, method) && reqType != "userdetails") ||
		(method == "DELETE" && reqType == "ip_rules") {
		req.Header.Set("Content-Type", "application/json")
	} else if method == "GET" {
		req.Header.Del("Content-Type")
	}

	if query != nil {
		q, err := ioutil.ReadAll(query)
		if err != nil {
			return nil, err
		}
		req.URL.RawQuery = string(q)
	}

	resp, err := api.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}

	return resp, nil
}
