package timeSeries

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "gitlab.com/csc301-assignments/a2/internal/db"
	"gitlab.com/csc301-assignments/a2/internal/utils"
)

// Testing helper functions
func TestGetType(t *testing.T) {
	// Both false
	d, r := false, false
	expect := "Confirmed"
	result := getType(d, r)

	if expect != result {
		t.Fatalf("Test failed: expected %s, got %s", expect, result)
	}

	d = true
	r = false
	expect = "Death"
	result = getType(d, r)
	if expect != result {
		t.Fatalf("Test failed: expected %s, got %s", expect, result)
	}

	d = false
	r = true
	expect = "Recovered"
	result = getType(d, r)
	if expect != result {
		t.Fatalf("Test failed: expected %s, got %s", expect, result)
	}
}

func TestNullHandler(t *testing.T) {
	// All null
	ts := TimeSeries{}
	ns := map[string]*sql.NullString{}
	ns["id"] = &sql.NullString{
		String: "1",
		Valid:  true,
	}
	ns["admin2"] = &sql.NullString{}
	ns["address1"] = &sql.NullString{}
	ns["address2"] = &sql.NullString{}

	nullHandler(&ts, ns)
	if ts.ID != "1" {
		t.Fatalf("Test failed: expected 1, got %s", ts.ID)
	}
	if ts.Admin2 != "" {
		t.Fatalf("Test failed: expected \"\", got %s", ts.Admin2)
	}
	if ts.Address1 != "" {
		t.Fatalf("Test failed: expected \"\", got %s", ts.Address1)
	}
	if ts.Address2 != "" {
		t.Fatalf("Test failed: expected \"\", got %s", ts.Address2)
	}

	// All not null
	ns["admin2"].String = "admin2"
	ns["admin2"].Valid = true
	ns["address1"].String = "address1"
	ns["address1"].Valid = true
	ns["address2"].String = "address2"
	ns["address2"].Valid = true

	nullHandler(&ts, ns)
	if ts.ID != "1" {
		t.Fatalf("Test failed: expected 1, got %s", ts.ID)
	}
	if ts.Admin2 != "admin2" {
		t.Fatalf("Test failed: expected admin2, got %s", ts.Admin2)
	}
	if ts.Address1 != "address1" {
		t.Fatalf("Test failed: expected address1, got %s", ts.Address1)
	}
	if ts.Address2 != "address2" {
		t.Fatalf("Test failed: expected address2, got %s", ts.Address2)
	}
}

func TestWriteHeader(t *testing.T) {
	// Both false
	death, recovered := false, false
	expected := []string{"ID", "Address", "Date", "Confirmed"}
	result := writeHeader(death, recovered)
	if len(expected) != len(result) {
		t.Fatalf("Test Failed: length unequal (%d != %d)",
			len(expected), len(result))
	}
	for i := range expected {
		if expected[i] != result[i] {
			t.Fatalf("Test Failed: expected %s, got != %s",
				expected[i], result[i])
		}
	}

	// Death
	death = true
	recovered = false
	expected = []string{"ID", "Address", "Date", "Death"}
	result = writeHeader(death, recovered)
	if len(expected) != len(result) {
		t.Fatalf("Test Failed: length unequal (%d != %d)",
			len(expected), len(result))
	}
	for i := range expected {
		if expected[i] != result[i] {
			t.Fatalf("Test Failed: expected %s, got != %s",
				expected[i], result[i])
		}
	}

	// Recovered
	death = false
	recovered = true
	expected = []string{"ID", "Address", "Date", "Recovered"}
	result = writeHeader(death, recovered)
	if len(expected) != len(result) {
		t.Fatalf("Test Failed: length unequal (%d != %d)",
			len(expected), len(result))
	}
	for i := range expected {
		if expected[i] != result[i] {
			t.Fatalf("Test Failed: expected %s, got != %s",
				expected[i], result[i])
		}
	}
}

