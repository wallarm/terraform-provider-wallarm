package wallarm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	expectedNodeRead = NodeRead{
		Status: 200,
		Body: []NodeReadBody{
			{
				NodeReadBodyPOST: &NodeReadBodyPOST{
					Type:              "cloud_node",
					Hostname:          "k8s",
					Enabled:           true,
					Clientid:          0,
					CreateTime:        1575272181,
					CreateFrom:        "1.1.1.1",
					ProtondbUpdatedAt: nil,
					LomUpdatedAt:      nil,
					Active:            false,
				},
				ID:   10101019292,
				UUID: "13ef5fsde-01ca-0000-85ac-78aaf987",
				IP:   nil,

				LastActivity:        nil,
				LastAnalytic:        nil,
				ProtondbVersion:     nil,
				LomVersion:          nil,
				InstanceCount:       0,
				ActiveInstanceCount: 0,
				Token:               "aaaaaaaaaaaaaaaaaaaaaaaaaa",
				RequestsAmount:      0,
			},
			{
				NodeReadBodyPOST: &NodeReadBodyPOST{
					Type:              "cloud_node",
					Hostname:          "TEST",
					Enabled:           true,
					Clientid:          0,
					CreateTime:        1586445477,
					CreateFrom:        "1.1.1.1",
					ProtondbUpdatedAt: nil,
					LomUpdatedAt:      nil,
					Active:            false,
				},
				ID:                  1111111,
				UUID:                "cc735e20-0268-7821-ad5d-263016eeb7aa",
				IP:                  nil,
				LastActivity:        nil,
				LastAnalytic:        nil,
				ProtondbVersion:     nil,
				LomVersion:          nil,
				InstanceCount:       0,
				ActiveInstanceCount: 0,
				Token:               "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				RequestsAmount:      0,
			},
		},
	}
)

func TestNodeReads(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
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

	mux.HandleFunc("/v2/node", handler)
	actual, err := client.NodeRead(0, "all")
	if assert.NoError(t, err) {
		assert.Equal(t, expectedNodeRead, *actual)
	}

}

func TestNodeReads_CloudNodes(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
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

	mux.HandleFunc("/v2/node", handler)
	actual, err := client.NodeRead(0, "cloud_node")
	if assert.NoError(t, err) {
		assert.Equal(t, expectedNodeRead, *actual)
	}

}

func TestNodeReads_IncorrectJSON(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "GET", "Expected method 'GET', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			[]
		}`)
	}

	mux.HandleFunc("/v2/node", handler)
	_, err := client.NodeRead(0, "fast_node")
	assert.EqualError(t, err, "invalid character '[' looking for beginning of object key string")

}

func TestCreateNode(t *testing.T) {
	setup()
	defer teardown()
	expectedCreateNode := NodeCreateResp{
		Status: 200,
		Body: &NodeReadBody{
			NodeReadBodyPOST: &NodeReadBodyPOST{},
		},
	}
	expectedCreateNode.Body.Type = "cloud_node"
	expectedCreateNode.Body.ID = 15049247
	expectedCreateNode.Body.UUID = "e3b8d4db-4924-431f-838e-9ab08ebb0c90"
	expectedCreateNode.Body.IP = nil
	expectedCreateNode.Body.Hostname = "TESTING"
	expectedCreateNode.Body.LastActivity = nil
	expectedCreateNode.Body.Enabled = true
	expectedCreateNode.Body.Clientid = 0
	expectedCreateNode.Body.LastAnalytic = nil
	expectedCreateNode.Body.CreateTime = 1594230993
	expectedCreateNode.Body.CreateFrom = "1.1.1.1"
	expectedCreateNode.Body.ProtondbVersion = nil
	expectedCreateNode.Body.LomVersion = nil
	expectedCreateNode.Body.ProtondbUpdatedAt = nil
	expectedCreateNode.Body.LomUpdatedAt = nil
	expectedCreateNode.Body.Active = false
	expectedCreateNode.Body.InstanceCount = 0
	expectedCreateNode.Body.ActiveInstanceCount = 0
	expectedCreateNode.Body.Token = "fzGbSyMPwo69s3gLQc3fBBg5ODDdXZjNbH+T3ESpHPdsjuNTbSSRDH4OOmrCOuwyQ"
	expectedCreateNode.Body.Secret = "7f319b4b230fc28ebdb3778b43cddf0418393830dg5d98cd6c7f93dc44a91cf0"

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"status": 200,
			"body": {
				"type": "cloud_node",
				"id": 15049247,
				"uuid": "e3b8d4db-4924-431f-838e-9ab08ebb0c90",
				"ip": null,
				"hostname": "TESTING",
				"last_activity": null,
				"enabled": true,
				"clientid": 0,
				"last_analytic": null,
				"create_time": 1594230993,
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
				"token": "fzGbSyMPwo69s3gLQc3fBBg5ODDdXZjNbH+T3ESpHPdsjuNTbSSRDH4OOmrCOuwyQ",
				"secret": "7f319b4b230fc28ebdb3778b43cddf0418393830dg5d98cd6c7f93dc44a91cf0"
			}
		}
		`)
	}

	mux.HandleFunc("/v2/node", handler)

	actual, err := client.NodeCreate(
		&NodeCreate{
			Hostname: "TESTING",
			Type:     "cloud_node",
			Clientid: 0,
		},
	)

	if assert.NoError(t, err) {
		assert.Equal(t, expectedCreateNode, *actual)
	}
}

