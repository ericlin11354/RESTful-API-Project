package main

import (
	// Built-ins

	"log"
	"net/http"
	"os"

	// External imports
	"github.com/go-chi/chi"

	// Internal imports
	"gitlab.com/csc301-assignments/a2/internal/dailyReports"
	db "gitlab.com/csc301-assignments/a2/internal/db"
	"gitlab.com/csc301-assignments/a2/internal/timeSeries"
)

func main() {
	// Get port
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	db.InitDb()

	// Initizalize Router
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte("CSC301 Assignment 2 Group 69")); err != nil {
			log.Fatal(err)
		}
	})

	r.Mount("/api/time_series", timeSeries.Routes())
	r.Mount("/api/daily_reports", dailyReports.Routes())

	log.Printf("Listening for requests on http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
