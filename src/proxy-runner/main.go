package main

import (
	"log"
	"net/http"

	"github.com/go-zoo/bone"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Proxying request")

}

func main() {
	log.Println("Starting test proxy")

	mux := bone.New()

	mux.Handle("/*", http.HandlerFunc(proxyHandler))

	log.Fatal(http.ListenAndServe(":9090", mux))
}
