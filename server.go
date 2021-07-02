package cservice

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type RouteHandler httprouter.Handle

type route struct {
	method  string
	path    string
	handler RouteHandler
}

type server struct {
	routes []*route
}

type iserver interface {
	Get(string, RouteHandler)
	Post(string, RouteHandler)
	Put(string, RouteHandler)
	Patch(string, RouteHandler)
	Delete(string, RouteHandler)

	Resource(string, interface{})

	Start(port int)
}

func (s *server) Resource(path string, model interface{}) {
	log.Fatal("Not Yet Implemented")
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

func (s *server) Start(port int) {
	router := httprouter.New()

	for _, route := range s.routes {
		router.Handle(route.method, route.path, httprouter.Handle(route.handler))
	}

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		panic(err)
	}
}

func NewServer() iserver {
	return &server{}
}
