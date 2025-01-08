package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/wake", wake_pc).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
