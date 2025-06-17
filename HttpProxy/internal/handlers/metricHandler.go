package handlers

import (
	"Go_Playground/HttpProxy/internal/state"
	"fmt"
	"net/http"
)

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "proxy_requests_total %d\n", state.LoadTotalRequests())
	fmt.Fprintf(w, "proxy_errors_total %d\n", state.LoadErrorCount())

	// Get route counts
	routeCounts := state.GetRouteCountCopy()
	for prefix, count := range routeCounts {
		fmt.Fprintf(w, "proxy_requests_%s %d\n", prefix[1:len(prefix)-1], count)
	}

	// Get status code counts
	statusCounts := state.GetStatusCodeCountCopy()
	for statusCode, count := range statusCounts {
		fmt.Fprintf(w, "proxy_status_code_total{code=\"%d\"} %d\n", statusCode, count)
	}

	// Backend health status
	for prefix, routePool := range state.Routes {
		backends := routePool.GetBackends()
		fmt.Fprintf(w, "\nPrefix:%s\n", prefix)
		for _, backend := range backends {
			fmt.Fprintf(w, "backend_health{prefix=\"%s\",url=\"%s\"} %t\n", prefix, backend.URL, backend.Health)
		}
	}
}
