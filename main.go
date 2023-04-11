package main

import (
	"apartment-parser/database"
	"apartment-parser/parser"
	"fmt"
	"log"
)

func main() {
	url := "https://www.olx.pl/poznan/q-mieszkanie/?search%5Border%5D=created_at:desc&search%5Bfilter_float_price:from%5D=1000&search%5Bfilter_float_price:to%5D=3000"
	db, err := database.CreateDatabase("apartments.db")

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	body, err := parser.FetchHTMLPage(url)

	if err != nil {
		log.Fatal(err)
	}

	data := parser.ParseHtml(body)
	for _, offer := range data {
		err := database.AddOffer(db, offer)
		if err != nil {
			log.Fatal(err)
		}
	}

	all_offers, _ := database.ListOffers(db)
	for _, offer := range all_offers {
		fmt.Println(offer)
	}
}
