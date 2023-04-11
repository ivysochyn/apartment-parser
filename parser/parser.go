// Package parser contains functions for parsing the HTML code of the website.
//
// It is used to extract all the non-featured offers of apartments for rent from a given link.
package parser

import (
	"golang.org/x/net/html"
	"strings"
)

// Offer represents a single offer from the website
// It contains the title, price, location, time and url.
//
// Example:
// 	{
// 		Title: "Mieszkanie 2 pokojowe",
// 		Price: "1 000 zł",
// 		Location: "Warszawa",
// 		Time: "dzisiaj 12:00",
// 		Url: "https://www.olx.pl/oferta/mieszkanie-2-pokojowe-ID6Q2Zr.html"
// 	}
type Offer struct {
	Title    string
	Price    string
	Location string
	Time     string
	Url      string
}


// checkAttr checks if the given attribute is present in the list of attributes
// and if it has the given value.
//
// Returns true if the attribute is present and has the given value.
// Returns false otherwise.
//
// Example:
// 	attrs := []html.Attribute{
// 		{Key: "class", Val: "css-10b0gli er34gjf0"},
// 		{Key: "data-testid", Val: "adCard-featured"},
// 	}
// 	checkAttr(attrs, "class", "css-10b0gli er34gjf0") // returns true
// 	checkAttr(attrs, "data-testid", "adCard-featured") // returns true
// 	checkAttr(attrs, "class", "css-10b0gli er34gjf1") // returns false
func checkAttr(attrs []html.Attribute, key, value string) bool {
    for _, attr := range attrs {
        if attr.Key == key && attr.Val == value {
            return true
        }
    }
    return false
}

// getAttr returns the value of the given attribute.
// If the attribute is not present, it returns an empty string.
//
// Example:
// 	attrs := []html.Attribute{
// 		{Key: "class", Val: "css-10b0gli er34gjf0"},
// 		{Key: "data-testid", Val: "adCard-featured"},
// 	}
// 	getAttr(attrs, "class") // returns "css-10b0gli er34gjf0"
// 	getAttr(attrs, "data-testid") // returns "adCard-featured"
func getAttr(attrs []html.Attribute, key string) string {
    for _, attr := range attrs {
        if attr.Key == key {
            return attr.Val
        }
    }
    return ""
}

// parseOffer parses the given text and returns an Offer struct.
// The text should be the content of a single offer.
//
// Example:
// 	offer := parseOffer(`
// 		<div class="css-1sw7q4x">
// 			<a href="/oferta/mieszkanie-2-pokojowe-ID6Q2Zr.html">
// 				<h6 class="css-1j9dxys e1n63ojh0">Mieszkanie 2 pokojowe</h6>
// 				<p class="css-10b0gli er34gjf0">1 000 zł</p>
// 				<p class="css-veheph er34gjf0">Warszawa, dzisiaj 12:00</p>
// 			</a>
// 		</div>
// 	`)
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
                isPrice = checkAttr(t.Attr, "class", "css-10b0gli er34gjf0")
                isTimeAndLoc = checkAttr(t.Attr, "class", "css-veheph er34gjf0")
			case "a":
                offer.Url = getAttr(t.Attr, "href")
                if offer.Url[0] == '/' {
                    offer.Url = "https://www.olx.pl" + offer.Url
                }
			case "div":
                // Check if the offer is featured
                if (checkAttr(t.Attr, "data-testid", "adCard-featured")) {
                    return Offer{}
                }
			}

		case html.TextToken:
			t := tkn.Token()
			if isTitle {
				offer.Title = t.Data
				isTitle = false
			} else if isPrice {
				offer.Price = t.Data
				isPrice = false
			} else if isTimeAndLoc {
				offer.Location = t.Data
				for i := 0; i < 4; i++ {
					tkn.Next()
				}
				offer.Time = tkn.Token().Data
				isTimeAndLoc = false
			}
		}
	}
}

// ParseHtml parses the given text and returns a list of offers.
// The text should be the content of the page with the offers.
//
// Example:
// 	offers := ParseHtml(`
// 		<div class="css-1sw7q4x">
// 			<a href="/oferta/mieszkanie-2-pokojowe-ID6Q2Zr.html">
// 				<h6 class="css-1j9dxys e1n63ojh0">Mieszkanie 2 pokojowe</h6>
// 				<p class="css-10b0gli er34gjf0">1 000 zł</p>
// 				<p class="css-veheph er34gjf0">Warszawa, dzisiaj 12:00</p>
// 			</a>
// 		</div>`)
func ParseHtml(text string) []Offer {
	tokenizer := html.NewTokenizer(strings.NewReader(text))

	offers := make([]Offer, 0)
	isOffer := false
	var offerContent string
	offerSeparator := "css-1sw7q4x"
	depth := 0

	for {
		tokenType := tokenizer.Next()

		switch tokenType {
		case html.ErrorToken:
			// End of the document, we're done
			return offers

        case html.StartTagToken:
			token := tokenizer.Token()
			if !isOffer {
				if token.Data == "div" {
                    isOffer = checkAttr(token.Attr, "class", offerSeparator)
				}
			} else {
				if token.Data == "div" {
					depth++
				}
				offerContent += token.String()
			}

        case html.EndTagToken:
			token := tokenizer.Token()
			if isOffer && token.Data == "div" && depth == 0 {
				isOffer = false
				offer := parseOffer(offerContent)

				// TODO: For some reason, the last div recognized as offer is empty
				// Inspect this later
				if offer.Title != "" {
					offers = append(offers, offer)
				}
				offerContent = ""
				depth = 0
			} else if isOffer {
				if token.Data == "div" {
					depth--
				}
				offerContent += token.String()
			}

        default:
			if isOffer {
				offerContent += tokenizer.Token().String()
				continue
			}
		}
	}
}
