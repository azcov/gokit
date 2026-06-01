package response

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Success bool  `json:"success"`
	Data    any   `json:"data,omitempty"`
	Error   *Err  `json:"error,omitempty"`
	Meta    *Meta `json:"meta,omitempty"`
}

type Err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Success: true, Data: data})
}

func WithMeta(w http.ResponseWriter, status int, data any, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Success: true, Data: data, Meta: meta})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Success: false, Error: &Err{Code: code, Message: message}})
}

func OK(w http.ResponseWriter, data any)      { JSON(w, http.StatusOK, data) }
func Created(w http.ResponseWriter, data any) { JSON(w, http.StatusCreated, data) }
func NoContent(w http.ResponseWriter)         { w.WriteHeader(http.StatusNoContent) }
func BadRequest(w http.ResponseWriter, code, msg string) {
	Error(w, http.StatusBadRequest, code, msg)
}
func Unauthorized(w http.ResponseWriter, code, msg string) {
	Error(w, http.StatusUnauthorized, code, msg)
}
func Forbidden(w http.ResponseWriter, code, msg string) {
	Error(w, http.StatusForbidden, code, msg)
}
func NotFound(w http.ResponseWriter, code, msg string) {
	Error(w, http.StatusNotFound, code, msg)
}
func InternalServerError(w http.ResponseWriter, code, msg string) {
	Error(w, http.StatusInternalServerError, code, msg)
}
