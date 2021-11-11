package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func ParamValidate(param string) (string, bool) {
	param = strings.ToLower(param)
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
	temp := strings.Split(date, "/") // i.e. "mm/dd/yy" -> [ "mm", "dd", "yy" ]
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
	if month < 0 || month > 12 {
		return time.Time{}, errors.New("Month Syntax Error")
	}
	day, err := strconv.Atoi(temp[1])
	if err != nil {
		return time.Time{}, err
	}
<<<<<<< HEAD
	if day < 0 || day > 31 {
		return time.Time{}, errors.New("Day Syntax Error")
	}
=======
>>>>>>> 9aecf57 (Implemented timeSeries.go to account duplicate addresses. Replaced time.Local with time.UTC in utils.go)
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}
