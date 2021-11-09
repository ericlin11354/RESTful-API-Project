package main

import (
	// Built-ins

	"fmt"
	"log"
	"net/http"
	"os"

	// External imports
	"github.com/go-chi/chi"

	// Internal imports
	"gitlab.com/csc301-assignments/a2/internal/dailyReports"
	"gitlab.com/csc301-assignments/a2/internal/timeSeries"
)

func main() {
	fmt.Println("Hello World!")

	// Get port
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	// Initizalize Router
	router := chi.NewRouter()

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte("Hello World!")); err != nil {
			log.Fatal(err)
		}
	})

	router.Mount("/time_series", timeSeries.Routes())
	router.Mount("/daily_reports", dailyReports.Routes())

	log.Printf("Listening for requests on http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, router))

}
