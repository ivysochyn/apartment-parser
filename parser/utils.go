package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Struct to hold the search term.
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
	req, err := http.NewRequest("GET", url_string, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("TZ", "Europe/Warsaw")

	client := &http.Client{}
	resp, err := client.Do(req);
	if err != nil {
		return "", err
	} else if (resp.StatusCode != 200) {
		return "", errors.New("Status code: " + strconv.Itoa(resp.StatusCode))
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

func GetSearchShortInfo(url_string string) (string, error) {
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

func GetSearchFullInfo(url_string string) (string, error) {
	// If URL starts with "olx.pl"
	if strings.HasPrefix(url_string, "https://www.olx.pl") {
		// Split the URL into parts
		parts := strings.Split(url_string, "/")

		text := "🏠 Full info of the search:\n\n"
		// Get the city
		text += "📍 " + strings.ToUpper(parts[6][:1]) + parts[6][1:] + "\n"

		u, err := url.Parse(url_string)

		if err != nil {
			return "", err
		}

		q := u.Query()

		if price_from, ok := q["search[filter_float_price:from]"]; ok {
			if price_to, ok := q["search[filter_float_price:to]"]; ok {
				text += "💰 Price: " + price_from[0] + "-" + price_to[0] + " zł\n"
			}
		}

		if size_from, ok := q["search[filter_float_m:from]"]; ok {
			if size_to, ok := q["search[filter_float_m:to]"]; ok {
				text += "📐 Area: " + size_from[0] + "-" + size_to[0] + " m²\n"
			}
		}

		bedrooms := make([]string, 0)
		floors := make([]string, 0)

		for key, value := range q {
			if strings.HasPrefix(key, "search[filter_enum_floor_select]") {
				floors = append(floors, value[0])
			} else if strings.HasPrefix(key, "search[filter_enum_rooms]") {
				bedrooms = append(bedrooms, value[0])
			}
		}

		if len(bedrooms) > 0 {
			text += "🛏 Bedrooms:\n    - "
			for k, bedroom := range bedrooms {
				if k != len(bedrooms)-1 {
					text += strings.ToUpper(bedroom[:1]) + bedroom[1:] + ", "
				} else {
					text += strings.ToUpper(bedroom[:1]) + bedroom[1:] + "\n"
				}
			}
		}

		if len(floors) > 0 {
			text += "🏢 Floors:\n    - "
			for k, floor := range floors {
				if k != len(floors)-1 {
					text += floorEncodings[floor] + ", "
				} else {
					text += floorEncodings[floor] + "\n"
				}
			}
		}

		// Print the url as an hyperlink
		text += "\n🔗 <a href=\"" + url_string + "\">Link to the search</a>"

		return text, nil
	} else {
		return "", errors.New("Invalid URL")
	}
}

func DownloadImage(image_url string) ([]byte, error) {
	resp, err := http.Get(image_url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}

func parseOtodomImages(json_string string) ([]string, error) {
	var images []string

	// Create a new JSON decoder
	decoder := json.NewDecoder(strings.NewReader(json_string))

	// Decode the JSON
	var data map[string]interface{}
	err := decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	// Images are stored under "props" -> "pageProps" -> "ad" -> "images"
	images_data := data["props"].(map[string]interface{})["pageProps"].(map[string]interface{})["ad"].(map[string]interface{})["images"].([]interface{})
	for _, image := range images_data {
		images = append(images, image.(map[string]interface{})["large"].(string))
	}

	return images, nil
}
