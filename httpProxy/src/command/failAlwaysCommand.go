package command

import (
    "net/http"
    "log"
)

type FailAlwaysCommand struct {
}

func (failAlwaysCommand *FailAlwaysCommand) Execute(responseWriter http.ResponseWriter, request *http.Request) (err error, handled bool) {
    responseWriter.WriteHeader(http.StatusInternalServerError)
    err = nil; handled = true
    log.Println("Preventing attempt to ", request.Method, ": ", request.RequestURI)
    return
}
