package command

import (
    "net/http"
    "log"
    "strconv"
    "fmt"
    "io"
)

type PartialDataFromGetCommand struct {
    RemoteHost string
    numRetries int
    MaxNumRetries int
}

func (partialDataFromGetCommand *PartialDataFromGetCommand) Execute(responseWriter http.ResponseWriter, request *http.Request) (err error, handled bool) {
    handled = true

    request.URL.Host = partialDataFromGetCommand.RemoteHost
    request.RequestURI = ""

    if request.URL.Scheme == "" {
        request.URL.Scheme = "http"
    }

    httpClient := http.Client{}
    var response *http.Response
    response, err = httpClient.Do(request)

    if err != nil {
        log.Println(err)
        return
    }

    defer response.Body.Close()

    contentLengthHeaderId := "Content-Length"

    contentLen, err := strconv.Atoi(response.Header.Get(contentLengthHeaderId))

    if err != nil {
        log.Println(err)
        return
    }

    log.Printf("=========>, Http request content length: %d\n", contentLen)

    var numBytesToTransfer int

    partialDataFromGetCommand.numRetries++

    if partialDataFromGetCommand.numRetries <= partialDataFromGetCommand.MaxNumRetries {
        numBytesToTransfer = contentLen / 4
    } else {
        numBytesToTransfer = contentLen
    }

    log.Printf("=========>, Content length we will return: %d\n", numBytesToTransfer)

    writeHeaders(responseWriter, response)
    responseWriter.Header().Set(contentLengthHeaderId, fmt.Sprintf("%d", contentLen))
    responseWriter.WriteHeader(response.StatusCode)

    readBuffer := make([]byte, 1024 * 1024)
    var totalNumBytesRead int
    var dataToTransfer []byte
    limitedReader := io.LimitedReader{R : response.Body, N : int64(numBytesToTransfer)}

    for totalNumBytesRead < numBytesToTransfer {
        var numBytesRead int
        numBytesRead, err = limitedReader.Read(readBuffer)

        log.Printf("=========>, Num bytes read: %d\n", numBytesRead)

        if err == nil || err == io.EOF {
            if numBytesRead > 0 {
                dataToTransfer = append(dataToTransfer, readBuffer[0:numBytesRead]...)
            }

            if err == io.EOF {
                break
            }
        } else {
            log.Println(err)
            return
        }

        totalNumBytesRead += numBytesRead
    }

    var numBytesWritten int

    numBytesWritten, err = responseWriter.Write(dataToTransfer)

    if err != nil {
        log.Println(err)
        return
    }

    if numBytesWritten != len(dataToTransfer) {
        log.Printf("numBytesWritten: %d, data length: %d\n", numBytesWritten, len(dataToTransfer))
    }

    log.Println("  Response header: ", response.Header)
    log.Println("  Response body: ", response.Body)

    return
}
