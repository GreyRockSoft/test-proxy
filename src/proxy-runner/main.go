package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-zoo/bone"
)

type proxyServer struct {
	Client *http.Client
}

func writeHeaders(writer http.ResponseWriter, response *http.Response) {
	for key, value := range response.Header {
		writer.Header().Set(key, strings.Join(value, ","))
	}
}

func (p *proxyServer) ProxyHandler(response http.ResponseWriter, request *http.Request) {
	log.Printf("Proxying request to: %s\n", request.Host)
	request.URL.Host = request.Host
	request.RequestURI = ""

	if request.URL.Scheme == "" {
		request.URL.Scheme = "http"
	}

	clientResponse, err := p.Client.Do(request)

	if err != nil {
		log.Printf("ERROR - %s\n", err.Error())
		response.WriteHeader(500)
	} else {
		// write over the headers then the response, then the body
		writeHeaders(response, clientResponse)
		response.WriteHeader(clientResponse.StatusCode)

		io.Copy(response, clientResponse.Body)

		if err != nil {
			log.Printf("ERROR - %s\n", err.Error())
		}
	}
	request.Body.Close()
	clientResponse.Body.Close()
}

func main() {
	log.Println("Starting test proxy")

	proxyServer := proxyServer{&http.Client{}}
	mux := bone.New()

	mux.Handle("/*", http.HandlerFunc(proxyServer.ProxyHandler))

	log.Fatal(http.ListenAndServe(":9090", mux))
}
