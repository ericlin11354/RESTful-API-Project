package timeSeries

import (
	// Built-ins

	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	query, dates, death, recovered, status := makeQuery(r.URL.Query())
	if status == 400 {
		utils.HandleErr(w, 400, errors.New("Invalid input"))
		return
	}

	typeStr := getType(death, recovered)

	stmt, err := db.Db.Prepare(query)
	if err != nil {
		utils.HandleErr(w, 500, err)
		return
	}

	defer stmt.Close()

	row, err := stmt.Query()
	if err != nil {
		utils.HandleErr(w, 500, err)
		return
	}
	defer row.Close()

	// Initializing array of TimeSeries
	tsArr := []TimeSeries{}
	for row.Next() {
		ts := TimeSeries{}

		// Handling null values
		temp := map[string]*sql.NullString{
			"id":       {},
			"admin2":   {},
			"address1": {},
			"address2": {},
		}
		err := row.Scan(temp["id"], temp["admin2"],
			temp["address1"], temp["address2"])
		if err != nil {
			utils.HandleErr(w, 500, err)
			return
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
		// Querying from db
		query := fmt.Sprintf(`
			SELECT %s FROM TimeSeries%s
			WHERE ID = %s %s
		`, columns, typeStr, ts.ID, dates)

		stmt, err := db.Db.Prepare(query)
		if err != nil {
			utils.HandleErr(w, 500, err)
			return
		}

		defer stmt.Close()
		rows, err := stmt.Query()
		if err != nil {
			utils.HandleErr(w, 500, err)
			return
		}

		// Reading each row
		for rows.Next() {
			tsd := TimeSeriesDate{}
			var err error
			if typeStr == "Confirmed" {
				err = rows.Scan(&tsd.Date, &tsd.Confirmed)
				ts.Confirmed[tsd.Date] = tsd.Confirmed
			} else if typeStr == "Death" {
				err = rows.Scan(&tsd.Date, &tsd.Death)
				ts.Death[tsd.Date] = tsd.Death
			} else {
				err = rows.Scan(&tsd.Date, &tsd.Recovered)
				ts.Recovered[tsd.Date] = tsd.Recovered
			}

			if err != nil {
				utils.HandleErr(w, 500, err)
				return
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
			var data map[time.Time]int
			if typeStr == "Confirmed" {
				data = ts.Confirmed
			} else if typeStr == "Death" {
				data = ts.Death
			} else {
				data = ts.Recovered
			}

			// Create a row
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
		// Write to buffer
		if err := writer.WriteAll(csvArr); err != nil {
			utils.HandleErr(w, 500, err)
			return
		}
		// Write to response
		if _, err := w.Write(b.Bytes()); err != nil {
			utils.HandleErr(w, 500, err)
			return
		}
	} else {
		// Writing response in JSON
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(tsArr); err != nil {
			utils.HandleErr(w, 500, err)
			return
		}
	}

	// Successfully written the data
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

	//fmt.Println("running Create()")

	// get header names
	result, err := reader.Read()
	if err != nil {
		utils.HandleErr(w, 400, err)
		return
	}

	// Directly access column values
	var Admin2Index int = -1
	var Address1Index int
	var Address2Index int

	// allows for direct access to dates
	beginDate, endDate, beginDateIndex, err := getDates(result)
	if err != nil {
		utils.HandleErr(w, 500, errors.New("ParseDate() failed"))
		return
	}

	// Check for duplicate dates
	if utils.HasDupe(beginDateIndex, result) {
		utils.HandleErr(w, 400, errors.New("File has duplicate dates"))
		return
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
		if err == io.EOF {
			break
		}
		if err != nil {
			utils.HandleErr(w, 400, err)
			return
		}
		if Admin2Index >= 0 { // Admin2 exists
			ts.Admin2 = result[Admin2Index]
		}
		ts.Address1 = result[Address1Index]
		ts.Address2 = result[Address2Index]

		id, err := injectTimeSeries(Admin2Index, ts)
		if err != nil {
			utils.HandleErr(w, 500, errors.New("injectTimeSeries() failed"))
			return
		}

		//fmt.Println("hello")
		ts.Confirmed = make(map[time.Time]int)
		ts.Death = make(map[time.Time]int)
		ts.Recovered = make(map[time.Time]int)

		_, err = InjectTimeSeriesDate(beginDate, endDate, beginDateIndex, result, ts, id, filetype)
		if err != nil {
			utils.HandleErr(w, 500, errors.New("injectTimeSeriesDate() failed"))
			return
		}

	}

}

func getDates(result []string) (time.Time, time.Time, int, error) {
	// get beginDate and endDate
	var (
		beginDate      time.Time
		endDate        time.Time
		beginDateIndex int = -1
		err            error
	)
	beginFlag := false // true -> beginDate found; false Otherwise
	endFlag := false
	for i := range result {
		if !beginFlag && len(strings.Split(result[i], "/")) == 3 {
			beginDate, err = utils.ParseDate(result[i]) // parses Date string -> time.Time
			if err != nil {
				return beginDate, time.Time{}, i, err
			}
			beginDateIndex = i
			beginFlag = true
		}
		if !endFlag && len(strings.Split(result[len(result)-i-1], "/")) == 3 { // searches backwards in array
			endDate, err = utils.ParseDate(result[len(result)-i-1])
			if err != nil {
				return time.Time{}, endDate, -1, err
			}
			endFlag = true
		}
	}

	return beginDate, endDate, beginDateIndex, err
}

func injectTimeSeries(Admin2Index int, ts TimeSeries) (int64, error) {
	// check if address exists
	var (
		ID            int64
		Admin2        sql.NullString
		Address1      sql.NullString
		Address2      string
		AddressExists bool
	)
	rows, err := db.Db.Query(`
		SELECT ID, Admin2, Address1, Address2 FROM TimeSeries
		`)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&ID, &Admin2, &Address1, &Address2)
		if err != nil {
			return -1, err
		}
		if !Admin2.Valid {
			continue
		} else if Admin2.String == ts.Admin2 && Address1.String == ts.Address1 && Address2 == ts.Address2 {
			AddressExists = true
			break
		}
	}
	var (
		id    int64
		query string
	)
	if AddressExists { // If an address exists, we simply use its id
		id = ID
	} else { // Else, inject a new address
		/**
		The following cases determine the number of fields that we inject:
		1. Admin2 and Address1 exist
		2. Admin2 does not exist but Address1 exists / Admin2 exists but Address1 does not exist
		3. Both Admin2 and Address1 do not exist
		*/
		if Admin2Index >= 0 && len(ts.Address1) > 0 {
			query = "INSERT INTO TimeSeries(Admin2, Address1, Address2) VALUES(?,?,?)"
		} else if Admin2Index < 0 && len(ts.Address1) > 0 {
			query = "INSERT INTO TimeSeries(Address1, Address2) VALUES(?,?)"
		} else if Admin2Index >= 0 && len(ts.Address1) == 0 {
			query = "INSERT INTO TimeSeries(Admin2, Address2) VALUES(?,?)"
		} else if Admin2Index < 0 && len(ts.Address1) == 0 {
			query = "INSERT INTO TimeSeries(Address2) VALUES(?)"
		}
		stmt, err := db.Db.Prepare(query)
		if err != nil {
			return -1, err
		}
		var res sql.Result
		if Admin2Index >= 0 && len(ts.Address1) > 0 {
			res, err = stmt.Exec(ts.Admin2, ts.Address1, ts.Address2)
		} else if Admin2Index < 0 && len(ts.Address1) > 0 {
			res, err = stmt.Exec(ts.Address1, ts.Address2)
		} else if Admin2Index >= 0 && len(ts.Address1) == 0 {
			res, err = stmt.Exec(ts.Admin2, ts.Address2)
		} else if Admin2Index < 0 && len(ts.Address1) == 0 {
			res, err = stmt.Exec(ts.Address2)
		}
		if err != nil {
			return -1, err
		}
		id, err = res.LastInsertId()
		if err != nil {
			return id, err
		}
	}

	return id, nil
}

func InjectTimeSeriesDate(beginDate time.Time, endDate time.Time, beginDateIndex int, result []string, ts TimeSeries, id int64, filetype string) (bool, error) {
	query := fmt.Sprintf("INSERT INTO TimeSeries%s VALUES(?,?,?)", filetype)
	stmt, err := db.Db.Prepare(query)
	if err != nil {
		return false, err
	}

	var (
		ID   int64
		Date time.Time
		Type int
	)
	dateIndex := beginDateIndex
	for date := beginDate; date != endDate.Add(time.Hour*24); date = date.AddDate(0, 0, 1) { // iterate between beginDate and endDate inclusive, incrementing by 1 Day
		val, err := strconv.Atoi(result[dateIndex])
		if err != nil {
			return false, err
		}
		//fmt.Printf("hewwo %v\n", date.String())

		// Find row with existing id and date
		rows, err := db.Db.Query(fmt.Sprintf(`
		SELECT * FROM TimeSeries%s
		`, filetype))
		if err != nil {
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&ID, &Date, &Type)
			if err != nil {
				return false, err
			}
			//fmt.Println(ID, id, Date.Format("2006-1-2"), date.Format("2006-1-2"))
			layout := "2006-1-2"
			if ID == id && Date.Format(layout) == date.Format(layout) {
				_, err = db.Db.Exec(fmt.Sprintf(`
				DELETE FROM TimeSeries%s
				WHERE ID = %d AND Date = '%s'`, filetype, ID, Date.Format("2006-1-2"))) // remove based on
				if err != nil {
					return false, err
				}
			}
		}

		switch filetype {
		case "Confirmed":
			ts.Confirmed[date] = val
			//fmt.Println("lesgo")
			_, err = stmt.Exec(id, date, ts.Confirmed[date])
		case "Death":
			ts.Death[date] = val
			_, err = stmt.Exec(id, date, ts.Death[date])
		case "Recovered":
			ts.Recovered[date] = val
			_, err = stmt.Exec(id, date, ts.Recovered[date])
		}
		if err != nil {
			return false, err
		}
		dateIndex++
		//fmt.Println(date.String())
	}

	return true, nil
}

