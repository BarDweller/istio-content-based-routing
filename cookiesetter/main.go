package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func setCookie(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	if env, ok := params["id"]; ok {
		expires := time.Now().AddDate(0, 0, 1) // expire 1 day
		ck := http.Cookie{
			Name:    "Istio-NS-Hint",
			Path:    "/",
			Expires: expires,
		}
		ck.Value = env
		http.SetCookie(w, &ck)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/cookie/{id}", setCookie).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