func TestWriteAddress(t *testing.T) {
	ts := TimeSeries{}

	// Test no address
	expected := ""
	result := writeAddress(ts)
	if result != expected {
		t.Fatalf("Test failed: expected %s, received %s", expected, result)
	}

	// Test Admin2
	ts.Admin2 = "Admin2"
	ts.Address1 = ""
	ts.Address2 = ""
	expected = "Admin2, "
	result = writeAddress(ts)
	if result != expected {
		t.Fatalf("Test failed: expected %s, received %s", expected, result)
	}

	// Test Address1
	ts.Admin2 = ""
	ts.Address1 = "Address1"
	ts.Address2 = ""
	expected = "Address1, "
	result = writeAddress(ts)
	if result != expected {
		t.Fatalf("Test failed: expected %s, received %s", expected, result)
	}

	// Test Address2
	ts.Admin2 = ""
	ts.Address1 = ""
	ts.Address2 = "Address2"
	expected = "Address2"
	result = writeAddress(ts)
	if result != expected {
		t.Fatalf("Test failed: expected %s, received %s", expected, result)
	}

	// Test All
	ts.Admin2 = "Admin2"
	ts.Address1 = "Address1"
	ts.Address2 = "Address2"
	expected = "Admin2, Address1, Address2"
	result = writeAddress(ts)
	if result != expected {
		t.Fatalf("Test failed: expected %s, received %s", expected, result)
	}
}

func TestWriteRow(t *testing.T) {
	ts := TimeSeries{
		Confirmed: map[time.Time]int{},
		Death:     map[time.Time]int{},
		Recovered: map[time.Time]int{},
	}
	date, err := time.Parse("2006/1/2", "2021/10/31")
	if err != nil {
		t.Errorf("Error during creating Date: %v", err)
	}
	ts.Confirmed[date] = 10
	ts.Death[date] = 20
	ts.Recovered[date] = 30

	death, recovered := false, false

	// Both false
	expected := []string{"10"}
	result := writeRow(ts, date, death, recovered)
	if len(expected) != len(result) || expected[0] != result[0] {
		t.Fatalf("Test failed: expected %s, received %s", expected[0], result[0])
	}

	// Death
	death = true
	recovered = false
	expected = []string{"20"}
	result = writeRow(ts, date, death, recovered)
	if len(expected) != len(result) || expected[0] != result[0] {
		t.Fatalf("Test failed: expected %s, received %s", expected[0], result[0])
	}

	// Recovered
	death = false
	recovered = true
	expected = []string{"30"}
	result = writeRow(ts, date, death, recovered)
	if len(expected) != len(result) || expected[0] != result[0] {
		t.Fatalf("Test failed: expected %s, received %s", expected[0], result[0])
	}
}

