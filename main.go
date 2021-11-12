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

// @title CSC301 Assignment2 Group69
// @version 1.0
// @description An API for csc301 a2
// @termsOfService http://swagger.io/terms/

// @contact.name Supanat Wangsutthitham
// @contact.url https://github.com/SupaJuke
// @contact.email soup.wangsutthitham@mail.utoronto.ca

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host http://34.130.134.25:8080/
// @BasePath /api/v1
// @accept text/csv
// @produce json text/csv
// @schemes http

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

	r.Mount("/api/v1/time_series", timeSeries.Routes())
	r.Mount("/api/v1/daily_reports", dailyReports.Routes())

	log.Printf("Listening for requests on http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
