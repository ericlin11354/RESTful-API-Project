package dailyReports

import (
	// Built-ins
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
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

func List(w http.ResponseWriter, r *http.Request) {
	query, status := makeQuery(r.URL.Query())
	if status != 0 {
		utils.HandleErr(w, 400, errors.New("Input error"))
		return
	}
	stmt, err := db.Db.Prepare(query)

	if err != nil {
		utils.HandleErr(w, 500, err)
	}

	defer stmt.Close()

	row, err := stmt.Query()
	if err != nil {
		utils.HandleErr(w, 500, err)
	}
	defer row.Close()

	drArr := []DailyReports{}
	for row.Next() {
		dr := DailyReports{}

		// Handling null values
		ns := map[string]*sql.NullString{
			"admin2":   {},
			"address1": {},
			"address2": {},
		}

		ni := map[string]*sql.NullInt64{
			"confirmed": {},
			"death":     {},
			"recovered": {},
			"active":    {},
		}

		err := row.Scan(&dr.ID, &dr.Date,
			ns["admin2"], ns["address1"], ns["address2"],
			ni["confirmed"], ni["death"],
			ni["recovered"], ni["active"],
		)
		if err != nil {
			utils.HandleErr(w, 500, err)
		}

		nullStringHandler(&dr, ns)
		nullIntHandler(&dr, ni)

		drArr = append(drArr, dr)
	}

	// Checking for return response type
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
				dr.ID,
				dr.Date.Format("2006/01/02"),
				dr.Admin2,
				dr.Address1,
				dr.Address2,
				strconv.Itoa(dr.Confirmed),
				strconv.Itoa(dr.Death),
				strconv.Itoa(dr.Recovered),
				strconv.Itoa(dr.Active),
			}
			csvArr = append(csvArr, row)
		}
		// Write to buffer
		if err := writer.WriteAll(csvArr); err != nil {
			utils.HandleErr(w, 500, err)
		}
		// Write to response body
		if _, err := w.Write(b.Bytes()); err != nil {
			utils.HandleErr(w, 500, err)
		}
	} else {
		if err := json.NewEncoder(w).Encode(drArr); err != nil {
			utils.HandleErr(w, 500, err)
		}
	}

	w.WriteHeader(200)
}

func Create(w http.ResponseWriter, r *http.Request) {
	notAllRead := false
	date := r.Header.Get("Date")
	_, err := utils.ParseDate(date)
	if date == "" || err != nil {
		utils.HandleErr(w, 400, err)
		return
	}

	dr := DailyReports{}
	reader := csv.NewReader(r.Body)

	// get header names
	result, err := reader.Read()
	if err != nil {
		utils.HandleErr(w, 400, err)
		return
	}

	// Directly access column values
	indices := map[string]int{
		"admin2": -1,
		"add1":   0,
		"add2":   0,
		"c":      0,
		"d":      0,
		"r":      0,
		"a":      0,
	}

	for i := range result {
		switch strings.ToLower(result[i]) {
		case "admin2":
			indices["admin2"] = i
		case "province_state":
			indices["add1"] = i
		case "country_region":
			indices["add2"] = i
		case "confirmed":
			indices["c"] = i
		case "deaths":
			indices["d"] = i
		case "recovered":
			indices["r"] = i
		case "active":
			indices["a"] = i
		}
	}

	for {
		result, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// Admin2 exists
		if indices["admin2"] >= 0 && result[indices["admin2"]] != "" {
			dr.Admin2 = result[indices["admin2"]]
		} else {
			indices["admin2"] = -1
		}
		dr.Date, err = utils.ParseDate(date)
		if err != nil {
			utils.HandleErr(w, 400, err)
			return
		}

		dr.Address1 = result[indices["add1"]]
		dr.Address2 = result[indices["add2"]]

		dr.Confirmed, err = strconv.Atoi(result[indices["c"]])
		if err != nil {
			notAllRead = true
			continue
		}
		dr.Death, err = strconv.Atoi(result[indices["d"]])
		if err != nil {
			notAllRead = true
			continue
		}
		dr.Active, err = strconv.Atoi(result[indices["a"]])
		if err != nil {
			notAllRead = true
			continue
		}
		dr.Recovered, err = strconv.Atoi(result[indices["r"]])
		if err != nil {
			notAllRead = true
			continue
		}

		_, err := injectDailyReport(indices["admin2"], dr)
		if err != nil {
			utils.HandleErr(w, 500, err)
			return
		}

	}
	// Write to respond body
	if notAllRead {
		if _, err := w.Write([]byte("Input Error: could not parse some data into the system")); err != nil {
			utils.HandleErr(w, 500, err)
			return
		}
		w.WriteHeader(400)
	} else {
		if _, err := w.Write([]byte("Successfully create/update data to the system")); err != nil {
			utils.HandleErr(w, 500, err)
			return
		}
		w.WriteHeader(200)
	}
}

