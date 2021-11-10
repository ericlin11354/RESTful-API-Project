package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func ParamValidate(param string) (string, bool) {
	validator := map[string]string{
		"id":        "id",
		"admin2":    "admin2",
		"province":  "address1",
		"state":     "address1",
		"country":   "address2",
		"region":    "address2",
		"date":      "date",
		"from":      "from",
		"to":        "to",
		"death":     "death",
		"recovered": "recovered",
	}
	result, ok := validator[param]
	return result, ok
}

/**
Helper function for Create().
Takes Date string and returns type Date as type time.Time

*/
func ParseDate(date string) (time.Time, error) {
	temp := strings.Split(date, "/") // i.e. "1/23/20" -> [ "1", "23", "20" ]
	if len(temp) != 3 {
		return time.Time{}, errors.New("Syntax Error")
	}
	year, err := strconv.Atoi("20" + temp[2])
	if err != nil {
		return time.Time{}, err
	}
	month, err := strconv.Atoi(temp[0])
	if err != nil {
		return time.Time{}, err
	}
	day, err := strconv.Atoi(temp[1])
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local), nil
}
