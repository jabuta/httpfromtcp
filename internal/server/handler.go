package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
)

type HandlerError struct {
	Code    response.StatusCode
	Message string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func (he *HandlerError) respondWithError(w io.Writer) {
	response.WriteStatusLine(w, he.Code)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}
