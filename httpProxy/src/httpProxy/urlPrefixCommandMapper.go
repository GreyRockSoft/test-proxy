package main

import (
	"command"
	"strings"
	"net/http"
)

type urlPrefixCommandEntry struct {
	urlPrefix string
	httpVerb string
	commandToRun command.Command
}

var urlPrefixCommandTable []urlPrefixCommandEntry

var failFirstCommand command.FailFirstAttemptCommand

func init() {
	urlPrefixCommandTable = append(urlPrefixCommandTable,
		urlPrefixCommandEntry{"/Put_Job_Management_Test/lesmis-copies.txt", http.MethodPut, &failFirstCommand})
}

func CommandForUrlPrefix(urlPrefix string, httpVerb string) command.Command  {
	for _, tableEntry := range urlPrefixCommandTable {
		if strings.HasPrefix(urlPrefix, tableEntry.urlPrefix) && httpVerb == tableEntry.httpVerb {
			return tableEntry.commandToRun
		}
	}

	return nil
}

