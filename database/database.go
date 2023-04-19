// This package serves to manipulate different databases.
// Uses sqlite3 as the database engine.
package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Connect to the offers database.
// Creates a new database file if it does not exist.
//
// Parameters:
//
//	dbName: Name of the database file.
//
// Returns:
//
//	*sql.DB: Database object.
//	error: Error object.
//
// Example:
//
//	db, err := OpenOffersDatabase("offers.db")
func OpenOffersDatabase(dbName string) (*sql.DB, error) {
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

// Connect to the searches database.
// Creates a new database file if it does not exist.
//
// Parameters:
//
//	dbName: Name of the database file.
//
// Returns:
//
//	*sql.DB: Database object.
//	error: Error object.
//
// Example:
//
//	db, err := OpenSearchesDatabase("searches.db")
func OpenSearchesDatabase(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS searches (id INTEGER PRIMARY KEY AUTOINCREMENT, UserID INTEGER, url TEXT)")
	if err != nil {
		return nil, err
	}
	return db, nil
}