func TestMakeQuery(t *testing.T) {
	// Test no params
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)

	query, dates, death, recovered, status := makeQuery(r.URL.Query())

	query = strings.TrimSpace(query)
	expectedQuery := strings.TrimSpace(`
		SELECT DISTINCT TimeSeries.ID, Admin2, Address1, Address2
		FROM TimeSeries JOIN TimeSeriesConfirmed ON
		TimeSeries.ID = TimeSeriesConfirmed.ID
		JOIN TimeSeriesDeath ON TimeSeries.ID = TimeSeriesDeath.ID
		JOIN TimeSeriesRecovered ON TimeSeries.ID = TimeSeriesRecovered.ID
	`)
	if expectedQuery != query {
		t.Fatalf("Test failed: expected %s, got %s", expectedQuery, query)
	}

	expectedDates := ""
	if dates != expectedDates {
		t.Fatalf("Test failed: expected %s, got %s", expectedDates, dates)
	}

	if death || recovered {
		t.Fatalf("Test failed: expected both false, got %v and %v", death, recovered)
	}

	expectedStatus := 0
	if status != expectedStatus {
		t.Fatalf("Test failed: expected %d, got %d", expectedStatus, status)
	}

	// Test invalid params
	r = httptest.NewRequest("GET", "http://example.com/foo?asd=asd", nil)
	_, _, _, _, status = makeQuery(r.URL.Query())
	expectedStatus = 400
	if status != expectedStatus {
		t.Fatalf("Test failed: expected %d, got %d", expectedStatus, status)
	}

	// Test both death and recovered
	r = httptest.NewRequest("GET", "http://example.com/foo?death&recovered", nil)
	_, _, _, _, status = makeQuery(r.URL.Query())
	expectedStatus = 400
	if status != expectedStatus {
		t.Fatalf("Test failed: expected %d, got %d", expectedStatus, status)
	}

	// Test death
	r = httptest.NewRequest("GET", "http://example.com/foo?Death", nil)

	_, _, death, recovered, _ = makeQuery(r.URL.Query())
	expectedDeath, expectedRecovered := true, false
	if !death || recovered {
		t.Fatalf("Test failed: expected %v and %v, got %v and %v",
			expectedDeath, expectedRecovered, death, recovered)
	}

	// Test state and country params
	r = httptest.NewRequest(
		"GET",
		"http://example.com/foo?country=canada,us&state=ontario,ohio",
		nil)

	query, _, _, _, _ = makeQuery(r.URL.Query())
	query = strings.TrimSpace(query)

	lines := strings.Split(query, "\n")
	lastline := strings.TrimSpace(lines[len(lines)-1])

	checker := "address2='canada' OR address2='us'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "address1='ontario' OR address1='ohio'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	// Test admin2, province, and region params
	r = httptest.NewRequest(
		"GET",
		"http://example.com/foo?region=foo&province=bar&admin2=uwu",
		nil)

	query, _, _, _, _ = makeQuery(r.URL.Query())
	query = strings.TrimSpace(query)

	lines = strings.Split(query, "\n")
	lastline = strings.TrimSpace(lines[len(lines)-1])

	checker = "admin2='uwu'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "address1='bar'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "address2='foo'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	// Test date params
	r = httptest.NewRequest(
		"GET",
		"http://example.com/foo?date=1/2/30,4/5/60",
		nil)

	_, dates, _, _, _ = makeQuery(r.URL.Query())

	checker = "date=\"2030/1/2\""
	if !strings.Contains(dates, checker) {
		t.Fatalf("Test failed: expect %s, does not have %s", dates, checker)
	}

	checker = "date=\"2060/4/5\""
	if !strings.Contains(dates, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	// Test from & to
	r = httptest.NewRequest(
		"GET",
		"http://example.com/foo?from=1/2/30,4/5/60&to=1/2/30,4/5/60",
		nil)

	_, dates, _, _, _ = makeQuery(r.URL.Query())

	checker = "date>=\"2030/1/2\""
	if !strings.Contains(dates, checker) {
		t.Fatalf("Test failed: expect %s, does not have %s", dates, checker)
	}

	checker = "date>=\"2060/4/5\""
	if !strings.Contains(dates, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "date<=\"2030/1/2\""
	if !strings.Contains(dates, checker) {
		t.Fatalf("Test failed: expect %s, does not have %s", dates, checker)
	}

	checker = "date<=\"2060/4/5\""
	if !strings.Contains(dates, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}
}

func TestListDefault(t *testing.T) {
	db.InitDb("testing")
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	tsArr := []TimeSeries{}
	err := json.Unmarshal(body, &tsArr)
	if err != nil {
		t.Errorf("Error during converting JSON: %v", err)
	}

	// Assuming there is already data in db
	if len(tsArr) == 0 {
		t.Fatalf("Test failed: empty response")
	}

	for _, ts := range tsArr {
		if len(ts.Confirmed) == 0 {
			t.Fatalf("Test failed: empty response")
		}
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Test failed: expected code 200, got %d", resp.StatusCode)
	}

	expected := "application/json"
	if result := resp.Header.Get("Content-Type"); result != expected {
		t.Fatalf("Test failed: expected %s, got %s", expected, result)
	}
}

func TestListWithParams(t *testing.T) {
	db.InitDb("testing")
	r := httptest.NewRequest("GET", "http://example.com/foo?country=us,canada&from=1/1/20", nil)
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	tsArr := []TimeSeries{}
	err := json.Unmarshal(body, &tsArr)
	if err != nil {
		t.Errorf("Error during converting JSON: %v", err)
	}

	expect1 := "US"
	expect2 := "Canada"
	date1, err := utils.ParseDate("11/1/21")
	if err != nil {
		t.Errorf("Error during parsing date: %v", err)
	}
	date2, err := utils.ParseDate("10/31/21")
	if err != nil {
		t.Errorf("Error during parsing date: %v", err)
	}
	val1 := 420
	val2 := 343

	for _, ts := range tsArr {
		// the US
		if ts.Address2 == expect1 {
			if _, ok := ts.Confirmed[date1]; !ok {
				t.Fatalf("Test failed: value not updated %v %v", date1, ts.Confirmed)
			}
			if ts.Confirmed[date1] != val1 {
				t.Fatalf("Test failed: expect %d, got %d", val1, ts.Confirmed[date1])
			}
		} else if ts.Address2 == expect2 { // Canada
			if _, ok := ts.Confirmed[date2]; !ok {
				t.Fatalf("Test failed: value not updated %v %v", date2, ts.Confirmed)
			}
			if ts.Confirmed[date2] != val2 {
				t.Fatalf("Test failed: expect %d, got %d", val2, ts.Confirmed[date2])
			}
		} else {
			t.Fatalf("Test failed: expected %s or %s, got %s", expect1, expect2, ts.Address2)
		}
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Test failed: expected code 200, got %d", resp.StatusCode)
	}
}

func TestListBadRequests(t *testing.T) {
	db.InitDb("testing")
	r := httptest.NewRequest("GET", "http://example.com/foo?asdfjk", nil)
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedCode := 400
	if resp.StatusCode != expectedCode {
		t.Fatalf("Test failed: expected %d, got %d", expectedCode, resp.StatusCode)
	}
	expected := "Error status 400"
	if string(body) != expected {
		t.Fatalf("Test failed: expected %s, got %s", expected, string(body))
	}
}

func TestListCSVRequests(t *testing.T) {
	db.InitDb("testing")
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)
	r.Header.Set("Accept", "text/csv")
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	line := strings.Split(string(body), "\n")
	expected := "ID,Address,Date,Confirmed"
	if line[0] != expected {
		t.Fatalf("Test failed: expected %s, got %s", expected, line[0])
	}

	expected = "text/csv"
	if result := resp.Header.Get("Content-Type"); result != expected {
		t.Fatalf("Test failed: expected %s, got %s", expected, result)
	}

	// Test death
	r = httptest.NewRequest("GET", "http://example.com/foo?death", nil)
	r.Header.Set("Accept", "text/csv")
	w = httptest.NewRecorder()
	List(w, r)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)
	line = strings.Split(string(body), "\n")
	expected = "ID,Address,Date,Death"
	if line[0] != expected {
		t.Fatalf("Test failed: expected %s, got %s", expected, line[0])
	}
}