func TestCreate_DuplicatedNode(t *testing.T) {
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

	mux.HandleFunc("/v2/node", handler)

	_, err := client.NodeCreate(
		&NodeCreate{
			Hostname: "TESTING",
			Type:     "cloud_node",
			Clientid: 0,
		},
	)

	assert.EqualError(t, err, `HTTP Status: 400, Body: {
			"status": 400,
			"body": "Already exists"
		}`)

}

func TestCreate_EmptyResp(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		fmt.Fprint(w, `{
			[]
		}`)
	}

	mux.HandleFunc("/v2/node", handler)

	_, err := client.NodeCreate(
		&NodeCreate{
			Hostname: "TESTING",
			Type:     "cloud_node",
			Clientid: 0,
		},
	)

	assert.EqualError(t, err, `invalid character '[' looking for beginning of object key string`)

}

func TestCreateNode_WithIncorrectType(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(400)
		fmt.Fprintf(w, `{
			"status": 400,
			"body": "Already exists"
		}`)
	}

	mux.HandleFunc("/v2/node", handler)

	_, err := client.NodeCreate(
		&NodeCreate{
			Hostname: "TESTING",
			Type:     "cloud_node",
			Clientid: 0,
		},
	)

	assert.EqualError(t, err, `HTTP Status: 400, Body: {
			"status": 400,
			"body": "Already exists"
		}`)

}

func TestDeleteNode(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "DELETE", "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprintf(w, `{
			"status": 200,
			"body": {
				"type": "cloud_node",
				"id": 15049247,
				"uuid": "e3b8d4db-4924-431f-838e-9ab08ebb0c90",
				"ip": null,
				"hostname": "TESTING",
				"last_activity": null,
				"enabled": true,
				"clientid": 0,
				"last_analytic": null,
				"create_time": 1594230993,
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
				"token": "fzGbSyMPwo69s3gLQc3fBBg5ODDdXZjNbH+T3ESpHPdsjuNTbSSRDH4OOmrCOuwyQ"
			}
		}
		`)
	}

	mux.HandleFunc("/v2/node/15049247", handler)

	err := client.NodeDelete(15049247)
	assert.NoError(t, err)
}

func TestDeleteNode_WithMissingID(t *testing.T) {
	setup()
	defer teardown()
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "DELETE", "Expected method 'DELETE', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(404)
		fmt.Fprintf(w, `{
			"status": 404,
			"body": {
				"error": "node not found",
				"value": 15049247
			}
		}`)
	}

	mux.HandleFunc("/v2/node/15049247", handler)

	err := client.NodeDelete(15049247)
	assert.EqualError(t, err, `HTTP Status: 404, Body: {
			"status": 404,
			"body": {
				"error": "node not found",
				"value": 15049247
			}
		}`)
}

func TestNodeReads_ByFilter(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			"status": 200,
			"body": [
				{
					"type": "cloud_node",
					"id": "13ef5fsde-01ca-0000-85ac-78aaf987",
					"hostname": "TESTING",
					"enabled": true,
					"clientid": 0,
					"create_time": 1575272181,
					"create_from": "1.1.1.1",
					"protondb_updated_at": null,
					"lom_updated_at": null,
					"active": false
				}
			]
		}`)
	}

	mux.HandleFunc("/v1/objects/node", handler)
	expectedNodeReadByFilter := NodeReadPOST{
		Status: 200,
		Body: []NodeReadBodyPOST{
			{
				Type:              "cloud_node",
				ID:                "13ef5fsde-01ca-0000-85ac-78aaf987",
				Hostname:          "TESTING",
				Enabled:           true,
				Clientid:          0,
				CreateTime:        1575272181,
				CreateFrom:        "1.1.1.1",
				ProtondbUpdatedAt: nil,
				LomUpdatedAt:      nil,
				Active:            false,
			},
		},
	}

	actual, err := client.NodeReadByFilter(&NodeReadByFilter{
		Filter: &NodeFilter{
			Hostname: "TESTING",
		},
		Limit:     1000,
		Offset:    0,
		OrderBy:   "id",
		OrderDesc: false,
	})
	if assert.NoError(t, err) {
		assert.Equal(t, expectedNodeReadByFilter, *actual)
	}

}

func TestNodeReads_ByFilterWithError(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(404)
		fmt.Fprint(w, `{
			"status": 404,
			"body": {
				"error": "node not found",
				"value": 15049247
			}
		}`)
	}

	mux.HandleFunc("/v1/objects/node", handler)

	_, err := client.NodeReadByFilter(&NodeReadByFilter{
		Filter: &NodeFilter{
			Hostname: "Non-existent",
		},
		Limit:     1,
		Offset:    0,
		OrderBy:   "id",
		OrderDesc: false,
	})

	assert.EqualError(t, err, `HTTP Status: 404, Body: {
			"status": 404,
			"body": {
				"error": "node not found",
				"value": 15049247
			}
		}`)
}

func TestNodeReads_ByFilterIncorrectJSON(t *testing.T) {
	setup()
	defer teardown()

	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, "POST", "Expected method 'POST', got %s", r.Method)
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{
			[]
		}`)
	}

	mux.HandleFunc("/v1/objects/node", handler)

	_, err := client.NodeReadByFilter(&NodeReadByFilter{
		Filter: &NodeFilter{
			Hostname: "Non-existent",
		},
		Limit:     1,
		Offset:    0,
		OrderBy:   "id",
		OrderDesc: false,
	})

	assert.EqualError(t, err, `invalid character '[' looking for beginning of object key string`)
}
