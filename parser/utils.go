package parser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Stuct to hold the search term.
// The search term is the parameters that are used to search for a property.
// For example, a search term could be:
//   - location: "Stockholm"
//   - price_min: 1000000
//   - price_max: 2000000
//   - bedrooms: 3
//   - size_min: 100
//   - size_max: 200
type SearchTerm struct {
	Location  string
	Price_min float64
	Price_max float64
	Bedrooms  []string
	Size_min  float64
	Size_max  float64
}

// FetchHTMLPage fetches the HTML page from the given URL
// and returns the HTML page as a string.
// If an error occurs, it returns an empty string and the error.
//
// Example:
//
//	html, err := FetchHTMLPage("https://www.google.com")
//	if err != nil {
//	    // handle error
//	}
func FetchHTMLPage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// CreateUrl function is used to generate the URL for a given search term.
// Accepts a SearchTerm struct as input which contains the search term.
// Requires the location to be specified in the search term.
// Returns the URL as a string and an error.
//
// Example:
//
//		url, err := CreateUrl(SearchTerm{
//		    Location: "Poznan",
//		    Price_min: 1000,
//		    Price_max: 2000,
//		    Bedrooms: []string{"2", "3"},
//	     })
//		if err != nil {
//		    // handle error
//		}
func CreateUrl(searchTerm SearchTerm) (string, error) {
	url := "https://www.olx.pl/nieruchomosci/mieszkania/wynajem/"

	// Check if the search term has a location.
	if searchTerm.Location != "" {
		url += searchTerm.Location + "/q-mieszkanie/"
	} else {
		return "", errors.New("No location specified in search term.")
	}

	if searchTerm.Price_min != 0 {
		url += "?search[filter_float_price:from]=" + strconv.FormatFloat(searchTerm.Price_min, 'f', 0, 64)
	}

	if searchTerm.Price_max != 0 {
		if url[len(url)-1] != '?' {
			url += "&"
		} else {
			url += "?"
		}
		url += "search[filter_float_price:to]=" + strconv.FormatFloat(searchTerm.Price_max, 'f', 0, 64)
	}

	if searchTerm.Size_min != 0 {
		if url[len(url)-1] != '?' {
			url += "&"
		} else {
			url += "?"
		}
		url += "search[filter_float_m:from]=" + strconv.FormatFloat(searchTerm.Size_min, 'f', 0, 64)
	}

	if searchTerm.Size_max != 0 {
		if url[len(url)-1] != '?' {
			url += "&"
		} else {
			url += "?"
		}
		url += "search[filter_float_m:to]=" + strconv.FormatFloat(searchTerm.Size_max, 'f', 0, 64)
	}

	if searchTerm.Bedrooms != nil {
		for i, bedroom := range searchTerm.Bedrooms {
			if url[len(url)-1] != '?' {
				url += "&"
			} else {
				url += "?"
			}
			url += "search[filter_enum_rooms][" + strconv.Itoa(i) + "]=" + bedroom
		}
	}
	return url, nil
}

func GetSearchInfo(url string) (string, error) {
	// If URL starts with "olx.pl"
	if strings.HasPrefix(url, "https://www.olx.pl") {
		// Split the URL into parts
		parts := strings.Split(url, "/")
		// for i, part := range parts {
		//     fmt.Println(i, part)
		// }

		city := strings.ToUpper(parts[6][:1]) + parts[6][1:]

		// Extract the price from '?search[filter_float_price:from]=1000&search[filter_float_price:to]=2000'
		price := strings.Split(parts[8], "=")
		price_min := price[1]
		price_min = strings.Split(price_min, "&")[0]
		price_max := price[2]
		fmt.Println(price_min, price_max)

		city = city + " (" + price_min + "-" + price_max + " zł)"
		return city, nil
	} else {
		return "", errors.New("Invalid URL")
	}
}
