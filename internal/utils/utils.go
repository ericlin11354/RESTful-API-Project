package utils

func ParamValidate(param string) (string, bool) {
	validator := map[string]string{
		"id":       "id",
		"admin2":   "admin2",
		"province": "address1",
		"state":    "address1",
		"country":  "address2",
		"region":   "address2",
		"date":     "date",
		"from":     "from",
		"to":       "to",
	}
	result, ok := validator[param]
	return result, ok
}
