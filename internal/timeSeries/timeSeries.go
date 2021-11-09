package timeSeries

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

type TimeSeries struct {
	ID         string `json:"ID"`
	Admin2     string `json:"Admin2"`
	Address1   string `json:"Province/State"`
	Address2   string `json:"Country/Region"`
	Population int    `json:"Population"`

	Confirmed map[time.Time]int `json:"Confirmed"`
	Death     map[time.Time]int `json:"Death"`
	Recovered map[time.Time]int `json:"Recovered"`
}

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", List)
	r.Post("/", Create)
	r.Put("/", Update)

	// TODO: Add querying

	return r
}

func List(w http.ResponseWriter, r *http.Request) {
	// TODO: Get all existing data
	body := "Executing List"
	if _, err := w.Write([]byte(body)); err != nil {
		log.Fatal(err)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	reader := csv.NewReader(r.Body)

	for {
		result, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Parse here

		for i := range result {
			// do something with result[i] <- value at ith column in 'result'
			fmt.Printf("	%v'n", result[i])
		}

		/*
			if file, err := os.Open(FilePath); err != nil {
				log.Fatal("File read error", err)
				return
			}
			csvReader := csv.NewReader(f)
			if records, err := csvReader.ReadAll(); err != nil {
				log.Fatal(filePath + ": Parse file error", err)
			}
		*/

		fmt.Println(result)
	}

	if _, err := w.Write([]byte("Successfully Uploaded Data")); err != nil {
		log.Fatal(err)
	}
}

func Update(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)
	body := "" + id
	fmt.Println(body)
}
