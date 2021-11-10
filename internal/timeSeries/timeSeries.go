package timeSeries

import (
	// Built-ins

	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
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

type TimeSeries struct {
	ID       string `json:"ID"`
	Admin2   string `json:"Admin2"`
	Address1 string `json:"Province/State"`
	Address2 string `json:"Country/Region"`

	Confirmed map[time.Time]int `json:"Confirmed"`
	Death     map[time.Time]int `json:"Death"`
	Recovered map[time.Time]int `json:"Recovered"`
}

type TimeSeriesDate struct {
	Date      time.Time
	Confirmed int
	Death     int
	Recovered int
}

func Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", List)
	r.Post("/", Create)

	return r
}

func nullHandler(ts *TimeSeries, values map[string]*sql.NullString) {
	ts.ID = values["id"].String
	if values["admin2"].Valid {
		ts.Admin2 = values["admin2"].String
	}
	if values["address1"].Valid {
		ts.Address1 = values["address1"].String
	}
	if values["address2"].Valid {
		ts.Address2 = values["address2"].String
	}
}

func List(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	query := `
		SELECT DISTINCT TimeSeries.ID, Admin2, Address1, Address2
		FROM TimeSeries JOIN TimeSeriesDate
		ON TimeSeries.ID = TimeSeriesDate.ID
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
				if param == "date" || param == "from" || param == "to" {
					if _, err := utils.ParseDate(v); err != nil {
						w.WriteHeader(400)
						if _, err := w.Write([]byte("Error 400: Invalid Input")); err != nil {
							log.Fatal(err)
						}
						return
					}
				}
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

	tsArr := []TimeSeries{}
	for row.Next() {
		ts := TimeSeries{}

		// Handling null values
		temp := map[string]*sql.NullString{}
		temp["id"] = &sql.NullString{}
		temp["admin2"] = &sql.NullString{}
		temp["address1"] = &sql.NullString{}
		temp["address2"] = &sql.NullString{}
		err := row.Scan(temp["id"], temp["admin2"], temp["address1"], temp["address2"])
		if err != nil {
			log.Fatal(err)
		}
		nullHandler(&ts, temp)

		// Initializing empty maps (to be filled)
		ts.Confirmed = map[time.Time]int{}
		ts.Death = map[time.Time]int{}
		ts.Recovered = map[time.Time]int{}
		tsArr = append(tsArr, ts)
	}

	// Filling maps
	for _, ts := range tsArr {
		query := fmt.Sprintf(`
			SELECT Date, Confirmed, Death, Recovered FROM TimeSeriesDate
			WHERE ID = %s
		`, ts.ID)
		stmt, err := db.Db.Prepare(query)
		if err != nil {
			log.Fatal(err)
		}

		defer stmt.Close()

		rows, err := stmt.Query()
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			tsd := TimeSeriesDate{}
			err := rows.Scan(&tsd.Date, &tsd.Confirmed, &tsd.Death, &tsd.Recovered)
			if err != nil {
				log.Fatal(err)
			}

			ts.Confirmed[tsd.Date] = tsd.Confirmed
			ts.Death[tsd.Date] = tsd.Death
			ts.Recovered[tsd.Date] = tsd.Recovered
		}
	}

	if r.Header.Get("Accept") == "text/csv" {
		w.Header().Set("Content-Type", "text/csv")

		b := new(bytes.Buffer)
		writer := csv.NewWriter(b)
		csvArr := [][]string{
			{"ID", "Date", "Confirmed", "Death", "Recovered"},
		}

		// Filling in respond in csv format
		for _, ts := range tsArr {
			for date := range ts.Confirmed {
				row := []string{
					ts.ID,
					date.Format("2006/01/02"),
					strconv.Itoa(ts.Confirmed[date]),
					strconv.Itoa(ts.Death[date]),
					strconv.Itoa(ts.Recovered[date]),
				}
				csvArr = append(csvArr, row)
			}
		}
		if err := writer.WriteAll(csvArr); err != nil {
			log.Fatal(err)
		}

		if _, err := w.Write(b.Bytes()); err != nil {
			log.Fatal(err)
		}

	} else {
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(tsArr); err != nil {
			log.Fatal(err)
		}
	}

	w.WriteHeader(200)
}

/* Preconditions:
1. Dates are contiguous.
2. The only column values with '/' are dates.
*/
func Create(w http.ResponseWriter, r *http.Request) {
	filetype := r.Header.Get("FileType") // i.e. Recovered, Confirms, Deaths
	ts := TimeSeries{}
	reader := csv.NewReader(r.Body)

	fmt.Println("running Create()")

	// get header names
	result, err := reader.Read()
	if err != nil {
		log.Fatal(err)
	}

	// Directly access column values
	var Admin2Index int
	var Address1Index int
	var Address2Index int
	var beginDateIndex int

	// get beginDate and endDate
	var beginDate time.Time
	var endDate time.Time
	beginFlag := false // true -> beginDate found; false Otherwise
	endFlag := false
	// Broken
	for i := range result {
		if !beginFlag && len(strings.Split(result[i], "/")) == 3 {
			beginDate, err = utils.ParseDate(result[i]) // parses Date string -> time.Time
			if err != nil {
				log.Fatal(err)
			}
			beginDateIndex = i
			beginFlag = true
		}
		if !endFlag && len(strings.Split(result[len(result)-i-1], "/")) == 3 { // searches backwards in array
			endDate, err = utils.ParseDate(result[len(result)-i-1])
			if err != nil {
				log.Fatal(err)
			}
			endFlag = true
		}
	}
	for i := range result {
		switch result[i] {
		case "Admin2":
			Admin2Index = i
		case "Province/State":
			Address1Index = i
		case "Country/Region":
			Address2Index = i
		}
	}
	for {
		result, err = reader.Read()
		if err != nil {
			log.Fatal(err)
		}
		ts.Admin2 = result[Admin2Index]
		ts.Address1 = result[Address1Index]
		ts.Address2 = result[Address2Index]
		// Assuming that we have successfully parsed to a TimeSeries struct
		stmt, err := db.Db.Prepare("INSERT INTO TimeSeries(Admin2, Address1, Address2) VALUES(?,?,?)")
		if err != nil {
			log.Fatal(err)
		}
		res, err := stmt.Exec(ts.Admin2, ts.Address1, ts.Address2)
		if err != nil {
			log.Fatal(err)
		}

		id, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		query := fmt.Sprintf("INSERT INTO TimeSeries%s VALUES(?,?,?)", filetype)
		stmt, err = db.Db.Prepare(query)
		if err != nil {
			log.Fatal(err)
		}

		ts.Confirmed = make(map[time.Time]int)
		ts.Death = make(map[time.Time]int)
		ts.Recovered = make(map[time.Time]int)

		dateIndex := beginDateIndex
		for date := beginDate; date != endDate.Add(time.Hour*24); date = date.AddDate(0, 0, 1) { // iterate between beginDate and endDate inclusive, incrementing by 1 Day
			val, err := strconv.Atoi(result[dateIndex])
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Printf("hewwo %v\n", date.String())
			switch filetype {
			case "Confirmed":
				ts.Confirmed[date] = val
				_, err = stmt.Exec(id, date, ts.Confirmed[date])
			case "Death":
				ts.Death[date] = val
				_, err = stmt.Exec(id, date, ts.Death[date])
			case "Recovered":
				ts.Recovered[date] = val
				_, err = stmt.Exec(id, date, ts.Recovered[date])
			}
			if err != nil {
				log.Fatal(err)
			}
			dateIndex++
			//fmt.Println(date.String())
		}
	}

}

func Update(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)
	body := "" + id
	fmt.Println(body)
}