func injectDailyReport(Admin2Index int, dr DailyReports) (bool, error) {
	// check if address exists
	var (
		ID            int64
		Date          time.Time
		Admin2        sql.NullString
		Address1      sql.NullString
		Address2      string
		Confirmed     int
		Death         int
		Recovered     int
		Active        int
		AddressExists bool
	)
	rows, err := db.Db.Query(`
		SELECT ID, Date, Admin2, Address1, Address2, Confirmed, Death, Recovered, Active FROM DailyReports
		`)
	if err != nil {
		return false, err
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&ID, &Date, &Admin2, &Address1, &Address2, &Confirmed, &Death, &Recovered, &Active)
		if err != nil {
			return false, err
		}

		layout := "2006-1-2"
		if Admin2.String == dr.Admin2 && Date.Format(layout) == dr.Date.Format(layout) &&
			Address1.String == dr.Address1 && Address2 == dr.Address2 &&
			Confirmed == dr.Confirmed && Death == dr.Death && Recovered == dr.Recovered && Active == dr.Active {
			AddressExists = true
			break
		}
	}

	var query string
	if AddressExists { //
		_, err = db.Db.Exec(fmt.Sprintf(`
		DELETE FROM DailyReports
		WHERE ID = %d AND Date = '%s' AND Address1 = '%s' AND Address2 = '%s'
		`, ID, Date.Format("2006-1-2"), dr.Address1, dr.Address2)) // remove based on
		if err != nil {
			return false, err
		}
	}

	/**
	The following cases determine the number of fields that we inject:
	1. We inject existing ID and Admin2 exists
	2. We inject existing ID and Admin2 does not exist / We don't inject existing ID and Admin2 exists
	3. We don't inject existing ID and Admin2 does not exist
	*/
	if Admin2Index >= 0 && AddressExists {
		query = "INSERT INTO DailyReports(ID, Date, Admin2, Address1, Address2, Confirmed, Death, Recovered, Active) VALUES(?,?,?,?,?,?,?,?,?)"
	} else if (Admin2Index >= 0 && !AddressExists) || (Admin2Index < 0 && AddressExists) {
		if Admin2Index >= 0 {
			query = "INSERT INTO DailyReports(Date, Admin2, Address1, Address2, Confirmed, Death, Recovered, Active) VALUES(?,?,?,?,?,?,?,?)"
		} else {
			query = "INSERT INTO DailyReports(ID, Date, Address1, Address2, Confirmed, Death, Recovered, Active) VALUES(?,?,?,?,?,?,?,?)"
		}
	} else if Admin2Index < 0 && !AddressExists {
		query = "INSERT INTO DailyReports(Date, Address1, Address2, Confirmed, Death, Recovered, Active) VALUES(?,?,?,?,?,?,?)"
	}
	stmt, err := db.Db.Prepare(query)
	if err != nil {
		return false, err
	}
	if Admin2Index >= 0 && AddressExists {
		_, err = stmt.Exec(ID, dr.Date, dr.Admin2, dr.Address1, dr.Address2, dr.Confirmed, dr.Death, dr.Recovered, dr.Active)
		if err != nil {
			return false, err
		}
	} else if (Admin2Index >= 0 && !AddressExists) || (Admin2Index < 0 && AddressExists) {
		if Admin2Index >= 0 {
			_, err = stmt.Exec(dr.Date, dr.Admin2, dr.Address1, dr.Address2, dr.Confirmed, dr.Death, dr.Recovered, dr.Active)
		} else {
			_, err = stmt.Exec(ID, dr.Date, dr.Address1, dr.Address2, dr.Confirmed, dr.Death, dr.Recovered, dr.Active)
		}
	} else if Admin2Index < 0 && !AddressExists {
		_, err = stmt.Exec(dr.Date, dr.Address1, dr.Address2, dr.Confirmed, dr.Death, dr.Recovered, dr.Active)
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// Helper functions
func makeQuery(params map[string][]string) (string, int) {
	query := `
		SELECT ID, Date, Admin2, Address1, Address2,
		Confirmed, Death, Recovered, Active
		FROM DailyReports
	`
	status := 0
	if len(params) == 0 {
		return query, status
	}

	whereCounter := 0
	for param, value := range params {
		param = strings.ToLower(param)

		// Validate parameters
		var valid bool
		if param, valid = utils.ParamValidate(param); !valid {
			return "", 400
		}

		// Ignore these parameters
		if param == "confirmed" ||
			param == "death" ||
			param == "recovered" ||
			param == "active" {
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
				"date":     fmt.Sprintf(`'%s'`, v),
			}
			_, ok := stringParams[param]
			if ok {
				if param == "date" {
					// mm/dd/yy
					temp, err := utils.ParseDate(v)
					if err != nil {
						return "", 400
					}
					value[i] = temp.Format("2006/1/2")
					value[i] = fmt.Sprintf(`"%s"`, value[i])
				} else {
					value[i] = stringParams[param]
				}
			}

			if param == "id" {
				if _, err := strconv.Atoi(v); err != nil {
					return "", 400
				}
			}

			// Format first param and after
			if whereCounter == 0 {
				query += "WHERE " + param + op + value[i]
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
	return query, status
}

func nullStringHandler(dr *DailyReports, ns map[string]*sql.NullString) {
	if ns["admin2"].Valid {
		dr.Admin2 = ns["admin2"].String
	}
	if ns["address1"].Valid {
		dr.Address1 = ns["address1"].String
	}
	if ns["address2"].Valid {
		dr.Address2 = ns["address2"].String
	}
}

func nullIntHandler(dr *DailyReports, ni map[string]*sql.NullInt64) {
	if ni["confirmed"].Valid {
		dr.Confirmed = int(ni["confirmed"].Int64)
	}
	if ni["death"].Valid {
		dr.Death = int(ni["death"].Int64)
	}
	if ni["recovered"].Valid {
		dr.Recovered = int(ni["recovered"].Int64)
	}
	if ni["active"].Valid {
		dr.Active = int(ni["active"].Int64)
	}
}