// Helper functions
func makeQuery(params map[string][]string) (string, string, bool, bool, int) {
	query := `
		SELECT DISTINCT TimeSeries.ID, Admin2, Address1, Address2
		FROM TimeSeries JOIN TimeSeriesConfirmed ON
		TimeSeries.ID = TimeSeriesConfirmed.ID
		JOIN TimeSeriesDeath ON TimeSeries.ID = TimeSeriesDeath.ID
		JOIN TimeSeriesRecovered ON TimeSeries.ID = TimeSeriesRecovered.ID
	`
	dates := ""
	death, recovered := false, false
	if len(params) == 0 {
		return query, dates, death, recovered, 0
	}

	formattedParams := map[string][]string{}
	// Validating params
	for param, value := range params {
		param, valid := utils.ParamValidate(param)
		if !valid {
			return "", dates, false, false, 400
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
		return "", dates, false, false, 400
	}

	dateCounter := 0
	whereCounter := 0 // counter for 'WHERE'
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
		for i, v := range value {
			// Format string for SQL
			stringParams := map[string]string{
				"admin2":   fmt.Sprintf(`'%s'`, v),
				"address1": fmt.Sprintf(`'%s'`, v),
				"address2": fmt.Sprintf(`'%s'`, v),
				"date":     "",
			}
			_, ok := stringParams[param]
			if ok {
				if param == "date" {
					// mm/dd/yy
					temp, err := utils.ParseDate(v)
					if err != nil {
						return "", "", false, false, 400
					}
					value[i] = temp.Format("2006/1/2")
					value[i] = fmt.Sprintf(`"%s"`, value[i])

					// Put to dates string, skip to next param
					if dateCounter == 0 {
						dates += "AND " + param + op + value[i]
						dateCounter++
					} else {
						dates += " OR " + param + op + value[i]
					}
					if i == len(value)-1 {
						param = fmt.Sprintf("TimeSeries%s.Date", getType(death, recovered))
					}
					continue
				} else {
					value[i] = stringParams[param]
				}
			}

			// Format first param and after
			if whereCounter == 0 {
				query += "\tWHERE " + param + op + value[i]
				whereCounter++
			} else {
				if i != 0 {
					query += " OR " + param + op + value[i]
				} else {
					query += " AND " + param + op + value[i]
				}
			}
		}
	}
	return query, dates, death, recovered, 0
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
