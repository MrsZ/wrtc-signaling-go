package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// an instance of a hub to work with
var hub = NewHub()

func main() {

	// start the hub
	go hub.Start()

	// top level router
	r := mux.NewRouter()

	// signaling
	r.HandleFunc("/signaling", SignalingHandler)

	// static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// attach router
	http.Handle("/", r)

	// start listenting
	err := http.ListenAndServe(":1337", nil)
	if err != nil {
		log.Fatal(err)
	}

	// stop the hub
	hub.Stop()

}
