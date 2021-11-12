package dailyReports

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	db "gitlab.com/csc301-assignments/a2/internal/db"
)

func TestNullStringHandler(t *testing.T) {
	// Non-null strings
	dr := DailyReports{}
	expect1, expect2, expect3 := "admin2", "address1", "address2"

	ns := map[string]*sql.NullString{
		"admin2": {
			String: expect1,
			Valid:  true,
		},
		"address1": {
			String: expect2,
			Valid:  true,
		},
		"address2": {
			String: expect3,
			Valid:  true,
		},
	}

	nullStringHandler(&dr, ns)

	if dr.Admin2 != "admin2" ||
		dr.Address1 != "address1" ||
		dr.Address2 != "address2" {
		t.Fatalf("Test failed: expect %s %s and %s, got %s %s and %s",
			expect1, expect2, expect3, dr.Admin2, dr.Address1, dr.Address2)
	}

	// null strings
	dr = DailyReports{}
	expect := ""
	ns = map[string]*sql.NullString{
		"admin2": {
			String: "garbage",
			Valid:  false,
		},
		"address1": {
			String: "garbage",
			Valid:  false,
		},
		"address2": {
			String: "garbage",
			Valid:  false,
		},
	}

	nullStringHandler(&dr, ns)
	if dr.Admin2 != "" ||
		dr.Address1 != "" ||
		dr.Address2 != "" {
		t.Fatalf("Test failed: expect %s %s and %s, got %s %s and %s",
			expect, expect, expect, dr.Admin2, dr.Address1, dr.Address2)
	}
}

func TestNullIntHandler(t *testing.T) {
	dr := DailyReports{}
	expect1, expect2, expect3, expect4 := 1, 2, 3, 4
	ni := map[string]*sql.NullInt64{
		"confirmed": {
			Int64: 1,
			Valid: true,
		},
		"death": {
			Int64: 2,
			Valid: true,
		},
		"recovered": {
			Int64: 3,
			Valid: true,
		},
		"active": {
			Int64: 4,
			Valid: true,
		},
	}

	nullIntHandler(&dr, ni)
	if dr.Confirmed != 1 || dr.Death != 2 ||
		dr.Recovered != 3 || dr.Active != 4 {
		t.Fatalf("Test failed: expect %d %d %d and %d, got %d %d %d and %d",
			expect1, expect2, expect3, expect4,
			dr.Confirmed, dr.Death, dr.Recovered, dr.Active)
	}
}

func TestMakeQueryNoParams(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)
	query, status := makeQuery(r.URL.Query())
	expected := strings.TrimSpace(`
		SELECT ID, Date, Admin2, Address1, Address2,
		Confirmed, Death, Recovered, Active
		FROM DailyReports
	`)
	query = strings.TrimSpace(query)
	if query != expected {
		t.Fatalf("Test failed: expected %s, got %s", expected, query)
	}

	expectedStatus := 0
	if status != expectedStatus {
		t.Fatalf("Test failed: expected %d, got %d", expectedStatus, status)
	}
}

func TestMakeQueryWithDateParams(t *testing.T) {
	// Date params
	r := httptest.NewRequest(
		"GET",
		"http://example.com/foo?from=1/1/20&to=1/1/22&date=11/16/20,2/14/21",
		nil)
	query, _ := makeQuery(r.URL.Query())
	lines := strings.Split(query, "\n")
	lastline := lines[len(lines)-1]

	checker := "date>=\"2020/1/1\""
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "date<=\"2022/1/1\""
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "date=\"2020/11/16\""
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "date=\"2021/2/14\""
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}
}

func TestMakeQueryWithAddressParams(t *testing.T) {
	r := httptest.NewRequest("GET", "http://example.com/foo?country=canada,us&province=ontario&admin2=toronto", nil)
	query, _ := makeQuery(r.URL.Query())
	lines := strings.Split(query, "\n")
	lastline := lines[len(lines)-1]

	checker := "address2='canada'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "address2='us'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "address1='ontario'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "admin2='toronto'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}
}

func TestMakeQueryInvalidParams(t *testing.T) {
	// params not right
	r := httptest.NewRequest("GET", "http://example.com/foo?abc=def", nil)
	_, status := makeQuery(r.URL.Query())
	if status != 400 {
		t.Fatalf("Test failed: response status (%d) not 400", status)
	}

	// date incorrect format
	r = httptest.NewRequest("GET", "http://example.com/foo?date=abc", nil)
	_, status = makeQuery(r.URL.Query())
	if status != 400 {
		t.Fatalf("Test failed: response status (%d) not 400", status)
	}
}

func TestListNoParams(t *testing.T) {
	db.InitDb("development")
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedCode := 200
	if resp.StatusCode != expectedCode {
		t.Fatalf("Test failed: expected code %d, got %d", expectedCode, resp.StatusCode)
	}

	// Assuming db not empty
	drArr := []DailyReports{}
	err := json.Unmarshal(body, &drArr)
	if err != nil {
		t.Errorf("Error during converting JSON: %v", err)
	}

	if len(drArr) == 0 {
		t.Fatalf("Test failed: got empty response")
	}
}

