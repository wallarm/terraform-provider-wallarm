package wallarm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the API client being tested
	client API

	// server is a test HTTP server used to provide mock API responses
	server *httptest.Server
)

func setup(opts ...Option) {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// Wallarm client configured to use test server
	authHeaders := make(http.Header)
	authHeaders.Add("X-WallarmAPI-Token", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	client, _ = New(UsingBaseURL(server.URL), Headers(authHeaders))
}

func teardown() {
	server.Close()
}

func TestClient_Headers(t *testing.T) {
	// it should set default headers
	setup()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", r.Header.Get("X-WallarmAPI-Token"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
	})
	teardown()

	// it should override appropriate default headers when custom headers given
	headers := make(http.Header)
	headers.Set("Content-Type", "application/xhtml+xml")
	headers.Add("X-Random", "a random header")
	setup(Headers(headers))
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", r.Header.Get("X-WallarmAPI-Token"))
		assert.Equal(t, "application/xhtml+xml", r.Header.Get("Content-Type"))
		assert.Equal(t, "a random header", r.Header.Get("X-Random"))
	})
	teardown()

	setup()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "Expected method 'GET', got %s", r.Method)
		assert.Equal(t, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", r.Header.Get("X-WallarmAPI-Token"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
	})
	teardown()

}

func TestClient_RetryCanSucceedAfterErrors(t *testing.T) {
	setup(UsingRetryPolicy(2, 0, 1))
	defer teardown()

	requestsReceived := 0
	// could test any function, using ListLoadBalancerPools
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		// we are doing three *retries*
		if requestsReceived == 0 {
			// return error causing client to retry
			w.WriteHeader(500)
			fmt.Fprint(w, `{
				"status": 500,
				"body": "Timed out connecting to server"
			}`)
		} else if requestsReceived == 1 {
			// return error causing client to retry
			w.WriteHeader(429)
			fmt.Fprint(w, `{
				"status": 429,
				"body": "Rate limiting"
			}`)
		} else {
			// return success response
			fmt.Fprint(w, `{
				"status": 200,
				"body": [
					{
						"type": "cloud_node",
						"id": 10101019292,
						"uuid": "13ef5fsde-01ca-0000-85ac-78aaf987",
						"ip": null,
						"hostname": "k8s",
						"last_activity": null,
						"enabled": true,
						"clientid": 0,
						"last_analytic": null,
						"create_time": 1575272181,
						"create_from": "1.1.1.1",
						"protondb_version": null,
						"lom_version": null,
						"protondb_updated_at": null,
						"lom_updated_at": null,
						"node_env_params": {
							"packages": {}
						},
						"active": false,
						"instance_count": 0,
						"active_instance_count": 0,
						"token": "aaaaaaaaaaaaaaaaaaaaaaaaaa",
						"requests_amount": 0
					},
					{
						"type": "cloud_node",
						"id": 1111111,
						"uuid": "cc735e20-0268-7821-ad5d-263016eeb7aa",
						"ip": null,
						"hostname": "TEST",
						"last_activity": null,
						"enabled": true,
						"clientid": 0,
						"last_analytic": null,
						"create_time": 1586445477,
						"create_from": "1.1.1.1",
						"protondb_version": null,
						"lom_version": null,
						"protondb_updated_at": null,
						"lom_updated_at": null,
						"node_env_params": {
							"packages": {}
						},
						"active": false,
						"instance_count": 0,
						"active_instance_count": 0,
						"token": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
						"requests_amount": 0
					}
				]
			}`)
		}
		requestsReceived++

	}

	mux.HandleFunc("/v2/node", handler)

	_, err := client.NodeRead(0, "all")
	assert.NoError(t, err)
}

func TestClient_RetryReturnsPersistentErrorResponse(t *testing.T) {
	setup(UsingRetryPolicy(2, 0, 1))
	defer teardown()

	// could test any function, using ListLoadBalancerPools
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")

		// return error causing client to retry
		w.WriteHeader(500)
		fmt.Fprint(w, `{
			"status": 500,
			"body": "Timed out connecting to server"
		}`)

	}

	mux.HandleFunc("/v2/node", handler)

	_, err := client.NodeRead(0, "all")
	assert.Error(t, err)
}

func TestUserDetails(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"status": 200,
			"body": {
				"id": 5528351,
				"uuid": "00000000-0000-0000-0000-000000000000",
				"clientid": 0,
				"permissions": [
					"admin"
				],
				"actual_permissions": [
					"view_denylist"
				],
				"mfa_enabled": false,
				"create_by": 9999999999,
				"create_at": 1569398384,
				"create_from": "99.99.99.99",
				"enabled": true,
				"validated": true,
				"username": "testuser@wallarm.com",
				"realname": "Test User",
				"email": "testuser@wallarm.com",
				"phone": "",
				"password_changed": 1569398433,
				"login_history": [
					{
						"time": 1591356099,
						"ip": "1.1.1.1"
					},
					{
						"time": 1591356129,
						"ip": "1.1.1.1"
					},
					{
						"time": 1591356152,
						"ip": "1.1.1.1"
					},
					{
						"time": 1591356197,
						"ip": "1.1.1.1"
					},
					{
						"time": 1591356571,
						"ip": "1.1.1.1"
					},
					{
						"time": 1591710777,
						"ip": "1.1.1.1"
					},
					{
						"time": 1592056196,
						"ip": "1.1.1.1"
					},
					{
						"time": 1592299799,
						"ip": "1.1.1.1"
					},
					{
						"time": 1592918297,
						"ip": "1.1.1.1"
					},
					{
						"time": 1593952677,
						"ip": "1.1.1.1"
					}
				],
				"timezone": "PDT",
				"results_per_page": 25,
				"default_pool": "test",
				"default_poolid": null,
				"last_read_news_id": 11,
				"search_templates": {
					"demo": "attacks incidents last month !falsepositive"
				},
				"notifications": {
					"report_daily": {
						"email": false,
						"sms": false
					},
					"report_weekly": {
						"email": false,
						"sms": true
					},
					"report_monthly": {
						"email": false,
						"sms": true
					},
					"vuln": {
						"email": false,
						"sms": true
					},
					"scope": {
						"email": false,
						"sms": false
					},
					"system": {
						"email": false,
						"sms": true
					},
					"billing": {
						"email": true,
						"sms": false
					}
				},
				"components": [
					"waf",
					"fast"
				],
				"language": "en",
				"last_login_time": 1593952677,
				"date_format": "ddmmyy",
				"time_format": "24h",
				"job_title": null,
				"available_authentications": [
					"default"
				],
				"frontend_url": "https://my.wallarm.com"
			}
		}`)
	}

	mux.HandleFunc("/v1/user", handler)

	userInfo, err := client.UserDetails()
	if assert.NoError(t, err) {
		assert.Equal(t, userInfo.Body.Clientid, 0)
	}

}

func TestErrorFromResponse(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(403)
		fmt.Fprintf(w, `{
			"status": 403,
			"body": "User not authenticated"
		}`)
	}

	mux.HandleFunc("/v1/user", handler)

	_, err := client.UserDetails()
	assert.EqualError(t, err, `HTTP Status: 403, Body: {
			"status": 403,
			"body": "User not authenticated"
		}`)
}
