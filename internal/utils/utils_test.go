package utils

import (
	"testing"
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

	var res string
	var ok bool
	for _, v := range params {
		res, ok = ParamValidate(v)
		if !ok {
			if res != "" {
				t.Fatalf("Test failed")
			}
		} else {
			// address 1
			if (v == "province" || v == "state") &&
				res != "address1" {
				t.Fatalf("Test failed: expect %s, got %s", "address1", res)
			} else if (v == "country" || v == "region") &&
				res != "address2" { // address2
				t.Fatalf("Test failed: expect %s, got %s", "address2", res)
			} else {
				if res != v {
					t.Fatalf("Test failed: expect %s, got %s", v, res)
				}
			}
		}
	}
}
