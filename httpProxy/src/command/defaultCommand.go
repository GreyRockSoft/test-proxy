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
        masterObjectList := NewMasterObjectList(responseBody)

        if masterObjectList != nil {
            masterObjectList.SetNodesEndpoint(proxyInfo.IpAddress())
            masterObjectList.SetNodesHttpPort(proxyInfo.HttpPort())

            for _, node := range masterObjectList.Nodes.Node {
                fmt.Printf("%+v", node)
            }

            var newResponseBody []byte

            newResponseBody, err = xml.Marshal(masterObjectList)

            if err != nil {
                fmt.Println("Error marshaling new master object list: ", err)
            } else {
                const contentLengthKeyName = "Content-Length"
                response.Header.Set(contentLengthKeyName, strconv.Itoa(len(newResponseBody)))

                fmt.Println(contentLengthKeyName, "; ", response.Header.Get(contentLengthKeyName), " Original response body size: ", len(responseBody),
                    " New response body  size: ", len(newResponseBody));

                responseBody = newResponseBody



                fmt.Println("--> ", contentLengthKeyName, "; ", response.Header.Get(contentLengthKeyName), " Original response body size: ", len(responseBody),
                    " New response body  size: ", len(newResponseBody));
            }
        }
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

func writeResponseBody(responseWriter http.ResponseWriter, response *http.Response) {
    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("Error reading response body: ", err)
        return
    }

    shouldWriteOriginalResponseBody := true

    responseBodyString := string(responseBody)

    if IsMasterObjectListXml(responseBodyString) {
        masterObjectList := NewMasterObjectList(responseBody)

        if masterObjectList != nil {
            masterObjectList.SetNodesEndpoint(proxyInfo.IpAddress())
            masterObjectList.SetNodesHttpPort(proxyInfo.HttpPort())

            // fmt.Printf("%+v", masterObjectList)

            var newResponseBody []byte

            newResponseBody, err = xml.Marshal(masterObjectList)

            if err == nil {
                fmt.Println("newResponseBody size: ", len(newResponseBody))
                responseWriter.Write(newResponseBody)
                shouldWriteOriginalResponseBody = false
            } else {
                fmt.Println("Error marshaling new master object list: ", err)
            }
        }
    }

    if shouldWriteOriginalResponseBody {
        responseWriter.Write(responseBody)
    }
}
