package api

import (
	"net/http"
)

// StartStreamHandler is a placeholder function for the /stream/start endpoint
func StartStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Start stream placeholder"))
}

// SendDataHandler is a placeholder function for the /stream/{stream_id}/send endpoint
func SendDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Send data placeholder"))
}

// GetResultsHandler is a placeholder function for the /stream/{stream_id}/results endpoint
func GetResultsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get results placeholder"))
}
