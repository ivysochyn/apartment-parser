// Responsible for adding and listing offers in the database.
package database

import (
	"apartment-parser/parser"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Add offer to the database.
//
// Parameters:
//
//	db - database connection
//	offer - offer struct
//
// Returns:
//
//	error - error if the database connection fails
//
// Example:
//
//	offer := parser.Offer{
//		Title: "Mieszkanie 2 pokojowe",
//		Price: "1 000 zł",
//		Location: "Warszawa",
//		Time: "dzisiaj 12:00",
//		Url: "https://www.olx.pl/oferta/mieszkanie-2-pokojowe-ID6Q2Zr.html"
//	}
//	err := AddOffer(db, offer)
func AddOffer(db *sql.DB, offer parser.Offer) error {
	exists, err := offerExists(db, offer)

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	stmt, err := db.Prepare("INSERT INTO offers(title, price, location, time, url) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(offer.Title, offer.Price, offer.Location, offer.Time, offer.Url)
	return err
}

// Check if offer exists in the database.
//
// Parameters:
//
//	db - database connection
//	offer - offer struct
//
// Returns:
//
//	bool - true if offer exists, false otherwise
//	error - error if the database connection fails
//
// Example:
//
//	offer := parser.Offer{
//		Title: "Mieszkanie 2 pokojowe",
//		Price: "1 000 zł",
//		Location: "Warszawa",
//		Time: "dzisiaj 12:00",
//		Url: "https://www.olx.pl/oferta/mieszkanie-2-pokojowe-ID6Q2Zr.html"
//	}
//	exists, err := offerExists(db, offer)
func offerExists(db *sql.DB, offer parser.Offer) (bool, error) {
	var exists bool
	// if offer with the same title and price exists
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM offers WHERE title = ? AND price = ?)", offer.Title, offer.Price).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// List all offers from the database.
//
// Parameters:
//
//	db - database connection
//	offer - offer struct
//
// Returns:
//
//	[]parser.Offer - list of offers
//	error - error if the database connection fails
//
// Example:
//
//	offers, err := ListOffers(db)
func ListOffers(db *sql.DB) ([]parser.Offer, error) {
	var offers []parser.Offer
	rows, err := db.Query("SELECT * FROM offers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var offer parser.Offer
		var id int
		err = rows.Scan(&id, &offer.Title, &offer.Price, &offer.Location, &offer.Time, &offer.Url)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}
	return offers, nil
}
