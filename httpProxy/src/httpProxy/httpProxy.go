package main

import (
	"os"
	"log"
	"net/http"
	"net"
	"command"
	"fmt"
)

type ds3HttpClientConnectionInfo struct {
	remoteHost string
	proxyPort  string
	adminPort  string
}

var httpClientClientConnectionInfo ds3HttpClientConnectionInfo
var httpListener net.Listener
var defaultCommand command.DefaultCommand

// TODO set the remote host and other configuration stuff from admin request

func main() {
	getDs3ConnectionInfoFromEnvironmentVars()

	defaultCommand = command.DefaultCommand{RemoteHost:httpClientClientConnectionInfo.remoteHost}

	go listenOnAdminPort()

	listenOnProxyPort(func () {
		fmt.Printf("Listening on proxy port %s and admin port %s\n",
			httpClientClientConnectionInfo.proxyPort,
			httpClientClientConnectionInfo.adminPort)
	})
}

func getDs3ConnectionInfoFromEnvironmentVars() {
	httpClientClientConnectionInfo.remoteHost = getEnvironmentVar("DS3_TARGET_SYSTEM_DNS_NAME", "sm2u-11.eng.sldomain.com")
	httpClientClientConnectionInfo.proxyPort = getEnvironmentVar("HTTP_PROXY_PORT", ":9080")
	httpClientClientConnectionInfo.adminPort = getEnvironmentVar("HTTP_PROXY_ADMIN_PORT", ":9090")
}

func getEnvironmentVar(envVarToLookFor string, defaultIfVarNotSet string) string {
	result := os.Getenv(envVarToLookFor)

	if len(result) == 0 {
		return defaultIfVarNotSet
	}

	return result
}

func listenOnAdminPort() {
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/close", finit)

	err := http.ListenAndServe(httpClientClientConnectionInfo.adminPort, adminMux)

	if err != nil {
		fmt.Printf("Error listening on admin port: %s\n", httpClientClientConnectionInfo.adminPort)
		panic(err)
	}
}

func listenOnProxyPort(onSuccess func()) {
	var err error

	httpListener, err = net.Listen("tcp", httpClientClientConnectionInfo.proxyPort)

	if err != nil {
		fmt.Printf("Error listening on port: %s\n", httpClientClientConnectionInfo.proxyPort)
		panic(err)
	}

	onSuccess()

	httpHandler := http.NewServeMux()
	httpHandler.HandleFunc("/", proxyHandler)
	http.Serve(httpListener, httpHandler)
}

var commandToRun command.Command

func proxyHandler(responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	commandToRun = CommandForUrlPrefix(request.RequestURI, request.Method)

	if commandToRun != nil {
		err, handled := commandToRun.Execute(responseWriter, request)

		if(err != nil) {
			log.Println("Error: ", err, " processing http request: ", request);
		}

		if handled {
			return
		}
	}

	commandToRun = &defaultCommand

	commandToRun.Execute(responseWriter, request)
}

func finit (responseWriter http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	if request.Method == "PUT" {
		responseWriter.Write([]byte("Closing...\n"))
		fmt.Printf("Later dude\n")
		httpListener.Close()
	}
}

