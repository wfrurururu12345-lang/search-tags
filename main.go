package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type Tag struct {
	Tag   string `json:"tag"`
	Users uint64 `json:"users"`
}

var spaces = regexp.MustCompile(` +`)

func normalize(s string) string {
	s = strings.Trim(s, " ")
	s = strings.ToLower(s)
	return spaces.ReplaceAllString(s, " ")
}

func tagsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := normalize(r.URL.Query().Get("q"))

		if query == "" {
			http.Error(w, "Empty query", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(
			r.Context(),
			3*time.Second,
		)
		defer cancel()

		rows, err := db.QueryContext(ctx, `
            SELECT tag, users
            FROM search_tags.tags
            WHERE query = {q:String}
            ORDER BY rank
            LIMIT 10
        `, sql.Named("q", query))

		if err != nil {
			log.Printf("Database error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		tags := []Tag{}

		for rows.Next() {
			var tag Tag

			if err := rows.Scan(&tag.Tag, &tag.Users); err != nil {
				log.Printf("Scan error: %v", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}

			tags = append(tags, tag)
		}

		if err := rows.Err(); err != nil {
			log.Printf("Rows error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(tags); err != nil {
			log.Printf("JSON error: %v", err)
		}
	}
}

func main() {
	addr := os.Getenv("CLICKHOUSE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:9000"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "search_tags",
			Username: "default",
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
	})
	defer db.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ClickHouse connection failed: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/tags", tagsHandler(db))
	registerPage(mux, db)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Server started on port %s", port)
	log.Fatal(server.ListenAndServe())
}
