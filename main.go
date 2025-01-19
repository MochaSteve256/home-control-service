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

func forward(w http.ResponseWriter, r *http.Request) {
	log.Println("Forwarding request")

	target_url := "http://stevepi:5000/"
	// get route and append to target url
	route := mux.CurrentRoute(r)
	path, _ := route.GetPathTemplate()
	target_url += path

	log.Println(target_url)

	http.Redirect(w, r, target_url, http.StatusTemporaryRedirect)

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/wake", wake_pc).Methods("POST")

	//forwarded routes
	router.HandleFunc("/psu", forward).Methods("POST", "GET")
	router.HandleFunc("/led", forward).Methods("POST", "GET")
	router.HandleFunc("/alarm", forward).Methods("POST", "GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
