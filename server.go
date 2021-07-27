package cservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

type RouteHandler httprouter.Handle

type link struct {
	Ref string `json:"ref"`
	Url string `json:"url"`
}

type response struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
	Links  []link      `json:"_links"`
}

type Controller interface {
	Index(r *http.Request) (interface{}, error)
	Create(r *http.Request) (interface{}, error)
	Read(r *http.Request) (interface{}, error)
	Update(r *http.Request) (interface{}, error)
	Delete(r *http.Request) (interface{}, error)

	SetDB(db *gorm.DB)
}

type route struct {
	method  string
	path    string
	handler RouteHandler
}

type server struct {
	server *http.Server
	routes []*route
}

type iserver interface {
	Get(string, RouteHandler)
	Post(string, RouteHandler)
	Put(string, RouteHandler)
	Patch(string, RouteHandler)
	Delete(string, RouteHandler)

	Resource(string, Controller)

	// Start the server.
	Start()

	// Stop the server gracefully.
	Stop()
}

func sendResponse(rw http.ResponseWriter, response interface{}) {
	rw.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(rw).Encode(response); err != nil {
		panic(err)
	}
}

// TODO can we make these handlers
func (s *server) Resource(path string, controller Controller) {
	controller.SetDB(db)

	// GET path
	s.Get(path, func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		res, err := controller.Index(r)

		if err != nil {
			sendResponse(rw, response{
				Error:  err.Error(),
				Status: false,
			})
			return
		}

		sendResponse(rw, response{
			Data:   res,
			Status: true,
		})
	})

	// POST path
	s.Post(path, func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		res, err := controller.Create(r)

		if err != nil {
			sendResponse(rw, response{
				Error:  err.Error(),
				Status: false,
			})
			return
		}

		sendResponse(rw, response{
			Data:   res,
			Status: true,
		})
	})

	// GET path/:id

	// PUT path/:id
	// PATCH path/:id

	// DELETE path/:id
}

func (s *server) Get(path string, handler RouteHandler) {
	s.routes = append(s.routes, &route{
		method:  "GET",
		path:    path,
		handler: handler,
	})
}

func (s *server) Post(path string, handler RouteHandler) {
	s.routes = append(s.routes, &route{
		method:  "POST",
		path:    path,
		handler: handler,
	})
}

func (s *server) Put(path string, handler RouteHandler) {
	s.routes = append(s.routes, &route{
		method:  "PUT",
		path:    path,
		handler: handler,
	})
}

func (s *server) Patch(path string, handler RouteHandler) {
	s.routes = append(s.routes, &route{
		method:  "PATCH",
		path:    path,
		handler: handler,
	})
}

func (s *server) Delete(path string, handler RouteHandler) {
	s.routes = append(s.routes, &route{
		method:  "DELETE",
		path:    path,
		handler: handler,
	})
}

func (s *server) Start() {
	router := httprouter.New()

	for _, route := range s.routes {
		router.Handle(route.method, route.path, httprouter.Handle(route.handler))
	}

	s.server.Handler = router

	err := s.server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (s *server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := s.server.Shutdown(ctx)

	if err != nil {
		panic(err)
	}
}

func NewServer(port int) iserver {
	srv := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", port), // TODO configure host
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &server{
		server: srv,
	}
}
