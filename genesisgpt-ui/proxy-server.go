package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type QueryRequest struct {
	Query               string `json:"query"`
	ShowThinkingProcess bool   `json:"showThinkingProcess"`
}

type QueryResponse struct {
	Response string `json:"response"`
}

func main() {
	// GenesisGpt backend URL
	genesisgptURL := os.Getenv("GENESISGPT_URL")
	if genesisgptURL == "" {
		genesisgptURL = "http://localhost:8090"
	}

	// Serve the UI
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Proxy API requests to GenesisGpt
	http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request", http.StatusBadRequest)
			return
		}

		// Forward to GenesisGpt
		resp, err := http.Post(genesisgptURL+"/query", "application/json", bytes.NewBuffer(body))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "GenesisGpt service unavailable: " + err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		// Copy response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("GenesisGPT UI running on http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}