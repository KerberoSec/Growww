package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("default-development-jwt-secret-key-growww-32b")
	}
	demoHost := os.Getenv("DEMO_HOST")
	realHost := os.Getenv("REAL_HOST")

	gw := NewAPIGateway(jwtSecret, demoHost, realHost)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("[Growww API Gateway] Listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, gw); err != nil {
		fmt.Fprintf(os.Stderr, "Gateway exited with error: %v\n", err)
	}
}
