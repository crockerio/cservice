package cservice_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crockerio/cservice"
	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

func TestServerGet(t *testing.T) {
	// Make Server
	server := cservice.NewServer(12340)

	// Make Get Endpoint
	endpointHit := false

	server.Get("/test", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		endpointHit = true
	})

	// Make the Recorder
	rr := httptest.NewRecorder()

	// Hit Endpoint
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	handler := server.BuildHandler()
	handler.ServeHTTP(rr, req)

	// Verify Response
	if !endpointHit {
		t.Errorf("GET endpoint not hit")
	}
}

func TestServerPost(t *testing.T) {
	// Make Server
	server := cservice.NewServer(12340)

	// Make Post Endpoint
	endpointHit := false

	server.Post("/test", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		endpointHit = true
	})

	// Make the Recorder
	rr := httptest.NewRecorder()

	// Hit Endpoint
	req := httptest.NewRequest(http.MethodPost, "/test", nil)

	handler := server.BuildHandler()
	handler.ServeHTTP(rr, req)

	// Verify Response
	if !endpointHit {
		t.Errorf("POST endpoint not hit")
	}
}

func TestServerPut(t *testing.T) {
	// Make Server
	server := cservice.NewServer(12340)

	// Make Put Endpoint
	endpointHit := false

	server.Put("/test", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		endpointHit = true
	})

	// Make the Recorder
	rr := httptest.NewRecorder()

	// Hit Endpoint
	req := httptest.NewRequest(http.MethodPut, "/test", nil)

	handler := server.BuildHandler()
	handler.ServeHTTP(rr, req)

	// Verify Response
	if !endpointHit {
		t.Errorf("PUT endpoint not hit")
	}
}

func TestServerPatch(t *testing.T) {
	// Make Server
	server := cservice.NewServer(12340)

	// Make Patch Endpoint
	endpointHit := false

	server.Patch("/test", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		endpointHit = true
	})

	// Make the Recorder
	rr := httptest.NewRecorder()

	// Hit Endpoint
	req := httptest.NewRequest(http.MethodPatch, "/test", nil)

	handler := server.BuildHandler()
	handler.ServeHTTP(rr, req)

	// Verify Response
	if !endpointHit {
		t.Errorf("PATCH endpoint not hit")
	}
}

func TestServerDelete(t *testing.T) {
	// Make Server
	server := cservice.NewServer(12340)

	// Make Delete Endpoint
	endpointHit := false

	server.Delete("/test", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		endpointHit = true
	})

	// Make the Recorder
	rr := httptest.NewRecorder()

	// Hit Endpoint
	req := httptest.NewRequest(http.MethodDelete, "/test", nil)

	handler := server.BuildHandler()
	handler.ServeHTTP(rr, req)

	// Verify Response
	if !endpointHit {
		t.Errorf("DELETE endpoint not hit")
	}
}

func TestServerResource(t *testing.T) {
	expected := make(map[string]string)
	expected["index"] = "index"
	expected["create"] = "create"
	expected["read"] = "read 1"
	expected["update"] = "update 1"
	expected["delete"] = "delete 1"

	rr := make(map[string]*httptest.ResponseRecorder)
	req := make(map[string]*http.Request)

	// Make Server
	server := cservice.NewServer(12340)

	// Make Endpoints
	server.Resource("/test", &testController{})

	// Make the Recorder
	rr["index"] = httptest.NewRecorder()
	rr["create"] = httptest.NewRecorder()
	rr["read"] = httptest.NewRecorder()
	rr["update"] = httptest.NewRecorder()
	rr["delete"] = httptest.NewRecorder()

	// Hit Endpoints
	req["index"] = httptest.NewRequest(http.MethodGet, "/test", nil)
	req["create"] = httptest.NewRequest(http.MethodPost, "/test", nil)
	req["read"] = httptest.NewRequest(http.MethodGet, "/test/1", nil)
	req["update"] = httptest.NewRequest(http.MethodPut, "/test/1", nil)
	req["delete"] = httptest.NewRequest(http.MethodDelete, "/test/1", nil)

	handler := server.BuildHandler()

	for endpoint, expectedRes := range expected {
		response := rr[endpoint]

		handler.ServeHTTP(response, req[endpoint])

		res := cservice.Response{}
		err := json.NewDecoder(response.Body).Decode(&res)

		if response.Result().StatusCode != 200 {
			t.Errorf("status code of %s endpoint is %d, expected 200", endpoint, response.Result().StatusCode)
		}

		if response.Header().Get("Content-Type") != "application/json" {
			t.Errorf("content type of %s endpoint is %s, expected application/json", endpoint, response.Result().Header.Get("Content-Type"))
		}

		if err != nil {
			t.Errorf("error decoding %s endpoint response: %s", endpoint, err)
		}

		if res.Data.(string) != expectedRes {
			t.Errorf("incorrect response from %s endpoint. Expected %s, got %s", endpoint, expectedRes, res.Data)
		}
	}
}

// TODO test error responses
// TODO test links

// TEST CONTROLLER
type testController struct {
	DB *gorm.DB
}

func (c *testController) SetDB(db *gorm.DB) {
	// Do Nothing
}

func (c *testController) Index(r *http.Request) (interface{}, error) {
	return "index", nil
}

func (c *testController) Create(r *http.Request) (interface{}, error) {
	return "create", nil
}

func (c *testController) Read(r *http.Request, id int) (interface{}, error) {
	return fmt.Sprintf("read %d", id), nil
}

func (c *testController) Update(r *http.Request, id int) (interface{}, error) {
	return fmt.Sprintf("update %d", id), nil
}

func (c *testController) Delete(r *http.Request, id int) (interface{}, error) {
	return fmt.Sprintf("delete %d", id), nil
}
