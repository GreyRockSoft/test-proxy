package main

import (
    "log"
    "net/http"
    "net/http/httputil"
)

func NewReverseProxy() *httputil.ReverseProxy {
    director := func(req *http.Request) {
        log.Printf("Redirecting to %s\n", req.Host)
        req.URL.Host = req.Host
    }

    return &httputil.ReverseProxy{Director: director}
}

func main(){
    log.Println("Starting test proxy")

    proxy := NewReverseProxy()

    log.Fatal(http.ListenAndServe(":9090", proxy))
}
