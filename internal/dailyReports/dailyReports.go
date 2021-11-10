package dailyReports

import (
	// Built-ins
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	// External imports
	"github.com/go-chi/chi"

	// Internal imports
	db "gitlab.com/csc301-assignments/a2/internal/db"
	"gitlab.com/csc301-assignments/a2/internal/utils"
)

type DailyReports struct {
	ID        string    `json:"id"`
	Admin2    string    `json:"Admin2"`
	Address1  string    `json:"Province/State"`
	Address2  string    `json:"Country/Region"`
	Date      time.Time `json:"Date"`
	Confirmed int       `json:"Confirmed"`
	Death     int       `json:"Death"`
	Recovered int       `json:"Recovered"`
	Active    int       `json:"Active"`
}

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", List)
	r.Post("/", Create)

	return r
}

func nullStringHandler(dr *DailyReports, values map[string]*sql.NullString) {
	dr.ID = values["id"].String
	if values["admin2"].Valid {
		dr.Admin2 = values["admin2"].String
	}
	if values["address1"].Valid {
		dr.Address1 = values["address1"].String
	}
	if values["address2"].Valid {
		dr.Address2 = values["address2"].String
	}
}

func nullIntHandler(dr *DailyReports, values map[string]*sql.NullInt64) {
	if values["Confirmed"].Valid {
		dr.Confirmed = int(values["Confirmed"].Int64)
	}
	if values["Death"].Valid {
		dr.Death = int(values["Death"].Int64)
	}
	if values["Recovered"].Valid {
		dr.Recovered = int(values["Recovered"].Int64)
	}
	if values["Active"].Valid {
		dr.Active = int(values["Active"].Int64)
	}
}

func List(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	query := `
		SELECT ID, Date, Admin2, Address1, Address2,
		Confirmed, Death, Recovered, Active
		FROM DailyReports
	`

	i := 0
	for param, value := range params {
		param = strings.ToLower(param)

		var valid bool
		if param, valid = utils.ParamValidate(param); !valid {
			w.WriteHeader(400)
			if _, err := w.Write([]byte("Error 400: Invalid Input")); err != nil {
				log.Fatal(err)
			}
			return
		}

		// Time interval
		op := "="
		if param == "from" {
			param = "date"
			op = ">="
		}
		if param == "to" {
			param = "date"
			op = "<="
		}

		value := strings.Split(value[0], ",")

		for j, v := range value {
			// Format string for SQL
			stringParams := map[string]string{
				"admin2":   fmt.Sprintf(`'%s'`, v),
				"address1": fmt.Sprintf(`'%s'`, v),
				"address2": fmt.Sprintf(`'%s'`, v),
				"date":     fmt.Sprintf(`'%s'`, v),
				"from":     fmt.Sprintf(`'%s'`, v),
				"to":       fmt.Sprintf(`'%s'`, v),
			}
			_, ok := stringParams[param]
			if ok {
				value[j] = stringParams[param]
			}

			// Format first param and after
			if i == 0 {
				query += "WHERE " + param + op + value[j]
				i++
			} else {
				if j != 0 {
					query += " OR " + param + op + value[j]
				} else {
					query += " AND " + param + op + value[j]
				}

			}
		}
	}
	stmt, err := db.Db.Prepare(query)

	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	row, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()

	drArr := []DailyReports{}
	for row.Next() {
		dr := DailyReports{}

		// Handling null values
		nullStrings := map[string]*sql.NullString{}
		nullStrings["id"] = &sql.NullString{}
		nullStrings["admin2"] = &sql.NullString{}
		nullStrings["address1"] = &sql.NullString{}
		nullStrings["address2"] = &sql.NullString{}
		nullInts := map[string]*sql.NullInt64{}
		nullInts["Confirmed"] = &sql.NullInt64{}
		nullInts["Death"] = &sql.NullInt64{}
		nullInts["Recovered"] = &sql.NullInt64{}
		nullInts["Active"] = &sql.NullInt64{}
		err := row.Scan(nullStrings["id"], &dr.Date, nullStrings["admin2"],
			nullStrings["address1"], nullStrings["address2"],
			nullInts["Confirmed"], nullInts["Death"],
			nullInts["Recovered"], nullInts["Active"],
		)
		if err != nil {
			log.Fatal(err)
		}
		nullStringHandler(&dr, nullStrings)
		nullIntHandler(&dr, nullInts)

		drArr = append(drArr, dr)
	}
	if r.Header.Get("Accept") == "text/csv" {
		w.Header().Set("Content-Type", "text/csv")

		b := new(bytes.Buffer)
		writer := csv.NewWriter(b)
		csvArr := [][]string{
			{"ID", "Date", "Admin2", "Province/State", "Country/Region",
				"Confirmed", "Death", "Recovered", "Active"},
		}

		// Filling in respond in csv format
		for _, dr := range drArr {
			row := []string{
				dr.ID, dr.Date.Format("2006/01/02"), dr.Admin2, dr.Address1, dr.Address2,
				strconv.Itoa(dr.Confirmed),
				strconv.Itoa(dr.Death),
				strconv.Itoa(dr.Recovered),
				strconv.Itoa(dr.Active),
			}
			csvArr = append(csvArr, row)
		}
		if err := writer.WriteAll(csvArr); err != nil {
			log.Fatal(err)
		}

		if _, err := w.Write(b.Bytes()); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := json.NewEncoder(w).Encode(drArr); err != nil {
			log.Fatal(err)
		}
	}

	w.WriteHeader(200)
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
