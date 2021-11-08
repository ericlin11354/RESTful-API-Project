package dailyReports

type DailyReports struct {
	UID         string `json:"id"`
	Admin2      string `json:"Admin2"`
	Address1    string `json:"Province/State"`
	Address2    string `json:"Country/Region"`
	CombinedKey string `json:"CombinedKey"`
	Confirmed   int    `json:"Confirmed"`
	Deaths      int    `json:"Deaths"`
	Recovered   int    `json:"Recovered"`
	Active      int    `json:"Active"`
}
