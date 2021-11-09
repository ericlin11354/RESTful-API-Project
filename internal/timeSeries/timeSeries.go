package timeSeries

import (
	// Built-ins

	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	// External imports
	"github.com/go-chi/chi"

	// Internal imports
	db "gitlab.com/csc301-assignments/a2/internal/db"
)

type TimeSeries struct {
	ID         string         `json:"ID"`
	Admin2     sql.NullString `json:"Admin2"`
	Address1   sql.NullString `json:"Province/State"`
	Address2   sql.NullString `json:"Country/Region"`
	Population sql.NullInt64  `json:"Population"`

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

func List(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	query := `
	SELECT TimeSeries.ID, Admin2, Address1, Address2, Population
	FROM TimeSeries JOIN TimeSeriesDate
	ON TimeSeries.ID = TimeSeriesDate.ID
	`
	i := 0
	for param, value := range params {
		param = strings.ToLower(param)

		// Time interval
		op := "="
		if param == "from" {
			param = "Date"
			op = ">="
		}
		if param == "to" {
			param = "Date"
			op = "<="
		}

		// Format string for SQL
		stringParams := map[string]string{
			"admin2":   fmt.Sprintf(`'%s'`, value[0]),
			"address1": fmt.Sprintf(`'%s'`, value[0]),
			"address2": fmt.Sprintf(`'%s'`, value[0]),
		}
		_, ok := stringParams[param]
		if ok {
			value[0] = stringParams[param]
		}

		// Format first param and after
		if i == 0 {
			query += "\nWHERE " + param + op + value[0]
			i++
		} else {
			query += " AND " + param + op + value[0]
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
		err := row.Scan(&ts.ID, &ts.Admin2, &ts.Address1, &ts.Address2, &ts.Population)
		if err != nil {
			log.Fatal(err)
		}
		ts.Confirmed = map[time.Time]int{}
		ts.Death = map[time.Time]int{}
		ts.Recovered = map[time.Time]int{}
		tsArr = append(tsArr, ts)
	}

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

	if err = json.NewEncoder(w).Encode(tsArr); err != nil {
		log.Fatal(err)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	// ts := TimeSeries{}

	// reader := csv.NewReader(r.Body)

	// for {
	// 	result, err := reader.Read()
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	// Parse here

	// 	/*for i := range result {
	// 		// do something with result[i] <- value at ith column in 'result'
	// 		fmt.Printf("	%v'n", result[i])
	// 	}*/

	// 	/*
	// 		if file, err := os.Open(FilePath); err != nil {
	// 			log.Fatal("File read error", err)
	// 			return
	// 		}
	// 		csvReader := csv.NewReader(f)
	// 		if records, err := csvReader.ReadAll(); err != nil {
	// 			log.Fatal(filePath + ": Parse file error", err)
	// 		}
	// 	*/

	// 	// Assuming that we have successfully parsed to a TimeSeries struct
	// 	stmt, err := db.Db.Prepare("INSERT INTO TimeSeries(Admin2, Address1, Address2, Population) VALUES(?,?,?,?)")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	res, err := stmt.Exec(ts.Admin2, ts.Address1, ts.Address2, ts.Population)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	id, err := res.LastInsertId()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	stmt, err = db.Db.Prepare("INSERT INTO TimeSeriesDate VALUES(?,?,?,?,?)")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	/* Preconditions:
	// 	1. Dates are contiguous.
	// 	2. The only column values with '/' are dates.
	// 	*/
	// 	colNames, err := db.Db.Query("SELECT Admin2, Address1, Address2, Population FROM TimeSeries")
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	var beginDate time.Time
	// 	var endDate time.Time
	// 	beginFlag := false // true -> beginDate found; false Otherwise
	// 	endFlag := false
	// 	for i := range result {
	// 		if !beginFlag && strings.Contains(colNames[i], "/") {
	// 			temp := strings.Split(colNames[i], "/") // i.e. [ "1", "23", "20" ]
	// 			beginDate = time.Date(temp[2].(int), temp[0].(int), temp[1].(int))
	// 			beginFlag = true
	// 		}
	// 		if !endFlag && strings.Contains(colNames[len(colNames)-i-1], "/") { // searches backwards in array
	// 			temp := strings.Split(colNames[i], "/")
	// 			endDate = time.Date(temp[2].(int), temp[0].(int), temp[1].(int))
	// 			endFlag = true
	// 		}
	// 	}
	// 	for date := beginDate; date != endDate; date = date.AddDate(0, 0, 1) { // increment by 1 day
	// 		date_str := date.String()
	// 		res, err := stmt.Exec(id, date_str, ts.Confirmed[date_str], ts.Death[date_str], ts.Recovered[date_str]) // gets values using Key "date_str"
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	}

	// 	fmt.Println(result)
	// }

	// if _, err := w.Write([]byte("Successfully Uploaded Data")); err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("Successfully Uploaded Data")
}

func Update(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)
	body := "" + id
	fmt.Println(body)
}
