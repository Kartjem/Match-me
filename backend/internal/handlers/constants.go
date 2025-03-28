package handlers

var validLocations = []string{"USA", "Canada", "UK", "Mexico", "Germany", "Estonia"}

var cityOptionsMap = map[string][]string{
	"USA":     {"New York", "Los Angeles", "Chicago", "Houston", "Dallas"},
	"Canada":  {"Toronto", "Vancouver", "Montreal", "Calgary"},
	"UK":      {"London", "Manchester", "Liverpool", "Birmingham"},
	"Mexico":  {"Mexico City", "Guadalajara", "Monterrey"},
	"Germany": {"Berlin", "Hamburg", "Munich", "Frankfurt"},
	"Estonia": {"Tallinn", "Tartu", "Narva", "Pärnu"},
}
