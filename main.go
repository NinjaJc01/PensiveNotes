package main

import (
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	go startServer()
}
func startServer() {
	portPtr := flag.Int("p", 8081, "Port number to run the server on")
	flag.Parse()
	port := *portPtr
	mr := mux.NewRouter()
	mr.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	apiRouter := mr.PathPrefix("/api").Subrouter()
	go mime.AddExtensionType(".css", "text/css; charset=utf-8")
	go mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	//Setup a static router for HTML/CSS/JS
	mr.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./resources"))))
	//CRUD API routes
	aRouter := apiRouter.PathPrefix("/something").Subrouter()
	/*A route		*/ aRouter.HandleFunc("/{id}", reqHandler).Methods("POST")
	/*Another Route	*/ aRouter.HandleFunc("/list", reqHandler).Methods("GET")
	log.Println("Listening for requests")
	http.ListenAndServe(fmt.Sprintf(":%v", port), mr)
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}
func notFoundHandler(w http.ResponseWriter, r *http.Request) { //Handle 404s
	w.WriteHeader(404)
}
