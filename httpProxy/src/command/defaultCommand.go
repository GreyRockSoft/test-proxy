package command

import (
	"net/http"
	"fmt"
	"strings"
    "io/ioutil"
    "encoding/xml"
    "strconv"
)

type DefaultCommand struct {
	RemoteHost string
}

var proxyInfo ProxyInfo

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
		fmt.Println("Error forwarding request to black pearl: ", err)
		return
	}

	defer response.Body.Close()

    var responseBody []byte
    responseBody, err = ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("Error reading response body: ", err)
        return
    }

    if IsMasterObjectListXml(string(responseBody)) {
        responseBody = generateResponseBody(responseBody, response)
    }

	writeHeaders(responseWriter, response)
	responseWriter.WriteHeader(response.StatusCode)

    responseWriter.Write(responseBody)

	return
}

func writeHeaders(responseWriter http.ResponseWriter, response *http.Response) {
	for key, value := range response.Header {
        responseWriter.Header().Set(key, strings.Join(value, ","))
	}
}

func generateResponseBody(originalResponseBody []byte, response *http.Response) []byte {
    result := originalResponseBody

    masterObjectList := NewMasterObjectList(originalResponseBody)

    if masterObjectList != nil {
        masterObjectList.SetNodesEndpoint(proxyInfo.IpAddress())
        masterObjectList.SetNodesHttpPort(proxyInfo.HttpPort())

        var newResponseBody []byte

        newResponseBody, err := xml.Marshal(masterObjectList)

        if err != nil {
            fmt.Println("Error marshaling new master object list: ", err)
        } else {
            response.Header.Set("Content-Length", strconv.Itoa(len(newResponseBody)))

            result = newResponseBody
        }
    }

    return result
}