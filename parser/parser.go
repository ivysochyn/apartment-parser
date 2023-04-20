// Package parser contains functions for parsing the HTML code of the website.
//
// It is used to extract all the non-featured offers of apartments for rent from a given link.
package parser

import (
	"golang.org/x/net/html"
	"strings"
)

// Offer of an apartment for rent.
//
// Attributes:
//
//	Title: The title of the offer.
//	Price: The price of the offer.
//	Location: The location of the offer.
//	Time: The time offer was posted or updated.
//	Url: The url of the offer.
//	AdditionalPayment: The additional payment for the offer.
//	Description: The description of the offer.
//	Rooms: The number of rooms of the offer.
//	Area: The area of the offer.
//	Floor: The floor of the offer.
type Offer struct {
	Title             string
	Price             string
	Location          string
	Time              string
	Url               string
	AdditionalPayment string
	Description       string
	Rooms             string
	Area              string
	Floor             string
}

// Check if the given attribute is present in the given list of attributes.
//
// Parameters:
//
//	attrs: The list of attributes.
//	key: The key of the attribute.
//	value: The value of the attribute.
//
// Returns:
//
//	True if the attribute is present, false otherwise.
func checkAttr(attrs []html.Attribute, key, value string) bool {
	for _, attr := range attrs {
		if attr.Key == key && attr.Val == value {
			return true
		}
	}
	return false
}

// Get the value of the given attribute.
//
// Parameters:
//
//	attrs: The list of attributes.
//	key: The key of the attribute.
//
// Returns:
//
//	The value of the attribute.
func getAttr(attrs []html.Attribute, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// Parse the given offer by following the url and extracting the missing data.
//
// Parameters:
//
//	offer: The offer to parse.
//
// Returns:
//
//	The parsed offer.
func ParseOffer(offer Offer) Offer {
	// If url starts with www.olx.pl
	if strings.HasPrefix(offer.Url, "https://www.olx.pl") {
		offer = parseOlxOffer(offer)
	} else if strings.HasPrefix(offer.Url, "https://www.otodom.pl") {
		offer = parseOtodomOffer(offer)
	}
	return offer
}

// Parse the olx offer.
//
// Parameters:
//
//	offer: The offer to parse.
//
// Returns:
//
//	The parsed offer.
func parseOlxOffer(offer Offer) Offer {
	text, err := FetchHTMLPage(offer.Url)

	if err != nil {
		return offer
	}

	tkn := html.NewTokenizer(strings.NewReader(text))

	var isDescription bool
	var isTag bool

	for {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			// End of the document, we're done
			return offer

		case html.StartTagToken:
			t := tkn.Token()
			switch t.Data {
			case "div":
				isDescription = checkAttr(t.Attr, "class", "css-bgzo2k er34gjf0")
			case "p":
				isTag = checkAttr(t.Attr, "class", "css-b5m1rv er34gjf0")
			}

		case html.TextToken:
			if isDescription {
				offer.Description += string(tkn.Text())
			} else if isTag {
				data := string(tkn.Text())
				if strings.HasPrefix(data, "Czynsz") {
					// TODO: Extract the additional payment and convert it to a number
					offer.AdditionalPayment = data
				} else if strings.HasPrefix(data, "Liczba pokoi") {
					// TODO: Extract the number of rooms and convert it to a number
					offer.Rooms += data
				} else if strings.HasPrefix(data, "Powierzchnia") {
					// TODO: Extract the area number and convert it to a number
					offer.Area += data
				} else if strings.HasPrefix(data, "Poziom") {
					// TODO: Extract the floor number and convert it to a number
					offer.Floor += data
				}
			}

		case html.EndTagToken:
			t := tkn.Token()
			if t.Data == "div" {
				if isDescription {
					isDescription = false
				}
			} else if t.Data == "p" {
				if isTag {
					isTag = false
				}
			}

		}
	}
}

// Parse the otodom offer.
//
// Parameters:
//
//	offer: The offer to parse.
//
// Returns:
//
//	The parsed offer.
func parseOtodomOffer(offer Offer) Offer {
	text, err := FetchHTMLPage(offer.Url)

	if err != nil {
		return offer
	}

	tkn := html.NewTokenizer(strings.NewReader(text))

	var isArea, isRooms, isFloor, isAdditionalPayment, isDescription, isText bool
	var deep_counter int

	for {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			// End of the document, we're done
			return offer

		case html.StartTagToken:
			t := tkn.Token()
			switch t.Data {
			case "div":
				deep_counter++
				if checkAttr(t.Attr, "class", "css-kkaknb enb64yk0") {
					// Get the 'aria-label' attribute
					ariaLabel := getAttr(t.Attr, "aria-label")
					switch ariaLabel {
					case "Powierzchnia":
						isArea = true
						isText, isRooms, isFloor, isAdditionalPayment, isDescription = false, false, false, false, false
						deep_counter = 1
					case "Liczba pokoi":
						isRooms = true
						isText, isArea, isFloor, isAdditionalPayment, isDescription = false, false, false, false, false
						deep_counter = 1
					case "Piętro":
						isFloor = true
						isText, isArea, isRooms, isAdditionalPayment, isDescription = false, false, false, false, false
						deep_counter = 1
					case "Czynsz":
						isAdditionalPayment = true
						isText, isArea, isRooms, isFloor, isDescription = false, false, false, false, false
						deep_counter = 1
					}
				} else if checkAttr(t.Attr, "class", "css-1wi2w6s enb64yk4") {
					isText = true
				} else if checkAttr(t.Attr, "class", "css-1wekrze e1lbnp621") {
					isDescription = true
				}
			}

		case html.TextToken:
			if isText {
				if isArea {
					offer.Area = string(tkn.Text())

				} else if isRooms {
					offer.Rooms = string(tkn.Text())

				} else if isFloor {
					offer.Floor = string(tkn.Text())

				} else if isAdditionalPayment {
					offer.AdditionalPayment = string(tkn.Text())
				}
			}

		case html.EndTagToken:
			t := tkn.Token()
			if t.Data == "div" {
				deep_counter--
				if deep_counter == 0 {
					if isArea {
						isArea = false
					} else if isRooms {
						isRooms = false
					} else if isFloor {
						isFloor = false
					} else if isAdditionalPayment {
						isAdditionalPayment = false
					} else if isDescription {
						isDescription = false
					}
				}

			}
		}
	}
}

// Extract all the offers from the given block of code.
//
// Parameters:
//
//	text: The block of code to parse.
//
// Returns:
//
//	The offer extracted from the block of code.
func extractOffer(text string) Offer {
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
			if offer.Url == "" {
				return Offer{}
			}
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
				if checkAttr(t.Attr, "data-testid", "adCard-featured") {
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

// Parse the HTML code and extract all the offers.
//
// Parameters:
//
//	text: The HTML code to parse.
//
// Returns:
//
//	The offers extracted from the HTML code.
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
				offer := extractOffer(offerContent)

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