func TestGetDateSingleDate(t *testing.T) {
	arr := []string{"1/20/21"}
	beginDate, endDate, beginDateIndex, err := getDates(arr)
	expectedBeginDate := time.Date(2021, 1, 20, 0, 0, 0, 0, time.UTC)
	expectedEndDate := beginDate

	if beginDate != expectedBeginDate {
		t.Fatalf("Test failed: expected value %s, got %s", expectedBeginDate.String(), beginDate.String())
	}

	if endDate != expectedEndDate {
		t.Fatalf("Test failed: expected value %s, got %v", expectedEndDate.String(), endDate.String())
	}

	if beginDateIndex != 0 {
		t.Fatalf("Test failed: expected value 0, got %s", beginDate.String())
	}

	if err != nil {
		t.Errorf("Error occured when getting single date: %v", err)
	}
}

func TestGetDataEmptyData(t *testing.T) {
	var arr []string
	beginDate, endDate, beginDateIndex, err := getDates(arr)
	expectedBeginDate := time.Time{}

	if beginDate != expectedBeginDate {
		t.Fatalf("Test failed: expected value %s, got %s", expectedBeginDate.String(), beginDate.String())
	}
	if endDate != expectedBeginDate {
		t.Fatalf("Test failed: expected value %s, got %s", expectedBeginDate.String(), endDate.String())
	}
	if beginDateIndex != -1 {
		t.Fatalf("Test failed: expected value -1, got %s", beginDate.String())
	}
	if err != nil {
		t.Errorf("Error occured when getting no date: %v", err)
	}
}

