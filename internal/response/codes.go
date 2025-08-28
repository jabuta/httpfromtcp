package response

type StatusCode int

const (
	HttpOK          StatusCode = 200
	HttpNotFoud     StatusCode = 400
	HttpServerError StatusCode = 500
)
