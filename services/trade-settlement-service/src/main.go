package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	coordinator := NewDvPSettlementCoordinator()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"HEALTHY","service":"trade-settlement-service"}`))
	})

	fmt.Printf("[Growww Trade Settlement Service] Listening on :%s\n", port)
	_ = coordinator
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Settlement service exited with error: %v\n", err)
	}
}
