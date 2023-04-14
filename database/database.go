// This package serves to create, update, and delete database entries of offers
// and uses sqlite3 as the database engine.
package database

import (
	"apartment-parser/parser"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Search struct {
	ID  int64
	URL string
}

// AddOffer function creates a new database entry for a new offer.
// It takes in an offer struct and database connection as parameters.
// If the offer already exists in the database, it will not be added.
// It returns an error if the database connection fails.
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

// AddSearch function creates a new database entry for a new search.
// It takes in a user id and a search url as parameters.
// If the search already exists in the database, it will not be added.
// It returns an error if the database connection fails.
//
// Example:
//
//	err := AddSearch(db, 1, "https://www.olx.pl/nieruchomosci/mieszkania/wynajem/warszawa/")
func AddSearch(db *sql.DB, userID int64, url string) error {
	// If the search already exists, do not add it
	exists, err := searchExists(db, userID, url)
	if err != nil || exists {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO searches(UserID, url) VALUES(?, ?)")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(userID, url)
	return err
}

func DeleteSearch(db *sql.DB, ID int64) error {
	stmt, err := db.Prepare("DELETE FROM searches WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(ID)
	return err
}

// searchExists function checks if a search already exists in the database.
// It takes in a user id and a search url as parameters.
// It returns a boolean value and an error if the database connection fails.
//
// Example:
//
//	exists, err := searchExists(db, 1, "https://www.olx.pl/nieruchomosci/mieszkania/wynajem/warszawa/")
func searchExists(db *sql.DB, userID int64, url string) (bool, error) {
	var exists bool
	// if search with the same url exists
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM searches WHERE UserID = ? AND url = ?)", userID, url).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// offerExists function checks if an offer already exists in the database.
// It takes in an offer struct and database connection as parameters.
// It returns a boolean value and an error if the database connection fails.
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
//	exists, err := OfferExists(db, offer)
func offerExists(db *sql.DB, offer parser.Offer) (bool, error) {
	var exists bool
	// if offer with the same title and price exists
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM offers WHERE title = ? AND price = ?)", offer.Title, offer.Price).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CreateOffersDatabase function creates a new database file.
// It takes in a database file name as a parameter.
// It returns a database connection and an error if the database connection fails.
// If the database file already exists, it will be opened instead.
//
// Example:
//
//	db, err := CreateOffersDatabase("offers.db")
func CreateOffersDatabase(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS offers (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, price TEXT, location TEXT, time TEXT, url TEXT)")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// CreateSearchesDatabase function creates a new database file.
// It takes in a database file name as a parameter.
// It returns a database connection and an error if the database connection fails.
// If the database file already exists, it will be opened instead.
//
// Example:
//
//	db, err := CreateSearchesDatabase("searches.db")
func CreateSearchesDatabase(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	// Create a table with a single column for the url and primary key which represents the UserIDs
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS searches (id INTEGER PRIMARY KEY AUTOINCREMENT, UserID INTEGER, url TEXT)")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// GetSearches function gets all searches from the database for a specific user
// It takes in a database connection and a user id as parameters.
// It returns a slice of strings and an error if the database connection fails.
//
// Example:
//
//	searches, err := GetSearches(db, 1)
func GetSearches(db *sql.DB, userID int64) ([]Search, error) {
	var searches []Search
	rows, err := db.Query("SELECT id, url FROM searches WHERE UserID = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var search Search
		err = rows.Scan(&search.ID, &search.URL)
		if err != nil {
			return nil, err
		}
		searches = append(searches, search)
	}
	return searches, nil
}

func GetSearch(db *sql.DB, id int64) (Search, error) {
	var search Search
	err := db.QueryRow("SELECT id, url FROM searches WHERE id = ?", id).Scan(&search.ID, &search.URL)
	if err != nil {
		return Search{}, err
	}
	return search, nil
}

// ListOffers function lists all offers in the database.
// It takes in a database connection as a parameter.
// It returns a slice of offer structs and an error if the database connection fails.
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
