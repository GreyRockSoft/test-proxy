package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/boltdb/bolt"
	"github.com/go-zoo/bone"
)

type ProxyIteraction struct {
	Id             uuid.UUID `json:"id"`
	IteractionType string    `json:"type"`
	Path           string    `json:"path"`
	Payload        string    `json:"payload"`
	Size           uint32    `json:"size"`
	StatusCode     uint32    `json:"statusCode"`
}

type proxyAdminServer struct {
	Db *bolt.DB
}

func (p *proxyAdminServer) GetProxies(response http.ResponseWriter, request *http.Request) {

	proxies := make([]ProxyIteraction, 0, 20)

	p.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Proxies"))

		if b == nil {
			return nil
		}

		b.ForEach(func(k, v []byte) error {

			var proxy ProxyIteraction

			err := json.Unmarshal(v, &proxy)

			if err != nil {
				return err
			}

			proxies = append(proxies, proxy)

			return nil
		})
		return nil
	})

	response.WriteHeader(200)
	encoder := json.NewEncoder(response)
	err := encoder.Encode(proxies)
	if err != nil {
		log.Printf("Error: %s\n", err.Error())
	}
}

func (p *proxyAdminServer) NewProxyIteraction(response http.ResponseWriter, request *http.Request) {
	var proxy ProxyIteraction
	decoder := json.NewDecoder(request.Body)
	defer request.Body.Close()

	err := decoder.Decode(&proxy)
	if err != nil {

		writeError(response, 400, err)
		return
	}
	err = p.Db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Proxies"))

		if err != nil {
			return fmt.Errorf("create bucket %s", err)
		}

		proxy.Id = uuid.NewRandom()

		buf, err := json.Marshal(&proxy)

		if err != nil {
			return err
		}

		err = b.Put([]byte(proxy.Id), buf)

		return err
	})

	if err != nil {
		response.WriteHeader(400)
		return
	}

	encoder := json.NewEncoder(response)

	response.WriteHeader(202)
	err = encoder.Encode(&proxy)

	if err != nil {
		log.Printf("Error: %s\n", err.Error())
	}
}

func (p *proxyAdminServer) DeleteProxyIteraction(response http.ResponseWriter, request *http.Request) {
	idString := bone.GetValue(request, "id")

	id := uuid.Parse(idString)

	p.Db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("Proxies"))
		err := b.Delete([]byte(id))

		return err
	})

	response.WriteHeader(204)
}

type proxyServer struct {
	Client *http.Client
	Db     *bolt.DB
}

func writeError(writer http.ResponseWriter, statusCode int, err error) {
	log.Printf("Error: %s\n", err.Error())
	writer.WriteHeader(statusCode)
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

	db, err := bolt.Open("proxy.db", 0600, nil)

	if err != nil {
		log.Fatal("Fatal: %s\n", err.Error())
	}

	defer db.Close()

	adminServer := proxyAdminServer{db}

	adminMux := bone.New()
	adminMux.Get("/proxy", http.HandlerFunc(adminServer.GetProxies))
	adminMux.Delete("/proxy/:id", http.HandlerFunc(adminServer.DeleteProxyIteraction))
	adminMux.Post("/proxy", http.HandlerFunc(adminServer.NewProxyIteraction))

	proxyServer := proxyServer{&http.Client{}, db}

	mux := bone.New()

	mux.Handle("/*", http.HandlerFunc(proxyServer.ProxyHandler))

	go func(port string) {
		log.Println("Starting admin server")
		log.Fatal(http.ListenAndServe(port, adminMux))
	}(":9080")
	log.Println("Starting test proxy")
	log.Fatal(http.ListenAndServe(":9090", mux))
}
