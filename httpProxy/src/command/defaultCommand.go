package command

import (
	"net/http"
	"fmt"
	"strings"
	"io"
)

type DefaultCommand struct {
	RemoteHost string
}

func (defaultCommand *DefaultCommand) Execute(responseWriter http.ResponseWriter, request *http.Request) (err error, handled bool) {
	handled = true

	request.URL.Host = defaultCommand.RemoteHost
	request.RequestURI = ""

	if request.URL.Scheme == "" {
		request.URL.Scheme = "http"
	}

	httpClient := http.Client{}
	var response *http.Response
	response, err = httpClient.Do(request)

	if err != nil {
		fmt.Print(err)
		return
	}

	defer response.Body.Close()

	responseWriter.WriteHeader(response.StatusCode)
	writeHeaders(responseWriter, response)

	io.Copy(responseWriter, response.Body)

	return
}

func writeHeaders(writer http.ResponseWriter, response *http.Response) {
	for key, value := range response.Header {
		writer.Header().Set(key, strings.Join(value, ","))
	}
}
