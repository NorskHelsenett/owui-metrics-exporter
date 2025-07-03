package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type UsageRequest struct {
	ModelIDs []string `json:"model_ids"`
	UserIDs  []string `json:"user_ids"`
}

type UsageResponse struct {
	UserIDs []string `json:"user_ids"`
}

type User struct {
	ID string `json:"id"`
}

type UserList struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

func fetchOWUIStats(baseURL, token string) (loggedIn, total int, err error) {
	req, _ := http.NewRequest("GET", baseURL+"/api/v1/users/?page=0&order_by=created_at&direction=asc", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var users UserList
	if err = json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return
	}
	total = users.Total

	usageReq, _ := http.NewRequest("GET", baseURL+"/api/usage", nil)
	usageReq.Header.Set("Authorization", "Bearer "+token)
	usageReq.Header.Set("Content-Type", "application/json")

	usageResp, err := client.Do(usageReq)
	if err != nil {
		return
	}
	defer usageResp.Body.Close()

	var usage UsageResponse
	if err = json.NewDecoder(usageResp.Body).Decode(&usage); err != nil {
		return
	}

	loggedIn = len(usage.UserIDs)
	return
}

func metricsHandler(baseURL, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedIn, total, err := fetchOWUIStats(baseURL, token)
		if err != nil {
			http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		fmt.Fprintf(w, "# HELP owui_logged_in_users Number of users currently logged in\n")
		fmt.Fprintf(w, "# TYPE owui_logged_in_users gauge\n")
		fmt.Fprintf(w, "owui_logged_in_users %d\n", loggedIn)

		fmt.Fprintf(w, "# HELP owui_total_users Total number of registered users\n")
		fmt.Fprintf(w, "# TYPE owui_total_users gauge\n")
		fmt.Fprintf(w, "owui_total_users %d\n", total)

		log.Printf("Logged in users: %d, Total users: %d", loggedIn, total)
	}
}

func main() {
	// Load .env file but don't override existing environment variables
	_ = godotenv.Load()

	baseURL := os.Getenv("OWUI_BASE_URL")
	token := os.Getenv("OWUI_JWT")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if baseURL == "" || token == "" {
		log.Fatal("Missing OWUI_BASE_URL or OWUI_JWT in env")
	}

	http.HandleFunc("/metrics", metricsHandler(baseURL, token))

	log.Printf("Exposing OWUI metrics at :%s/metrics", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