func TestGetDatesThreeDates(t *testing.T) {
	arr := []string{"1/20/21", "1/22/21", "1/30/21"}
	beginDate, endDate, beginDateIndex, err := getDates(arr)
	expectedBeginDate := time.Date(2021, 1, 20, 0, 0, 0, 0, time.UTC)
	expectedEndDate := time.Date(2021, 1, 30, 0, 0, 0, 0, time.UTC)

	if beginDate != expectedBeginDate {
		t.Fatalf("Test failed: expected value %s, got %s", expectedBeginDate.String(), beginDate.String())
	}

	if endDate != expectedEndDate {
		t.Fatalf("Test failed: expected value %s, got %v", expectedEndDate.String(), endDate.String())
	}

	if beginDateIndex != 0 {
		t.Fatalf("Test failed: expected value 0, got %s", beginDate.String())
	}

	if err != nil {
		t.Errorf("Error occured when getting single date: %v", err)
	}
}

// test injectTimeSeries
// NOTE: tests assume that database is setup according to create-tables.sql
func TestInjectTimeSeriesExistingTimeSeries(t *testing.T) {
	var ts TimeSeries
	ts.Admin2 = "Autauga"
	ts.Address1 = "Alabama"
	ts.Address2 = "US"
	id, err := injectTimeSeries(0, ts)
	var expectedId int64 = 1

	if expectedId != id {
		t.Fatalf("Test failed: expected id 1, got %d", id)
	}
	if err != nil {
		t.Errorf("Error occured when injecting existing record: %v", err)
	}

	// test existing Address1 and Address2 but New Admin2
	ts.Admin2 = "Madison"
	ts.Address1 = "Ontario"
	ts.Address2 = "Canada"
	id, err = injectTimeSeries(-1, ts)
	if id == 2 {
		t.Fatalf("Test failed: id should not be 2")
	}

	if err != nil {
		t.Errorf("Error occured when injecting existing record: %v", err)
	}

	// test existing Admin2 and Address2 but empty Address1
	ts.Admin2 = "Autauga"
	ts.Address1 = ""
	ts.Address2 = "US"
	id, err = injectTimeSeries(0, ts)
	if id == 1 {
		t.Fatalf("Test failed: id should not be 1")
	}

	if err != nil {
		t.Errorf("Error occured when injecting existing record: %v", err)
	}

	// test existing Address2 but empty Admin2 and Address1
	ts.Admin2 = ""
	ts.Address1 = ""
	ts.Address2 = "US"
	id, err = injectTimeSeries(-1, ts)
	if id == 1 {
		t.Fatalf("Test failed: id should not be 1")
	}

	if err != nil {
		t.Errorf("Error occured when injecting existing record: %v", err)
	}
}

