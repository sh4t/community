package errs
// Errors, error support from json-api stuff

import (
	"encoding/json"
	"net/http"
)

type Errors struct {
	Errors []*Error `json:"errors"`
}

type Error struct {
	Id     string `json:"id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}

var (
	ErrBadRequest           = &Err{"bad_request", 400, "Bad request", "Request body is not well-formed. It must be JSON."}
	ErrUnauthorized         = &Err{"unauthorized", 401, "Unauthorized", "Access token is invalid."}
	ErrNotFound             = &Err{"not_found", 404, "Not found", "Route not found."}
	ErrNotAcceptable        = &Err{"not_acceptable", 406, "Not acceptable", "Accept HTTP header must be \"application/vnd.api+json\"."}
	ErrUnsupportedMediaType = &Err{"unsupported_media_type", 415, "Unsupported Media Type", "Content-Type header must be \"application/vnd.api+json\"."}
	ErrInternalServer       = &Err{"internal_server_error", 500, "Internal Server Error", "Something went wrong."}
)