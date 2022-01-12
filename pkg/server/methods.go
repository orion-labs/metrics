package metrics

import "net/http"

type HttpMethod int

const (
	ANY HttpMethod = iota
	OPTIONS
	GET
	PUT
	POST
	PATCH
	DELETE
	CONNECT
	TRACE
)

func (m HttpMethod) String() string {
	switch m {
	case ANY:
		return "any"
	case OPTIONS:
		return http.MethodOptions
	case GET:
		return http.MethodGet
	case PUT:
		return http.MethodPut
	case POST:
		return http.MethodPost
	case PATCH:
		return http.MethodPatch
	case DELETE:
		return http.MethodDelete
	case CONNECT:
		return http.MethodConnect
	case TRACE:
		return http.MethodTrace
	}
	return "unknown"
}