func TestCreate(t *testing.T) {
	db.InitDb("testing")
	// rows, err := db.Db.Query(`SELECT * FROM TimeSeries"`)
	// if err != nil {
	// 	t.Errorf("Error ocurred when querying TimeSeries: %v", err)
	// }

	// Creating the body of the request
	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)

	// Fill in the 2d-array
	csvArr := [][]string{}
	header := []string{
		"Admin2", "Province/State", "Country/Region", "1/31/20",
	}
	csvArr = append(csvArr, header)
	row := []string{
		"", "Ontario", "Canada", "42069",
	}
	csvArr = append(csvArr, row)

	// Write to buffer
	if err := writer.WriteAll(csvArr); err != nil {
		t.Errorf("Error while converting csvArr to bytes")
	}

	w := httptest.NewRecorder()

	// curl --data-binary @timeSeries_test2.csv
	r := httptest.NewRequest("POST", "http://example.com/foo", b)
	r.Header.Set("FileType", "Confirmed")

	// Goal: call Create()
	Create(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedCode := 200
	if resp.StatusCode != expectedCode {
		t.Fatalf("Test failed: expected code %d, got %d", expectedCode, resp.StatusCode)
	}

	expectedBody := "Successfully create/update data to the system"
	if string(body) != expectedBody {
		t.Fatalf("Test failed: expected body %s, got %s", expectedBody, string(body))
	}

	// TODO: validate the data is actually updated
	// List(w, r)

	// tsArr := []TimeSeries{}
	// err := json.Unmarshal(body, &tsArr)
	// if err != nil {
	// 	t.Errorf("Error during converting JSON: %v", err)
	// }
}

func TestCreateBadHeader(t *testing.T) {
	db.InitDb("testing")
	//test no header

	//Create body
	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)

	//fill 2d array
	csvArr := [][]string{}
	header := []string{
		"Admin2", "Province/State", "Country/Region", "1/31/20",
	}
	csvArr = append(csvArr, header)
	row := []string{
		"", "Ontario", "Canada", "42069",
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
	r.Header.Set("FileType", "Active")

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

func TestCreateInvalidDateFormat(t *testing.T) {
	db.InitDb("testing")
	// rows, err := db.Db.Query(`SELECT * FROM TimeSeries"`)
	// if err != nil {
	// 	t.Errorf("Error ocurred when querying TimeSeries: %v", err)
	// }

	// Creating the body of the request
	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)

	// Fill in the 2d-array
	csvArr := [][]string{}
	header := []string{
		"Admin2", "Province/State", "Country/Region", "13/1/20",
	}
	csvArr = append(csvArr, header)
	row := []string{
		"", "Ontario", "Canada", "42069", "1337",
	}
	csvArr = append(csvArr, row)

	// Write to buffer
	if err := writer.WriteAll(csvArr); err != nil {
		t.Errorf("Error while converting csvArr to bytes")
	}

	w := httptest.NewRecorder()

	// curl --data-binary @timeSeries_test2.csv
	r := httptest.NewRequest("POST", "http://example.com/foo", b)
	r.Header.Set("FileType", "Confirmed")

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
}

func TestCreateDuplicatedDatesOneFile(t *testing.T) {
	// Test 2 same dates in one file

	db.InitDb("testing")
	// rows, err := db.Db.Query(`SELECT * FROM TimeSeries"`)
	// if err != nil {
	// 	t.Errorf("Error ocurred when querying TimeSeries: %v", err)
	// }

	// Creating the body of the request
	b := new(bytes.Buffer)
	writer := csv.NewWriter(b)

	// Fill in the 2d-array
	csvArr := [][]string{}
	header := []string{
		"Admin2", "Province/State", "Country/Region", "1/31/20", "1/31/20",
	}
	csvArr = append(csvArr, header)
	row := []string{
		"", "Ontario", "Canada", "42069", "1337",
	}
	csvArr = append(csvArr, row)

	// Write to buffer
	if err := writer.WriteAll(csvArr); err != nil {
		t.Errorf("Error while converting csvArr to bytes")
	}

	w := httptest.NewRecorder()

	// curl --data-binary @timeSeries_test2.csv
	r := httptest.NewRequest("POST", "http://example.com/foo", b)
	r.Header.Set("FileType", "Confirmed")

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
}
