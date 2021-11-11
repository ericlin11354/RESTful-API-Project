package dailyReports

import (
	"database/sql"
	"net/http/httptest"
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

func TestList(t *testing.T) {
	db.InitDb()
}
