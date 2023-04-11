package parser

import (
	"golang.org/x/net/html"
	"strings"
)

type Offer struct {
	Title    string
	Price    string
	Location string
	Time     string
	Url      string
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
						offer.Url = url
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

func ParseHtml(text string) (data []Offer) {
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
				if offer.Title != "" {
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

