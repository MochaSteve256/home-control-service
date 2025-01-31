package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/gorilla/mux"
)

func verify_token(token string) bool {
	if token == "br4d9c2ayqrk7iswse7v8t2x" {
		log.Println("Authenticated")
		return true
	}
	return false
}

func wake_pc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		var requestBody map[string]interface{}
		if err := json.Unmarshal(body, &requestBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, ok := requestBody["token"].(string)
		if !ok {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}

		if !verify_token(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("Waking PC")

		cmd := exec.Command("sudo", "wakeonlan", "a8:a1:59:a5:a1:9b")
		log.Println(cmd.Run())
	}
}

func forward(w http.ResponseWriter, r *http.Request) {
	log.Println("Forwarding request")

	// Define the base URL
	target_url := "http://stevepi:5000"
	fullURL, err := url.Parse(target_url + r.URL.Path)
	log.Println(fullURL)
	if err != nil {
		http.Error(w, "Failed to parse URL", http.StatusInternalServerError)
		log.Println("Error parsing URL:", err)
		return
	}

	// Create the new request to forward
	req, err := http.NewRequest(r.Method, fullURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		log.Println("Error creating new request:", err)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		log.Println("Error forwarding request:", err)
		return
	}
	defer resp.Body.Close()

	// Copy response headers and status
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("Error copying response body:", err)
	}
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/wake", wake_pc).Methods("POST")

	//forwarded routes
	router.HandleFunc("/psu", forward)
	router.HandleFunc("/led", forward)
	router.HandleFunc("/alarm", forward)
	router.HandleFunc("/dismiss", forward)

	log.Fatal(http.ListenAndServe(":8080", router))
}
