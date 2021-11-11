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

		err := row.Scan(&dr.Date, &dr.Date,
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
				dr.ID, dr.Date.Format("2006/01/02"), dr.Admin2, dr.Address1, dr.Address2,
				strconv.Itoa(dr.Confirmed),
				strconv.Itoa(dr.Death),
				strconv.Itoa(dr.Recovered),
				strconv.Itoa(dr.Active),
			}
			csvArr = append(csvArr, row)
		}
		if err := writer.WriteAll(csvArr); err != nil {
			utils.HandleErr(w, 500, err)
		}

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

// Helper functions
func makeQuery(params map[string][]string) (string, int) {
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
			return "", 400
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
	return query, 0
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
