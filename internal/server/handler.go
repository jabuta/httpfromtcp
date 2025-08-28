package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
)

// type HandlerError struct {
// 	Code    response.StatusCode
// 	Message string
// }

type Handler func(w *response.Writer, req *request.Request)

// func (he *HandlerError) respondWithError(w response.Writer) {
// 	w.WriteStatusLine(he.Code)
// 	messageBytes := []byte(he.Message)
// 	headers := response.GetDefaultHeaders(len(messageBytes))
// 	w.WriteHeaders(headers)
// 	w.WriteBody(messageBytes)
// }
