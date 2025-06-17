package handlers

import (
	"Go_Playground/HttpProxy/internal/model"
	"Go_Playground/HttpProxy/internal/state"
	"encoding/json"
	"net/http"
)

func RequestsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	output := model.RequestsResponse{
		TotalRequests: state.LoadTotalRequests(),
		TotalErrors:   state.LoadErrorCount(),
		Requests:      state.GetAllRequests(),
	}

	json.NewEncoder(w).Encode(output)
}
