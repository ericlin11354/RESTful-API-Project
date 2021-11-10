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

func List(w http.ResponseWriter, r *http.Request) {
	query, death, recovered, status := makeQuery(r.URL.Query())
	if status == 400 {
		w.WriteHeader(400)
		if _, err := w.Write([]byte("Error 400: Invalid Input")); err != nil {
			log.Fatal(err)
		}
		return
	}

	typeStr := getType(death, recovered)

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

	// Initializing array of TimeSeries
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
		if typeStr == "Confirmed" {
			ts.Confirmed = map[time.Time]int{}
		} else if typeStr == "Death" {
			ts.Death = map[time.Time]int{}
		} else {
			ts.Recovered = map[time.Time]int{}
		}
		tsArr = append(tsArr, ts)
	}

	// Filling maps
	columns := fmt.Sprintf("Date, %s", typeStr)
	for _, ts := range tsArr {
		query := fmt.Sprintf(`
			SELECT %s FROM TimeSeries%s
			WHERE ID = %s
		`, columns, typeStr, ts.ID)

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
			var err error
			if typeStr == "Confirmed" {
				err = rows.Scan(&tsd.Date, &tsd.Confirmed)
			} else if typeStr == "Death" {
				err = rows.Scan(&tsd.Date, &tsd.Death)
			} else {
				err = rows.Scan(&tsd.Date, &tsd.Recovered)
			}

			if err != nil {
				log.Fatal(err)
			}

			if typeStr == "Confirmed" {
				ts.Confirmed[tsd.Date] = tsd.Confirmed
			} else if typeStr == "Death" {
				ts.Death[tsd.Date] = tsd.Death
			} else {
				ts.Recovered[tsd.Date] = tsd.Recovered
			}
		}
	}

	// Check 'Accept' type
	if r.Header.Get("Accept") == "text/csv" {
		w.Header().Set("Content-Type", "text/csv")

		b := new(bytes.Buffer)
		writer := csv.NewWriter(b)
		csvArr := [][]string{}
		csvArr = append(csvArr, writeHeader(death, recovered))

		// Writing response in CSV
		for _, ts := range tsArr {
			data := map[time.Time]int{}
			if typeStr == "Confirmed" {
				data = ts.Confirmed
			} else if typeStr == "Death" {
				data = ts.Death
			} else {
				data = ts.Recovered
			}
			for date := range data {
				row := []string{
					ts.ID,
					writeAddress(ts),
					date.Format("2006/01/02"),
				}
				row = append(row, writeRow(ts, date, death, recovered)...)
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
		// Writing response in JSON
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
	var Admin2Index int = -1
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

	// Check for duplicate dates
	allKeys := make(map[string]bool)
	//fmt.Println(len(result))
	for i := beginDateIndex; i < len(result); i++ {
		//fmt.Println(allKeys)
		if value := allKeys[result[i]]; !value {
			allKeys[result[i]] = true
		} else { // duplicate found
			w.WriteHeader(400)
			w.Write([]byte("Error 400: Don't have files with duplicate dates"))
			return
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
		if Admin2Index >= 0 { // Admin2 exists
			ts.Admin2 = result[Admin2Index]
		}
		ts.Address1 = result[Address1Index]
		ts.Address2 = result[Address2Index]
		// Assuming that we have successfully parsed to a TimeSeries struct
		var query string
		if Admin2Index >= 0 {
			query = "INSERT INTO TimeSeries(Admin2, Address1, Address2) VALUES(?,?,?)"
		} else {
			query = "INSERT INTO TimeSeries(Address1, Address2) VALUES(?,?)"
		}
		stmt, err := db.Db.Prepare(query)
		if err != nil {
			log.Fatal(err)
		}
		var res sql.Result
		if Admin2Index >= 0 {
			res, err = stmt.Exec(ts.Admin2, ts.Address1, ts.Address2)
		} else {
			res, err = stmt.Exec(ts.Address1, ts.Address2)
		}
		if err != nil {
			log.Fatal(err)
		}

		id, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}

		query = fmt.Sprintf("INSERT INTO TimeSeries%s VALUES(?,?,?)", filetype)
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

// Helper functions
func makeQuery(params map[string][]string) (string, bool, bool, int) {
	death, recovered := false, false
	formattedParams := map[string][]string{}
	// Validating params
	for param, value := range params {
		param = strings.ToLower(param)

		var valid bool
		if param, valid = utils.ParamValidate(param); !valid {
			return "", false, false, 400
		}
		formattedParams[param] = value
	}

	if _, ok := formattedParams["death"]; ok {
		death = true
	}
	if _, ok := formattedParams["recovered"]; ok {
		recovered = true
	}

	// Mutually exclusive
	if death && recovered {
		return "", false, false, 400
	}

	query := `
		SELECT DISTINCT TimeSeries.ID, Admin2, Address1, Address2
		FROM TimeSeries JOIN TimeSeriesConfirmed ON
		TimeSeries.ID = TimeSeriesConfirmed.ID
		JOIN TimeSeriesDeath ON TimeSeries.ID = TimeSeriesDeath.ID
		JOIN TimeSeriesRecovered ON TimeSeries.ID = TimeSeriesRecovered.ID
	`

	i := 0 // counter for 'WHERE'
	for param, value := range formattedParams {
		// Displaying data
		if param == "death" {
			death = true
			continue
		}
		if param == "recovered" {
			recovered = true
			continue
		}

		// Time interval
		op := "="
		if param == "from" || param == "to" {
			if param == "from" {
				op = ">="
			} else {
				op = "<="
			}
			param = "date"
		}

		value := strings.Split(value[0], ",")

		for j, v := range value {
			// Format string for SQL
			stringParams := map[string]string{
				"admin2":   fmt.Sprintf(`'%s'`, v),
				"address1": fmt.Sprintf(`'%s'`, v),
				"address2": fmt.Sprintf(`'%s'`, v),
				"date":     fmt.Sprintf(`'%s'`, v),
			}
			_, ok := stringParams[param]
			if ok {
				if param == "date" {
					if _, err := utils.ParseDate(v); err != nil {
						return "", false, false, 400
					}
				}
				value[j] = stringParams[param]
			}

			// Format first param and after
			if i == 0 {
				query += "\tWHERE " + param + op + value[j]
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

	return query, death, recovered, 0
}

func getType(d bool, r bool) string {
	if !d && !r {
		return "Confirmed"
	}
	if d {
		return "Death"
	}
	return "Recovered"
}

func nullHandler(ts *TimeSeries, ns map[string]*sql.NullString) {
	ts.ID = ns["id"].String
	if ns["admin2"].Valid {
		ts.Admin2 = ns["admin2"].String
	}
	if ns["address1"].Valid {
		ts.Address1 = ns["address1"].String
	}
	if ns["address2"].Valid {
		ts.Address2 = ns["address2"].String
	}
}

func writeHeader(death bool, recovered bool) []string {
	col := getType(death, recovered)
	return []string{"ID", "Address", "Date", col}
}

func writeAddress(ts TimeSeries) string {
	address := ""
	if ts.Admin2 != "" {
		address += ts.Admin2 + ", "
	}
	if ts.Address1 != "" {
		address += ts.Address1 + ", "
	}
	if ts.Address2 != "" {
		address += ts.Address2
	}
	return address
}

func writeRow(ts TimeSeries, date time.Time, death bool, recovered bool) []string {
	typeStr := getType(death, recovered)
	arr := []string{}
	if typeStr == "Confirmed" {
		arr = append(arr, strconv.Itoa(ts.Confirmed[date]))
	} else if typeStr == "Death" {
		arr = append(arr, strconv.Itoa(ts.Death[date]))
	} else {
		arr = append(arr, strconv.Itoa(ts.Recovered[date]))
	}
	return arr
}
