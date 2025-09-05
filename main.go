package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"syscall"

	"github.com/gorilla/mux"
)

// CORS middleware adds the necessary headers to each response.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins or adjust as needed
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow the necessary methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Include 'token' in allowed headers along with others
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, token")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

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
			return
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
	targetURL := "http://stevepi:5000"
	// If the path is /volume, change the targetURL
	if r.URL.Path == "/volume" || r.URL.Path == "/music" || r.URL.Path == "/lock" {
		targetURL = "http://adrians-pc:8080"
	}
	fullURL, err := url.Parse(targetURL + r.URL.Path)
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

	// Copy headers from the original request
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, syscall.EHOSTUNREACH) || errors.Is(err, syscall.ENETUNREACH) {
			http.Error(w, "Server unreachable: no route to host", http.StatusServiceUnavailable)
			log.Println("Error forwarding request (no route to host):", err)
			return
		}
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		log.Println("Error forwarding request:", err)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("Error copying response body:", err)
	}
}

func main() {
	router := mux.NewRouter()
	// Use the CORS middleware on all routes
	router.Use(corsMiddleware)

	// Waking done by MochaPi
	router.HandleFunc("/wake", wake_pc).Methods("POST", "OPTIONS")

	// PowerHub routes
	router.HandleFunc("/psu", forward)
	router.HandleFunc("/led", forward)
	router.HandleFunc("/alarm", forward)
	router.HandleFunc("/alarm/{id}", forward) // <-- handles /alarm/<id>
	router.HandleFunc("/alarm/actions", forward)
	router.HandleFunc("/dismiss", forward)
	router.HandleFunc("/dim", forward)

	// PC routes
	router.HandleFunc("/volume", forward)
	router.HandleFunc("/music", forward)
	router.HandleFunc("/lock", forward)

	log.Fatal(http.ListenAndServe(":8080", router))
}
