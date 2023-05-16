package wallarm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppCreate(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"status": 200,
			"body": {
				"id": 0,
				"clientid": 0,
				"name": "Test",
				"deleted": false
			}
		}`)
	}

	appBody := &AppCreate{
		Name: "Test",
		AppFilter: &AppFilter{
			ID:       0,
			Clientid: 0},
	}

	mux.HandleFunc("/v1/objects/pool/create", handler)
	err := client.AppCreate(appBody)
	assert.NoError(t, err)
}

func TestAppCreate_Duplicated(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(400)
		fmt.Fprint(w, `{
			"status": 400,
			"body": "Already exists"
		}`)
	}

	appBody := &AppCreate{
		Name: "Test",
		AppFilter: &AppFilter{
			ID:       0,
			Clientid: 0},
	}

	mux.HandleFunc("/v1/objects/pool/create", handler)
	err := client.AppCreate(appBody)
	assert.EqualError(t, err, `HTTP Status: 400, Body: {
			"status": 400,
			"body": "Already exists"
		}`)
}

func TestAppDelete(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"status": 200,
			"body": [
				0
			]
		}`)
	}

	appBody := &AppDelete{
		Filter: &AppFilter{
			ID:       0,
			Clientid: 0,
		},
	}

	mux.HandleFunc("/v1/objects/pool/delete", handler)
	err := client.AppDelete(appBody)
	assert.NoError(t, err)
}

func TestAppDelete_NonExistent(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"status": 200,
			"body": [
			]
		}`)
	}

	appBody := &AppDelete{
		Filter: &AppFilter{
			ID:       0,
			Clientid: 0,
		},
	}

	mux.HandleFunc("/v1/objects/pool/delete", handler)
	err := client.AppDelete(appBody)
	assert.NoError(t, err)
}