func TestListBadInputs(t *testing.T) {
	db.InitDb("development")
	r := httptest.NewRequest("GET", "http://example.com/foo?asdf=asd", nil)
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedCode := 400
	if resp.StatusCode != expectedCode {
		t.Fatalf("Test failed: expected code %d, got %d", expectedCode, resp.StatusCode)
	}

	expected := "Error status 400"
	if string(body) != expected {
		t.Fatalf("Test failed: expected body %s, got %s", expected, string(body))
	}
}

func TestListAcceptCSV(t *testing.T) {
	db.InitDb("development")
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)
	r.Header.Set("Accept", "text/csv")
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	expected := "text/csv"
	if res := resp.Header.Get("Content-Type"); expected != res {
		t.Fatalf("Tested failed: expected header %s, got %s", expected, res)
	}

	lines := strings.Split(string(body), "\n")
	header := lines[0]
	expectedHeader := "ID,Date,Admin2,Province/State,Country/Region," +
		"Confirmed,Death,Recovered,Active"

	if header != expectedHeader {
		t.Fatalf("Test failed: expected csvheader %s, got %s", expectedHeader, header)
	}
}

func TestCreateDefault(t *testing.T) {
	db.InitDb("development")
	//Create body
	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)

	//fill 2d array
	csvArr := [][]string{}
	header := []string{
		"Admin2", "Province_State", "Country_Region", "Confirmed", "Deaths", "Recovered", "Active",
	}
	csvArr = append(csvArr, header)
	row := []string{
		"", "Ontario", "Canada", "420", "69", "1337", "96",
	}
	csvArr = append(csvArr, row)

	// Write to buffer
	if err := writer.WriteAll(csvArr); err != nil {
		t.Errorf("Error while converting csvArr to bytes")
	}

	w := httptest.NewRecorder()

	// curl --data-binary @timeSeries_test2.csv
	r := httptest.NewRequest("POST", "http://example.com/foo", b)
	r.Header.Set("Date", "1/20/21")

	// Goal: call Create()
	Create(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedCode := 200
	if resp.StatusCode != expectedCode {
		t.Fatalf("Test failed: expected code %d, got %d", expectedCode, resp.StatusCode)
	}

	expectedBody := "200 Success: created/updated data to the system"
	if string(body) != expectedBody {
		t.Fatalf("Test failed: expected body %s, got %s", expectedBody, string(body))
	}
}
func TestCreateBadHeader(t *testing.T) {
	db.InitDb("development")
	//test no header

	//Create body
	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)

	//fill 2d array
	csvArr := [][]string{}
	header := []string{
		"Admin2", "Province_State", "Country_Region", "Confirmed", "Deaths", "Recovered", "Active",
	}
	csvArr = append(csvArr, header)
	row := []string{
		"", "Ontario", "Canada", "420", "69", "1337", "96",
	}
	csvArr = append(csvArr, row)

	// Write to buffer
	if err := writer.WriteAll(csvArr); err != nil {
		t.Errorf("Error while converting csvArr to bytes")
	}

	w := httptest.NewRecorder()

	r := httptest.NewRequest("POST", "http://example.com/foo", b)

	// Goal: call Create()
	Create(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedCode := 400
	if resp.StatusCode != expectedCode {
		t.Fatalf("Test failed: expected code %d, got %d", expectedCode, resp.StatusCode)
	}

	expectedBody := "Error status 400"
	if string(body) != expectedBody {
		t.Fatalf("Test failed: expected body %s, got %s", expectedBody, string(body))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "http://example.com/foo", b)
	//test invalid header
	r.Header.Set("Date", "1/20/2021")

	// Goal: call Create()
	Create(w, r)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	expectedCode = 400
	if resp.StatusCode != expectedCode {
		t.Fatalf("Test failed: expected code %d, got %d", expectedCode, resp.StatusCode)
	}

	expectedBody = "Error status 400"
	if string(body) != expectedBody {
		t.Fatalf("Test failed: expected body %s, got %s", expectedBody, string(body))
	}
}

func TestCreateDuplicateAddress(t *testing.T) {
	db.InitDb("development")

	// get database before injecting

	// Create body
	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)

	w := httptest.NewRecorder()

	r := httptest.NewRequest("GET", "http://example.com/foo", b)
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	oldDailyReportsArr := []DailyReports{}
	err := json.Unmarshal(body, &oldDailyReportsArr)
	if err != nil {
		t.Errorf("Error during converting JSON: %v", err)
	}

	// test duplicate address

	//fill 2d array
	csvArr := [][]string{}
	header := []string{
		"Admin2", "Province_State", "Country_Region", "Confirmed", "Deaths", "Recovered", "Active",
	}
	csvArr = append(csvArr, header)
	row := []string{
		"Abbeville", "South Carolina", "US", "47", "0", "0", "47",
	}
	csvArr = append(csvArr, row)

	// Write to buffer
	if err := writer.WriteAll(csvArr); err != nil {
		t.Errorf("Error while converting csvArr to bytes")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "http://example.com/foo", b)

	Create(w, r)

	// validate the data is actually updated
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "http://example.com/foo", b)
	List(w, r)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	newDailyReportsArr := []DailyReports{}
	err = json.Unmarshal(body, &newDailyReportsArr)
	if err != nil {
		t.Errorf("Error during converting JSON: %v", err)
	}

	if !reflect.DeepEqual(oldDailyReportsArr, newDailyReportsArr) {
		t.Fatalf("Test failed: expected code %v, got %v", oldDailyReportsArr, newDailyReportsArr)
	}
}
