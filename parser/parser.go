// Package parser contains functions for parsing the HTML code of the website.
//
// It is used to extract all the non-featured offers of apartments for rent from a given link.
package parser

import (
	"golang.org/x/net/html"
	"log"
	"regexp"
	"strings"
	"time"
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
	Images            []string
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
		log.Printf("Error fetching the OLX page: %v", err)
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
				isDescription = checkAttr(t.Attr, "class", "css-1t507yq er34gjf0")
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
			if t.Data == "div" && isDescription {
				isDescription = false
			} else if t.Data == "p" && isTag {
				isTag = false
			}

		case html.SelfClosingTagToken:
			t := tkn.Token()
			if t.Data == "img" {
				if checkAttr(t.Attr, "class", "css-1bmvjcs") {
					offer.Images = append(offer.Images, getAttr(t.Attr, "src"))
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
		log.Printf("Error fetching the Otodom page: %v", err)
		return offer
	}

	tkn := html.NewTokenizer(strings.NewReader(text))

	var isRooms, isDescription, isArea, isFloor, isAdditionalPayment, isJson bool
	var json string = ""

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
				// Tags
				dataid := getAttr(t.Attr, "data-testid")
				switch dataid {
				case "table-value-floor":
					isFloor = true
				case "table-value-rent":
					isAdditionalPayment = true
				case "table-value-area":
					isArea = true
				case "table-value-rooms_num":
					isRooms = true
				}
				isDescription = getAttr(t.Attr, "data-cy") == "adPageAdDescription"
			case "a":
				isRooms = getAttr(t.Attr, "data-cy") == "ad-information-link"
			case "script":
				isJson = checkAttr(t.Attr, "type", "application/json")
			}

		case html.TextToken:
			if isArea {
				offer.Area = string(tkn.Text())
				isArea = false
			} else if isFloor {
				offer.Floor = string(tkn.Text())
				isFloor = false
			} else if isAdditionalPayment {
				offer.AdditionalPayment = string(tkn.Text())
				isAdditionalPayment = false
			} else if isRooms {
				offer.Rooms = string(tkn.Text())
				isRooms = false
			} else if isDescription {
				offer.Description += string(tkn.Text()) + "\n"
			}

			if isJson {
				json += string(tkn.Text())
			}

		case html.EndTagToken:
			t := tkn.Token()
			if t.Data == "script" && isJson {
				isJson = false
				offer.Images, err = parseOtodomImages(json)
				if err != nil {
					log.Println(err)
				}
			} else if t.Data == "div" && isDescription {
				isDescription = false
				// Strip the last newline character if present
				if len(offer.Description) > 0 {
					offer.Description = offer.Description[:len(offer.Description)-1]
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
				isTimeAndLoc = checkAttr(t.Attr, "class", "css-1a4brun er34gjf0")
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
				// Exit if the date is not today
				date_str := tkn.Token().Data
				if !strings.Contains(date_str, "Dzisiaj") {
					return Offer{}
				}
				// Convert the time_str from UTC to the GTM+1 timezone
				re_time := regexp.MustCompile(`\d{2}:\d{2}`)
				t, err := time.Parse("15:04", re_time.FindString(date_str))
				if err != nil {
					log.Println(err)
					return Offer{}
				}
				t = t.Add(time.Hour)
				offer.Time = t.Format("15:04")
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
