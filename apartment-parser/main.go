package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Offer struct {
	title    string
	price    string
	location string
	time     string
	url      string
}

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

func readHtmlFromFile(filename string) (string, error) {
	bs, err := ioutil.ReadFile(filename)

	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func parseOffer(text string) Offer {
	tkn := html.NewTokenizer(strings.NewReader(text))

	offer := Offer{}
	var isTitle bool
	var isPrice bool
	var isTimeAndLoc bool
	for {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			// End of the document, we're done
			return offer
		case html.StartTagToken:
			t := tkn.Token()
			switch t.Data {
			case "h6":
				isTitle = true
			case "p":
				for _, attr := range t.Attr {
					if attr.Key == "class" {
						switch attr.Val {
						case "css-veheph er34gjf0":
							isTimeAndLoc = true
						case "css-10b0gli er34gjf0":
							isPrice = true
						}
					}
				}
			case "a":
				for _, attr := range t.Attr {
					if attr.Key == "href" {
						url := attr.Val
						if url[0] == '/' {
							url = "https://www.olx.pl" + url
						}
						offer.url = url
					}
				}
			case "div":
				// Indicates that the offer is a featured one
				for _, attr := range t.Attr {
					if attr.Key == "data-testid" && attr.Val == "adCard-featured" {
						return Offer{}
					}
				}
			}

		case html.TextToken:
			t := tkn.Token()
			if isTitle {
				offer.title = t.Data
				isTitle = false
			} else if isPrice {
				offer.price = t.Data
				isPrice = false
			} else if isTimeAndLoc {
				offer.location = t.Data
				for i := 0; i < 4; i++ {
					tkn.Next()
				}
				offer.time = tkn.Token().Data
				isTimeAndLoc = false
			}
		}
	}
}

func parseHtml(text string) (data []Offer) {
	tkn := html.NewTokenizer(strings.NewReader(text))

	var offers []Offer

	var isOffer bool
	var offerContent string
	offerSeparator := "css-1sw7q4x"
	depth := 0

	for {
		tt := tkn.Next()

		switch tt {
		case html.ErrorToken:
			// End of the document, we're done
			return offers
		case html.StartTagToken:
			t := tkn.Token()
			if !isOffer {
				if t.Data == "div" {
					for _, attr := range t.Attr {
						if attr.Key == "class" && attr.Val == offerSeparator {
							isOffer = true
							break
						}
					}
				}
			} else {
				if t.Data == "div" {
					depth++
				}
				offerContent += t.String()
			}
		case html.EndTagToken:
			t := tkn.Token()
			if isOffer && t.Data == "div" && depth == 0 {
				isOffer = false
				offer := parseOffer(offerContent)

				// TODO: For some reason, the last div recognized as offer is empty
				// Inspect this later
				if offer.title != "" {
					offers = append(offers, offer)
				}
				offerContent = ""
				depth = 0
			} else if isOffer {
				if t.Data == "div" {
					depth--
				}
				offerContent += t.String()
			}
		default:
			if isOffer {
				offerContent += tkn.Token().String()
			}
		}
	}
}

func main() {
	fetchHTMLPage("https://www.olx.pl/poznan/q-mieszkanie/?search%5Border%5D=created_at:desc&search%5Bfilter_float_price:from%5D=1000&search%5Bfilter_float_price:to%5D=3000", "")

	filename := "index.html"
	text, err := readHtmlFromFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	data := parseHtml(text)
	for _, offer := range data {
		fmt.Println("--------------------")
		fmt.Println("Title: ", offer.title)
		fmt.Println("Price: ", offer.price)
		fmt.Println("Location: ", offer.location)
		fmt.Println("Time: ", offer.time)
		fmt.Println("URL: ", offer.url)
	}
}
