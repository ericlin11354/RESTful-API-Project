package timeSeries

import (
	"database/sql"
	"log"
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
	expected = []string{"10", "20"}
	result = writeRow(ts, date, death, recovered)
	if len(expected) == len(result) {
		for i := range result {
			if expected[i] != result[i] {
				t.Fatalf("Test failed: expected %s, received %s", expected[i], result[i])
			}
		}
	} else {
		t.Fatalf("Test failed: expected's array length %d, received %d",
			len(expected), len(result))
	}

	// Recovered
	death = false
	recovered = true
	expected = []string{"10", "30"}
	result = writeRow(ts, date, death, recovered)
	if len(expected) == len(result) {
		for i := range result {
			if expected[i] != result[i] {
				t.Fatalf("Test failed: expected %s, received %s", expected[i], result[i])
			}
		}
	} else {
		t.Fatalf("Test failed: expected's array length %d, received %d",
			len(expected), len(result))
	}

	// Both
	death = true
	recovered = true
	expected = []string{"10", "20", "30"}
	result = writeRow(ts, date, death, recovered)
	if len(expected) == len(result) {
		for i := range result {
			if expected[i] != result[i] {
				t.Fatalf("Test failed: expected %s, received %s", expected[i], result[i])
			}
		}
	} else {
		t.Fatalf("Test failed: expected's array length %d, received %d",
			len(expected), len(result))
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
	expected = []string{"ID", "Address", "Date", "Confirmed", "Death"}
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
	expected = []string{"ID", "Address", "Date", "Confirmed", "Recovered"}
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

	// Both
	death = true
	recovered = true
	expected = []string{"ID", "Address", "Date", "Confirmed", "Death", "Recovered"}
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

// Testing List
func TestMakeQuery(t *testing.T) {

}
