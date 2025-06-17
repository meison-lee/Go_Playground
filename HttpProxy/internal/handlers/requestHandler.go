package handlers

import (
	"Go_Playground/HttpProxy/internal/state"
	"Go_Playground/HttpProxy/internal/types"
	"encoding/json"
	"net/http"
)

func RequestsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	output := types.RequestsResponse{
		TotalRequests: state.LoadTotalRequests(),
		TotalErrors:   state.LoadErrorCount(),
		Requests:      state.GetAllRequests(),
	}

	json.NewEncoder(w).Encode(output)
}
