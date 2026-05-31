package pkg

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func JSONResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := Response{
		Status:  statusCode,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

func ErrorResponse(w http.ResponseWriter, statusCode int, message string, err interface{}) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	response := Response{
		Status:  statusCode,
		Message: message,
		Error:   err,
	}
	json.NewEncoder(w).Encode(response)
}

func ParseToDate(dateStr string) (time.Time, error) {
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, dateStr)
	if err != nil {
		return time.Time{}, err
	}

	return parsedDate, nil
}

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
