package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type AdminPageData struct {
	IsAuthenticated bool
	Error           string
	Success         string
	Stats           *StatsData
	Status          *StatusData
}

type StatsData struct {
	WordsTotal    int `json:"words_total"`
	WordsUnique   int `json:"words_unique"`
	ComicsFetched int `json:"comics_fetched"`
	ComicsTotal   int `json:"comics_total"`
}

type StatusData struct {
	Status string `json:"status"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func NewAdminHandler(log *slog.Logger, apiAddress string, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := AdminPageData{
			IsAuthenticated: false,
		}

		tokenCookie, err := r.Cookie("auth_token")
		if err == nil && tokenCookie.Value != "" {
			if verifyToken(apiAddress, tokenCookie.Value) {
				data.IsAuthenticated = true
			}
		}

		// Handle login
		if r.Method == http.MethodPost && r.URL.Path == "/admin/login" {
			var loginReq LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
				data.Error = "Invalid login data"
			} else {
				token, err := login(apiAddress, loginReq.Name, loginReq.Password)
				if err != nil {
					data.Error = "Invalid credentials"
				} else {
					http.SetCookie(w, &http.Cookie{
						Name:     "auth_token",
						Value:    token,
						Path:     "/",
						HttpOnly: true,
						MaxAge:   120,
					})
					data.IsAuthenticated = true
					data.Success = "Login successful"
				}
			}
		}

		// Handle logout
		if r.Method == http.MethodPost && r.URL.Path == "/admin/logout" {
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})
			data.IsAuthenticated = false
			data.Success = "Logged out successfully"
		}

		// Handle actions (update, drop, stats, status)
		if data.IsAuthenticated && tokenCookie != nil {
			if r.Method == http.MethodPost {
				action := r.URL.Query().Get("action")
				switch action {
				case "update":
					if err := callAPI(apiAddress, "POST", "/api/db/update", tokenCookie.Value, nil); err != nil {
						data.Error = fmt.Sprintf("Update failed: %v", err)
					} else {
						data.Success = "Database update started"
					}
				case "drop":
					if err := callAPI(apiAddress, "DELETE", "/api/db", tokenCookie.Value, nil); err != nil {
						data.Error = fmt.Sprintf("Drop failed: %v", err)
					} else {
						data.Success = "Database dropped successfully"
					}
				}
			}

			// Load stats and status
			if stats, err := getStats(apiAddress); err == nil {
				data.Stats = stats
			}
			if status, err := getStatus(apiAddress); err == nil {
				data.Status = status
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "admin.html", data); err != nil {
			log.Error("failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func login(apiAddress, name, password string) (string, error) {
	loginReq := LoginRequest{Name: name, Password: password}
	jsonData, _ := json.Marshal(loginReq)

	resp, err := http.Post(apiAddress+"/api/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed")
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(token)), nil
}

func verifyToken(apiAddress, token string) bool {
	req, _ := http.NewRequest("GET", apiAddress+"/api/db/status", nil)
	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func callAPI(apiAddress, method, path, token string, body []byte) error {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, apiAddress+path, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, apiAddress+path, nil)
	}
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

func getStats(apiAddress string) (*StatsData, error) {
	resp, err := http.Get(apiAddress + "/api/db/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var stats StatsData
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func getStatus(apiAddress string) (*StatusData, error) {
	resp, err := http.Get(apiAddress + "/api/db/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var status StatusData
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}
