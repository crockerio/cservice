package cservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

type RouteHandler httprouter.Handle

type link struct {
	Ref string `json:"ref"`
	Url string `json:"url"`
}

type Response struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error"`
	Links  []link      `json:"_links"`
}

type Controller interface {
	Index(r *http.Request) (interface{}, error)
	Create(r *http.Request) (interface{}, error)
	Read(r *http.Request, id int) (interface{}, error)
	Update(r *http.Request, id int) (interface{}, error)
	Delete(r *http.Request, id int) (interface{}, error)

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

	BuildHandler() *httprouter.Router

	// Start the server.
	Start()

	// Stop the server gracefully.
	Stop()
}

func rootResponse(cb func(*http.Request) (interface{}, error)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		res, err := cb(r)

		if err != nil {
			sendResponse(rw, Response{
				Error:  err.Error(),
				Status: false,
			})
			return
		}

		sendResponse(rw, Response{
			Data:   res,
			Status: true,
		})
	}
}

func parameterResponse(cb func(*http.Request, int) (interface{}, error)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		rawId := p.ByName("id")
		id, err := strconv.Atoi(rawId)

		if err != nil {
			sendResponse(rw, Response{
				Error:  err.Error(),
				Status: false,
			})
			return
		}

		res, err := cb(r, id)

		if err != nil {
			sendResponse(rw, Response{
				Error:  err.Error(),
				Status: false,
			})
			return
		}

		sendResponse(rw, Response{
			Data:   res,
			Status: true,
		})
	}
}

func sendResponse(rw http.ResponseWriter, response interface{}) {
	rw.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(rw).Encode(response); err != nil {
		panic(err)
	}
}

// TODO can we make these handlers
func (s *server) Resource(path string, controller Controller) {
	pathWithId := fmt.Sprintf("%s/:id", path)

	controller.SetDB(db)

	// GET path
	s.Get(path, rootResponse(controller.Index))

	// POST path
	s.Post(path, rootResponse(controller.Create))

	// GET path/:id
	s.Get(pathWithId, parameterResponse(controller.Read))

	// PUT path/:id
	// PATCH path/:id
	s.Put(pathWithId, parameterResponse(controller.Update))
	s.Patch(pathWithId, parameterResponse(controller.Update))

	// DELETE path/:id
	s.Delete(pathWithId, parameterResponse(controller.Delete))
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

func (s *server) BuildHandler() *httprouter.Router {
	router := httprouter.New()

	for _, route := range s.routes {
		router.Handle(route.method, route.path, httprouter.Handle(route.handler))
	}

	return router
}

func (s *server) Start() {
	log.Println("Starting server")
	s.server.Handler = s.BuildHandler()
	err := s.server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *server) Stop() {
	log.Println("Attempting to shut down server")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := s.server.Shutdown(ctx)

	if err != nil {
		log.Fatalln(err)
	}
}

func NewServer(port int) iserver {
	log.Printf("Creating new server on port %d", port)

	srv := &http.Server{
		Addr:         fmt.Sprintf("localhost:%d", port), // TODO configure host
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &server{
		server: srv,
	}
}
