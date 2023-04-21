package parser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
func FetchHTMLPage(url_string string) (string, error) {
	resp, err := http.Get(url_string)
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
	var builder strings.Builder
	builder.WriteString("https://www.olx.pl/nieruchomosci/mieszkania/wynajem/")

	// Check if the search term has a location.
	if searchTerm.Location == "" {
		return "", errors.New("No location specified in search term.")
	}
	builder.WriteString(searchTerm.Location)
	builder.WriteString("/q-mieszkanie/?search[order]=created_at:desc")

	if searchTerm.Price_min != 0 {
		fmt.Fprintf(&builder, "&search[filter_float_price:from]=%g", searchTerm.Price_min)
	}

	if searchTerm.Price_max != 0 {
		fmt.Fprintf(&builder, "&search[filter_float_price:to]=%g", searchTerm.Price_max)
	}

	if searchTerm.Size_min != 0 {
		fmt.Fprintf(&builder, "&search[filter_float_m:from]=%g", searchTerm.Size_min)
	}

	if searchTerm.Size_max != 0 {
		fmt.Fprintf(&builder, "&search[filter_float_m:to]=%g", searchTerm.Size_max)
	}

	if len(searchTerm.Bedrooms) > 0 {
		values := make([]string, len(searchTerm.Bedrooms))
		for i, bedroom := range searchTerm.Bedrooms {
			values[i] = "search[filter_enum_rooms][" + strconv.Itoa(i) + "]=" + bedroom
		}
		builder.WriteString("&")
		builder.WriteString(strings.Join(values, "&"))
	}

	return builder.String(), nil
}

func GetSearchInfo(url_string string) (string, error) {
	// If URL starts with "olx.pl"
	if strings.HasPrefix(url_string, "https://www.olx.pl") {
		// Split the URL into parts
		parts := strings.Split(url_string, "/")

		// Get the city
		text := strings.ToUpper(parts[6][:1]) + parts[6][1:]

		u, err := url.Parse(url_string)

		if err != nil {
			return "", err
		}

		// Get the price
		text += " (" + u.Query().Get("search[filter_float_price:from]") + "-" + u.Query().Get("search[filter_float_price:to]") + ")"

		return text, nil
	} else {
		return "", errors.New("Invalid URL")
	}
}
