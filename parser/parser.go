// Package parser contains functions for parsing the HTML code of the website.
//
// It is used to extract all the non-featured offers of apartments for rent from a given link.
package parser

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
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
	Price             int
	Location          string
	Time              string
	Url               string
	AdditionalPayment int
	Description       string
	Rooms             string
	Area              string
	Floor             string
	Images            []string
}

// ExtractorConfig holds configuration for the offer extractor
type ExtractorConfig struct {
	// Selectors for finding elements
	TitleSelector      Selector
	PriceSelector      Selector
	LocationSelector   Selector
	URLSelector        Selector

	// Parsing configuration
	DatePattern        *regexp.Regexp
	TimePattern        *regexp.Regexp
	PricePattern       *regexp.Regexp
	TodayKeyword       string
	BaseURL            string
	TimezoneOffset     time.Duration
}

// Selector represents how to find an element
type Selector struct {
	Tag       string
	Attribute string
	Value     string
}

// Default configuration for OLX
var OLXConfig = ExtractorConfig{
	TitleSelector: Selector{
		Tag:       "h4",  // More stable than h6
		Attribute: "",
		Value:     "",
	},
	PriceSelector: Selector{
		Tag:       "p",
		Attribute: "data-testid",
		Value:     "ad-price",
	},
	LocationSelector: Selector{
		Tag:       "p",
		Attribute: "data-testid",
		Value:     "location-date",
	},
	URLSelector: Selector{
		Tag:       "a",
		Attribute: "href",
		Value:     "",
	},
	DatePattern:    regexp.MustCompile(`\d{1,2}\s+\w+\s+\d{4}`),
	TimePattern:    regexp.MustCompile(`\d{2}:\d{2}`),
	PricePattern:   regexp.MustCompile(`\d+`),
	TodayKeyword:   "Dzisiaj",
	BaseURL:        "https://www.olx.pl",
	TimezoneOffset: 2 * time.Hour, // Poland is UTC+2
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
				isDescription = checkAttr(t.Attr, "class", "css-19duwlz")
			case "p":
				isTag = checkAttr(t.Attr, "class", "css-5l1a1j")
			}

		case html.TextToken:
			if isDescription {
				offer.Description += string(tkn.Text())
			} else if isTag {
				data := string(tkn.Text())
				if strings.HasPrefix(data, "Czynsz") {
					data = strings.ReplaceAll(data, " ", "")
					data = regexp.MustCompile(`\d+`).FindString(data)
					if data == "" {
						offer.AdditionalPayment = 0
					}
					offer.AdditionalPayment, err = strconv.Atoi(data)
					if err != nil {
						offer.AdditionalPayment = 0
					}
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

	var (
		isDescription, isJson bool
		currentTag            string
		isTagLabel            bool
		isTagValue            bool
		jsonText              string
	)

	for {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			// End of the document
			return offer

		case html.StartTagToken:
			t := tkn.Token()

			if t.Data == "p" && getAttr(t.Attr, "class") == "e1wd2yzk2 css-1airkmu" {
				if getAttr(t.Attr, "data-sentry-element") == "Item" {
					isTagLabel = true
				} else {
					isTagValue = true
				}
			}

			isDescription = getAttr(t.Attr, "data-cy") == "adPageAdDescription"
			isJson = (t.Data == "script") && checkAttr(t.Attr, "type", "application/json")

		case html.TextToken:
			text := strings.TrimSpace(string(tkn.Text()))

			if isTagLabel {
				currentTag = strings.TrimSuffix(text, ":")
				isTagLabel = false
			} else if isTagValue && currentTag != "" {
				switch currentTag {
				case "Powierzchnia":
					offer.Area = currentTag + ": " + text
				case "Liczba pokoi":
					offer.Rooms = currentTag + ": " + text
				case "Piętro":
					offer.Floor = currentTag + ": " + text
				case "Czynsz":
					val := strings.ReplaceAll(text, " ", "")
					val = regexp.MustCompile(`\d+`).FindString(val)
					if v, err := strconv.Atoi(val); err == nil {
						offer.AdditionalPayment = v
					}
				}
				currentTag = ""
				isTagValue = false
			}

			if isDescription {
				offer.Description += text + "\n"
			}

			if isJson {
				jsonText += text
			}

		case html.EndTagToken:
			t := tkn.Token()

			if t.Data == "script" && isJson {
				isJson = false
				offer.Images, err = parseOtodomImages(jsonText)
				if err != nil {
					log.Println(err)
				}
			} else if t.Data == "div" && isDescription {
				isDescription = false
				if len(offer.Description) > 0 {
					offer.Description = strings.TrimSuffix(offer.Description, "\n")
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
	return extractOfferWithConfig(text, OLXConfig)
}

func extractOfferWithConfig(text string, config ExtractorConfig) Offer {
	tkn := html.NewTokenizer(strings.NewReader(text))
	offer := Offer{}

	if strings.Contains(text, "Wyróżnione") {
		log.Println("[DEBUG] Skipping featured ad (Wyróżnione found)")
		return Offer{}
	}

	// State tracking
	var currentContext string
	var isInLink bool
	var depth int

	for {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			if offer.Url == "" {
				return Offer{}
			}
			return offer

		case html.StartTagToken:
			t := tkn.Token()

			// Track if we're inside a link
			if t.Data == "a" {
				isInLink = true
				// Extract URL
				if url := getAttr(t.Attr, "href"); url != "" && offer.Url == "" {
					offer.Url = normalizeURL(url, config.BaseURL)
				}
			}

			// Identify context based on data attributes
			if attr := getAttr(t.Attr, "data-cy"); attr != "" {
				currentContext = attr
			}
			if attr := getAttr(t.Attr, "data-testid"); attr != "" {
				currentContext = attr
			}

			// Check specific selectors
			if matchesSelector(t, config.PriceSelector) {
				currentContext = "price"
			} else if matchesSelector(t, config.LocationSelector) {
				currentContext = "location-date"
			} else if matchesSelector(t, config.TitleSelector) && isInLink {
				currentContext = "title"
			}

			depth++

		case html.EndTagToken:
			t := tkn.Token()
			if t.Data == "a" {
				isInLink = false
			}
			depth--

		case html.TextToken:
			text := strings.TrimSpace(tkn.Token().Data)
			if text == "" {
				continue
			}

			switch currentContext {
			case "title":
				if offer.Title == "" {
					offer.Title = text
					log.Printf("[DEBUG] Found title: %s", text)
				}

			case "price", "ad-price":
				price := extractPrice(text, config.PricePattern)
				if price > 0 {
					offer.Price = price
					log.Printf("[DEBUG] Found price: %d", price)
				}

			case "location-date":
				location, timeStr := extractLocationAndTime(text, config)
				if offer.Location == "" && location != "" {
					offer.Location = location
					log.Printf("[DEBUG] Found location: %s", location)
				}
				if offer.Time == "" && timeStr != "" {
					offer.Time = timeStr
					log.Printf("[DEBUG] Found time: %s", timeStr)
				}
			}
		}
	}
}

// Helper functions

func matchesSelector(token html.Token, selector Selector) bool {
	if selector.Tag != "" && token.Data != selector.Tag {
		return false
	}
	if selector.Attribute != "" && selector.Value != "" {
		return checkAttr(token.Attr, selector.Attribute, selector.Value)
	}
	return selector.Tag == token.Data
}

func normalizeURL(url, baseURL string) string {
	if strings.HasPrefix(url, "/") {
		return baseURL + url
	}
	if !strings.HasPrefix(url, "http") {
		return baseURL + "/" + url
	}
	return url
}

func extractPrice(text string, pattern *regexp.Regexp) int {
	// Remove spaces and find numbers
	text = strings.ReplaceAll(text, " ", "")
	text = strings.ReplaceAll(text, "\u00a0", "") // non-breaking space

	matches := pattern.FindAllString(text, -1)
	if len(matches) > 0 {
		// Join all numbers (handles prices like "1 700")
		priceStr := strings.Join(matches, "")
		if price, err := strconv.Atoi(priceStr); err == nil {
			return price
		}
	}
	return 0
}

func extractLocationAndTime(text string, config ExtractorConfig) (location, timeStr string) {
	// Split by common separators
	parts := strings.Split(text, " - ")
	if len(parts) >= 2 {
		location = strings.TrimSpace(parts[0])
		dateTimeStr := strings.TrimSpace(parts[1])

		// Check if it's today
		if !strings.Contains(dateTimeStr, config.TodayKeyword) {
			log.Printf("[DEBUG] Skipping non-today offer: %s", dateTimeStr)
			return location, ""
		}

		// Extract time
		if matches := config.TimePattern.FindStringSubmatch(dateTimeStr); len(matches) > 0 {
			if t, err := time.Parse("15:04", matches[0]); err == nil {
				// Adjust timezone
				t = t.Add(config.TimezoneOffset)
				timeStr = t.Format("15:04")
			}
		}
	} else {
		// Try to extract location from the whole text
		location = text
	}

	return location, timeStr
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
