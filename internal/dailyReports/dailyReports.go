package dailyReports

import (
	// Built-ins
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	// External imports
	"github.com/go-chi/chi"
)

type DailyReports struct {
	ID        string    `json:"id"`
	Admin2    string    `json:"Admin2"`
	Address1  string    `json:"Province/State"`
	Address2  string    `json:"Country/Region"`
	Date      time.Time `json:"Date"`
	Confirmed int       `json:"Confirmed"`
	Deaths    int       `json:"Deaths"`
	Recovered int       `json:"Recovered"`
	Active    int       `json:"Active"`
}

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", List)
	r.Post("/", Create)

	return r
}

func DailyReportCtx(next http.Handler) http.Handler {
	type key string
	const ctxKey key = "id"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), ctxKey, chi.URLParam(r, "id"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func List(w http.ResponseWriter, r *http.Request) {
	// TODO: Get all existing data
	body := "Existing data"
	if _, err := w.Write([]byte(body)); err != nil {
		log.Fatal(err)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv")
	reader := csv.NewReader(r.Body)

	body := ""
	for {
		result, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for i := range result {
			body += result[i] + " "
		}
		body += "\n"

		fmt.Println(result)
	}

	if _, err := w.Write([]byte(body)); err != nil {
		log.Fatal(err)
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)

	body := "" + id
	// TODO: Finish function

	if _, err := w.Write([]byte(body)); err != nil {
		log.Fatal(err)
	}

	// Might wanna look into this:
	// if _, err := io.Copy(w, resp.Body); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
}

func Update(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)
	body := "" + id
	fmt.Println(body)
}
