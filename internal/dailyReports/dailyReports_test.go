package dailyReports

import (
	"database/sql"
	"testing"
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
