package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
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
	if year < 0 || year > 9999 {
		return time.Time{}, errors.New("Year Syntax Error")
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
	if day < 0 || day > 31 {
		return time.Time{}, errors.New("Day Syntax Error")
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

/**
Helper function for Create().
Iterate from index until end of arr, checking for duplicate string
Return true if duplicate exists. Otherwise, return false
*/
func HasDupe(index int, arr []string) bool {
	// Check for duplicate dates
	allKeys := make(map[string]bool)
	//fmt.Println(len(result))
	for i := index; i < len(arr); i++ {
		//fmt.Println(allKeys)
		if value := allKeys[arr[i]]; !value {
			allKeys[arr[i]] = true
		} else { // duplicate found
			return true
		}
	}

	return false
}

func HandleErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	response := fmt.Sprintf("Error status %d", code)
	if _, err := w.Write([]byte(response)); err != nil {
		log.Fatal(err)
	}
	log.Println("Error: ", err)
}
