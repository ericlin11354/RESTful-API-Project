package timeSeries

import (
	"database/sql"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/joho/godotenv"

	db "gitlab.com/csc301-assignments/a2/internal/db"
)

func ConnectToDb() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// Initialize Chi Router
	r := chi.NewRouter()

	// Initizalize database and migrate
	db.InitDb()

	// Creating endpoints
	r.Mount("/time_series", Routes())
	r.Mount("/daily_reports", Routes())
}

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

// resp := w.Result()
// body, _ := io.ReadAll(resp.Body)

// fmt.Println(resp.StatusCode)
// fmt.Println(resp.Header.Get("Content-Type"))
// fmt.Println(string(body))

// Output:
// 200
// text/html; charset=utf-8
// <html><body>Hello World!</body></html>
func TestMakeQuery(t *testing.T) {
	// Test no params
	r := httptest.NewRequest("GET", "http://example.com/foo", nil)

	query, death, recovered, status := makeQuery(r.URL.Query())

	query = strings.TrimSpace(query)
	expected_query := strings.TrimSpace(`
		SELECT DISTINCT TimeSeries.ID, Admin2, Address1, Address2
		FROM TimeSeries JOIN TimeSeriesConfirmed ON
		TimeSeries.ID = TimeSeriesConfirmed.ID
		JOIN TimeSeriesDeath ON TimeSeries.ID = TimeSeriesDeath.ID
		JOIN TimeSeriesRecovered ON TimeSeries.ID = TimeSeriesRecovered.ID
	`)
	if expected_query != query {
		t.Fatalf("Test failed: expected %s, got %s", expected_query, query)
	}

	if death || recovered {
		t.Fatalf("Test failed: expected both false, got %v and %v", death, recovered)
	}

	if status != 0 {
		t.Fatalf("Test failed: expected 0, got %d", status)
	}

	// Test invalid params
	r = httptest.NewRequest("GET", "http://example.com/foo?asd=asd", nil)

	query, death, recovered, status = makeQuery(r.URL.Query())
	expected_query = ""
	if expected_query != query {
		t.Fatalf("Test failed: expected %s, got %s", expected_query, query)
	}

	if death || recovered {
		t.Fatalf("Test failed: expected both false, got %v and %v", death, recovered)
	}

	if status != 400 {
		t.Fatalf("Test failed: expected 400, got %d", status)
	}

	// Test both death and recovered
	r = httptest.NewRequest("GET", "http://example.com/foo?death&recovered", nil)

	query, death, recovered, status = makeQuery(r.URL.Query())
	expected_query = ""
	if expected_query != query {
		t.Fatalf("Test failed: expected %s, got %s", expected_query, query)
	}

	if death || recovered {
		t.Fatalf("Test failed: expected both false, got %v and %v", death, recovered)
	}

	if status != 400 {
		t.Fatalf("Test failed: expected 400, got %d", status)
	}

	// Test death
	r = httptest.NewRequest("GET", "http://example.com/foo?Death", nil)

	query, death, recovered, status = makeQuery(r.URL.Query())
	query = strings.TrimSpace(query)
	expected_query = strings.TrimSpace(`
		SELECT DISTINCT TimeSeries.ID, Admin2, Address1, Address2
		FROM TimeSeries JOIN TimeSeriesConfirmed ON
		TimeSeries.ID = TimeSeriesConfirmed.ID
		JOIN TimeSeriesDeath ON TimeSeries.ID = TimeSeriesDeath.ID
		JOIN TimeSeriesRecovered ON TimeSeries.ID = TimeSeriesRecovered.ID
	`)
	if expected_query != query {
		t.Fatalf("Test failed: expected %s, got %s", expected_query, query)
	}

	if !death || recovered {
		t.Fatalf("Test failed: expected both true, got %v and %v", death, recovered)
	}

	if status != 0 {
		t.Fatalf("Test failed: expected 0, got %d", status)
	}

	// Test state and country params
	r = httptest.NewRequest(
		"GET",
		"http://example.com/foo?country=canada,us&state=ontario,ohio",
		nil)

	query, death, recovered, status = makeQuery(r.URL.Query())
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

	if death || recovered {
		t.Fatalf("Test failed: expected both false, got %v and %v", death, recovered)
	}

	if status != 0 {
		t.Fatalf("Test failed: expected 0, got %d", status)
	}

	// Test admin2, province, and region params
	r = httptest.NewRequest(
		"GET",
		"http://example.com/foo?region=foo&province=bar&admin2=uwu",
		nil)

	query, death, recovered, status = makeQuery(r.URL.Query())
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

	if death || recovered {
		t.Fatalf("Expected both false, got %v and %v", death, recovered)
	}

	if status != 0 {
		t.Fatalf("Test failed: expected 0, got %d", status)
	}

	// Test date params
	r = httptest.NewRequest(
		"GET",
		"http://example.com/foo?date=1/2/3&from=4/5/6&to=7/8/9",
		nil)

	query, death, recovered, status = makeQuery(r.URL.Query())
	query = strings.TrimSpace(query)

	lines = strings.Split(query, "\n")
	lastline = strings.TrimSpace(lines[len(lines)-1])

	checker = "date='1/2/3'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "date>='4/5/6'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	checker = "date<='7/8/9'"
	if !strings.Contains(lastline, checker) {
		t.Fatalf("Test failed: query does not contain %s", checker)
	}

	if death || recovered {
		t.Fatalf("Expected both false, got %v and %v", death, recovered)
	}

	if status != 0 {
		t.Fatalf("Test failed: expected 0, got %d", status)
	}
}

func TestList(t *testing.T) {
	ConnectToDb()
	r := httptest.NewRequest("GET", "http://example.com/foo?country=us&date=21/11/1&death", nil)
	w := httptest.NewRecorder()
	List(w, r)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	t.Errorf("%s", string(body))

	if resp.StatusCode != 200 {
		t.Fatalf("Test failed: expected code 200, got %d", resp.StatusCode)
	}

}
