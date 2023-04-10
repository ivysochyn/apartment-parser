package main

import (
	"io/ioutil"
	"net/http"
	"os"
)

// fetchHTMLPage fetches the HTML page at the given URL and stores it in the given path.
// If the path is empty, the page is stored in the current directory.
func fetchHTMLPage(url string, path string) {
	if path == "" {
		path = "./index.html"
	}

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Create the file if it doesn't exist
	// overwrites the file if it already exists
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	// Write the body to file
	_, err = file.Write(body)
	if err != nil {
		panic(err)
	}
}

func main() {
	fetchHTMLPage("https://www.google.com", "")
}
