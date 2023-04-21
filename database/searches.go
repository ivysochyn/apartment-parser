// Responsible for managing searches in the database.
package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Search struct represents a search in the database.
//
// Attributes:
//
//	ID - search id
//	UserID - user id of the user who added the search
//	URL - search url
type Search struct {
	ID     int64
	UserID int64
	URL    string
}

// Create a new database entry for a new search.
// If the search already exists, it will not be added.
//
// Parameters:
//
//	db - database connection
//	userID - user id of the user who added the search
//	url - search url
//
// Returns:
//
//	error - error if the database connection fails
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

// Check if search already exists in the database.
//
// Parameters:
//
//	db - database connection
//	userID - user id of the user who added the search
//	url - search url
//
// Returns:
//
//	bool - true if the search exists, false otherwise
//	error - error if the database connection fails
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

// Delete a search from the database.
//
// Parameters:
//
//	db - database connection
//	ID - search id
//
// Returns:
//
//	error - error if the database connection fails
//
// Example:
//
//	err := DeleteSearch(db, 1)
func DeleteSearch(db *sql.DB, ID int64) error {
	stmt, err := db.Prepare("DELETE FROM searches WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(ID)
	return err
}

// Lists all searches from the database related to a specific user.
//
// Parameters:
//
//	db - database connection
//	userID - user id searches belong to
//
// Returns:
//
//	[]Search - list of searches
//	error - error if the database connection fails
//
// Example:
//
//	searches, err := ListSearches(db, 1)
func ListSearches(db *sql.DB, userID int64) ([]Search, error) {
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

// Get a search from the database by its id.
//
// Parameters:
//
//	db - database connection
//	id - search id
//
// Returns:
//
//	Search - search
//	error - error if the database connection fails
//
// Example:
//
//	search, err := GetSearch(db, 1)
func GetSearch(db *sql.DB, id int64) (Search, error) {
	var search Search
	err := db.QueryRow("SELECT id, url FROM searches WHERE id = ?", id).Scan(&search.ID, &search.URL)
	if err != nil {
		return Search{}, err
	}
	return search, nil
}

func GetAllSearches(db *sql.DB) ([]Search, error) {
	var searches []Search
	rows, err := db.Query("SELECT id, url, UserID FROM searches")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var search Search
		err = rows.Scan(&search.ID, &search.URL, &search.UserID)
		if err != nil {
			return nil, err
		}
		searches = append(searches, search)
	}

	return searches, nil
}
