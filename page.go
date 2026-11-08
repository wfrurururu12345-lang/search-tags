package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"
)

type PageData struct {
	Query    string
	Tags     []Tag
	Searched bool
}

func registerPage(mux *http.ServeMux, db *sql.DB) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		raw := r.URL.Query().Get("q")
		query := normalize(raw)

		data := PageData{
			Query:    raw,
			Tags:     []Tag{},
			Searched: query != "",
		}

		if data.Searched {
			ctx, cancel := context.WithTimeout(
				r.Context(), 3*time.Second,
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
				http.Error(w, "Database error", 500)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var tag Tag
				if err := rows.Scan(&tag.Tag, &tag.Users); err != nil {
					log.Printf("Scan error: %v", err)
					http.Error(w, "Database error", 500)
					return
				}
				data.Tags = append(data.Tags, tag)
			}

			if err := rows.Err(); err != nil {
				log.Printf("Rows error: %v", err)
				http.Error(w, "Database error", 500)
				return
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Template error: %v", err)
		}
	})
}
