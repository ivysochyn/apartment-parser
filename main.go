package main

import (
	"apartment-parser/parser"
	"fmt"
	"log"
)

func main() {
	page_url := "https://www.olx.pl/poznan/q-mieszkanie/?search%5Border%5D=created_at:desc&search%5Bfilter_float_price:from%5D=1000&search%5Bfilter_float_price:to%5D=3000"
	parser.FetchHTMLPage(page_url, "")

	filename := "index.html"
	text, err := parser.ReadHtmlFromFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	data := parser.ParseHtml(text)
	for _, offer := range data {
		fmt.Println("--------------------")
		fmt.Println("Title: ", offer.Title)
		fmt.Println("Price: ", offer.Price)
		fmt.Println("Location: ", offer.Location)
		fmt.Println("Time: ", offer.Time)
		fmt.Println("URL: ", offer.Url)
	}
}
