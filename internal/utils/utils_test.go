package utils

import (
	"testing"
	"time"
)

func TestParamValidate(t *testing.T) {
	params := []string{
		"asd", // bad param
		"id",
		"admin2",
		"province",
		"state",
		"country",
		"region",
		"date",
		"from",
		"to",
		"death",
		"recovered",
	}

	var res, expect string
	var ok bool
	for _, v := range params {
		res, ok = ParamValidate(v)
		if !ok {
			if res != "" {
				t.Fatalf("Test failed: does not return empty string upon key error")
			}
		} else {
			switch res {
			case "address1":
				expect = "address1"
				if v != "province" && v != "state" {
					t.Fatalf("Test failed: expect %s, got %s", expect, v)
				}
			case "address2":
				expect = "address2"
				if v != "country" && v != "region" {
					t.Fatalf("Test failed: expect %s, got %s", expect, v)
				}
			default:
				if res != v {
					t.Fatalf("Test failed: validating %s failed", v)
				}
			}
		}
	}
}

func TestParseDataValidInput(t *testing.T) {
	// 1 Jan 2020
	input := "1/31/20"
	date, err := ParseDate(input)
	if err != nil {
		t.Errorf("Error while parsing date: %v", err)
	}

	expect := time.Date(2020, time.Month(1), 31, 0, 0, 0, 0, time.UTC)

	if expect != date {
		t.Fatalf("Test failed: Expected %s, got %s", expect.String(), date.String())
	}
}

func TestParseDataInvalidInput(t *testing.T) {
	// Month over 12
	input := "31/1/20"
	date, err := ParseDate(input)
	if err == nil {
		t.Fatalf("Test failed: Error not raised")
	}
	expect := time.Time{}
	if date != expect {
		t.Fatalf("Test failed: Date not default:%s", date.String())
	}

	// Date over 99
	input = "1/99/20"
	date, err = ParseDate(input)
	if err == nil {
		t.Fatalf("Test failed: Error not raised")
	}
	expect = time.Time{}
	if date != expect {
		t.Fatalf("Test failed: Date not default:%s", date.String())
	}

	// Year over 999
	input = "1/1/999"
	date, err = ParseDate(input)
	if err == nil {
		t.Fatalf("Test failed: Error not raised")
	}
	expect = time.Time{}
	if date != expect {
		t.Fatalf("Test failed: Date not default:%s", date.String())
	}

	// Length after split not 3
	input = "abc"
	date, err = ParseDate(input)
	if err == nil {
		t.Fatalf("Test failed: Error not raised")
	}
	expect = time.Time{}
	if date != expect {
		t.Fatalf("Test failed: Date not default:%s", date.String())
	}

	// Bad inputs
	input = "a/1/20"
	date, err = ParseDate(input)
	if err == nil {
		t.Fatalf("Test failed: Error not raised")
	}
	expect = time.Time{}
	if date != expect {
		t.Fatalf("Test failed: Date not default:%s", date.String())
	}

	input = "1/b/20"
	date, err = ParseDate(input)
	if err == nil {
		t.Fatalf("Test failed: Error not raised")
	}
	expect = time.Time{}
	if date != expect {
		t.Fatalf("Test failed: Date not default:%s", date.String())
	}

	input = "1/1/c"
	date, err = ParseDate(input)
	if err == nil {
		t.Fatalf("Test failed: Error not raised")
	}
	expect = time.Time{}
	if date != expect {
		t.Fatalf("Test failed: Date not default:%s", date.String())
	}
}

func TestHasDupe(t *testing.T) {
	// No dupes
	i := 0
	arr := []string{"a", "b", "c", "d"}
	if HasDupe(i, arr) {
		t.Fatalf("Test failed: there is no dupe")
	}

	// Has dupes
	arr = []string{"a", "b", "c", "a"}
	if !HasDupe(i, arr) {
		t.Fatalf("Test failed: there is a dupe")
	}
}
