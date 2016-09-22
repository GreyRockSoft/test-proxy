package command

import (
	"net/http"
	"log"
)

type FailFirstAttemptCommand struct {
	numRetries int
}

func (failFirstAttemptCommand *FailFirstAttemptCommand) Execute(responseWriter http.ResponseWriter, request *http.Request) (err error, handled bool) {
	failFirstAttemptCommand.numRetries++

	if failFirstAttemptCommand.numRetries == 1 {
		responseWriter.WriteHeader(http.StatusInternalServerError)
		err = nil; handled = true
		log.Println("Preventing attempt to ", request.Method, ": ", request.RequestURI)
		return
	}

	log.Println("Allowing attempt to to ", request.Method, ": ", request.RequestURI)

	err = nil; handled = false
	return
}

