package parser

import (
	"io/ioutil"
	"net/http"
)

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
