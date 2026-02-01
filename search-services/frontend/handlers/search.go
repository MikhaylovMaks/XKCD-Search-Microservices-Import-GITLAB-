package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

type Comics struct {
	ID    int    `json:"id"`
	URL   string `json:"url"`
	Score int    `json:"score"`
}

type SearchResponse struct {
	Comics []Comics `json:"comics"`
	Total  int      `json:"total"`
}

type SearchPageData struct {
	Query      string
	Comics     []Comics
	Total      int
	Error      string
	HasResults bool
}

func NewSearchHandler(log *slog.Logger, apiAddress string, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("phrase")
		limitStr := r.URL.Query().Get("limit")
		limit := 10
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit < 1 {
				limit = 10
			}
		}

		data := SearchPageData{
			Query:      query,
			Comics:     []Comics{},
			HasResults: false,
		}

		if query != "" {
			// Make request to API
			apiURL := fmt.Sprintf("%s/api/search?phrase=%s&limit=%d", apiAddress, url.QueryEscape(query), limit)
			resp, err := http.Get(apiURL)
			if err != nil {
				log.Error("failed to call API", "error", err)
				data.Error = "Failed to connect to search service"
			} else {
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						log.Error("failed to read response", "error", err)
						data.Error = "Failed to read search results"
					} else {
						var searchResp SearchResponse
						if err := json.Unmarshal(body, &searchResp); err != nil {
							log.Error("failed to parse response", "error", err)
							data.Error = "Failed to parse search results"
						} else {
							data.Comics = searchResp.Comics
							data.Total = searchResp.Total
							data.HasResults = len(searchResp.Comics) > 0
						}
					}
				} else if resp.StatusCode == http.StatusNotFound {
					data.Error = "No comics found"
				} else {
					body, _ := io.ReadAll(resp.Body)
					log.Error("API returned error", "status", resp.StatusCode, "body", string(body))
					data.Error = "Search service error"
				}
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "search.html", data); err != nil {
			log.Error("failed to render template", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

