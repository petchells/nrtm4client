package rpc

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

// Server needs to know what port to run on
type Server struct {
	r *mux.Router
}

// NewServer creates a new server on `port`
func NewServer() Server {
	r := mux.NewRouter()
	return Server{r}
}

// Serve starts the server
func (s *Server) Serve(port int) {
	http.Handle("/", s.r)
	s.r.Walk(
		func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			ht, err := route.GetPathRegexp()
			log.Println("INFO serving route", ht)
			return err
		},
	)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}

// POSTHandler adds a handler to this server
func (s *Server) POSTHandler(subpath string, handler func(w http.ResponseWriter, r *http.Request)) {
	s.r.HandleFunc(subpath, handler).Methods("POST")
	s.r.HandleFunc(subpath, handler).Methods("OPTIONS")
}

// GETHandler registers a function handler for a GET request
func (s *Server) GETHandler(subpath string, handler func(w http.ResponseWriter, r *http.Request)) {
	s.r.HandleFunc(subpath, handler).Methods(http.MethodGet)
}

// Router returns the router for this server
func (s *Server) Router() *mux.Router {
	return s.r
}
