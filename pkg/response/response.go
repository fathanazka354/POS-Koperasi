package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, code int, success bool, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	})
}

func Success(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, true, message, data)
}

func Created(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusCreated, true, message, data)
}

func BadRequest(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Message: message})
}

func Unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Message: message})
}

func InternalError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Message: message})
}

func NotFound(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Message: message})
}
