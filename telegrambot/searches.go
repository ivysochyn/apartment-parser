package telegrambot

import (
	"apartment-parser/database"
	"apartment-parser/parser"

	"database/sql"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handle search actions from callback query.
// Decides what to do based on the data field of the callback query.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Telegram update.
//	db: Database instance of the search database.
func processSearchAction(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	data := strings.Split(update.CallbackQuery.Data, "|")
	switch data[1] {

	case "create_search":
		newSearchListCities(bot, update, db)

	case "list_info":
		displayFullSearchInfo(bot, update.CallbackQuery.Message.Chat.ID, data[2], db)

	case "remove_search":
		removeSearchFromDatabase(data[2], db)
		displayAllSearchesToUser(bot, update.CallbackQuery.Message.Chat.ID, db)

	case "choose_city":
		newSearchProcessCity(bot, update.CallbackQuery.Message.Chat.ID, data[2], db)

	case "cancel_new_search":
		delete(userStates, update.CallbackQuery.Message.Chat.ID)

	default:
		log.Println("Unknown callback query data for search: ", data[1])
	}
}

// Remove a search from the database.
//
// Parameters:
//
//	search_id_str: Search ID as string.
//	db: Database instance of the search database.
func removeSearchFromDatabase(search_id_str string, db *sql.DB) {
	search_id, err := strconv.Atoi(search_id_str)
	if err != nil {
		log.Println(err)
	}
	err = database.DeleteSearch(db, int64(search_id))
	if err != nil {
		log.Println(err)
	}
}

// Display a list of all cities that can be used to create a new search.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Telegram update.
//	db: Database instance of the search database.
func newSearchListCities(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.Text)
	msg.Text = "🌇 Choose the city you want to search in"
	reply_markup := tgbotapi.NewInlineKeyboardMarkup()

	for i := 0; i < len(cities); i += 3 {
		row := tgbotapi.NewInlineKeyboardRow()

		for j := 0; j < 3; j++ {
			if i+j < len(cities) {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(cities[i+j].Name, "search|choose_city|"+cities[i+j].Code))
			}
		}
		reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, row)
	}

	reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "remove_msg|"),
	))

	msg.ReplyMarkup = reply_markup
	sendMessage(bot, msg)
}

// Display a list of all searches that the user has created.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	userID: Telegram user ID.
//	db: Database instance of the search database.
func displayAllSearchesToUser(bot *tgbotapi.BotAPI, userID int64, db *sql.DB) {
	msg := tgbotapi.NewMessage(userID, "")

	searches, err := database.ListSearches(db, userID)
	if err != nil {
		log.Println(err)
	}

	reply_markup := tgbotapi.NewInlineKeyboardMarkup()

	if len(searches) == 0 {
		msg.Text = "❌ You have 0 active searches"
	} else {
		msg.Text = "🔍 You have " + strconv.Itoa(len(searches)) + " searches"

		for _, search := range searches {
			search_info, err := parser.GetSearchShortInfo(search.URL)
			if err != nil {
				log.Println(err)
			} else {
				// Add button to reply_markup
				reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💵 "+search_info, "search|list_info|"+strconv.Itoa(int(search.ID))),
				))
			}
		}
	}

	// Add button to create new search
	reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "remove_msg|"),
		tgbotapi.NewInlineKeyboardButtonData("🟢 Create new search", "search|create_search|"),
	))
	msg.ReplyMarkup = reply_markup
	sendMessage(bot, msg)
}

// Display the full search info of a search.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	userID: Telegram user ID.
//	search_id_str: Search ID as string.
//	db: Database instance of the search database.
func displayFullSearchInfo(bot *tgbotapi.BotAPI, userID int64, search_id_str string, db *sql.DB) {
	search_id, err := strconv.Atoi(search_id_str)
	if err != nil {
		log.Println(err)
		return
	}

	msg := tgbotapi.NewMessage(userID, "")
	search, err := database.GetSearch(db, int64(search_id))
	if err != nil {
		log.Println(err)
		return
	}

	search_info, err := parser.GetSearchFullInfo(search.URL)
	if err != nil {
		log.Println(err)
		return
	}

	msg.Text = search_info
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "remove_msg|"),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Delete search", "search|remove_search|"+strconv.Itoa(int(search.ID))),
		),
	)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	sendMessage(bot, msg)
}

// Process the city selection of a new search.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	userID: Telegram user ID.
//	city: Name of the city.
//	db: Database instance of the search database.
func newSearchProcessCity(bot *tgbotapi.BotAPI, userID int64, city string, db *sql.DB) {

	msg := tgbotapi.NewMessage(userID, "")
	msg.Text = "💵 Write the price range in PLN (e.g. 1000-2000)"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "search|cancel_new_search|"),
		),
	)

	// Add user to userStates
	userStates[userID] = UserNewSearch{
		user_id: userID,
		state:   "search|price",
		city:    city,
	}

	sendMessage(bot, msg)
}

// Process the price range of a new search and create the search.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Telegram update.
//	db: Database instance of the search database.
func newSearchProcessPrice(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	minPrice, maxPrice, err := processPriceStr(update.Message.Text)

	if err != nil {
		log.Println(err)

		msg.Text = "❌ Invalid price range. Please try again."
		sendMessage(bot, msg)
		delete(userStates, update.Message.Chat.ID)

		// Remove the previous message and display all searches again
		removeUpdateMessageRelative(bot, update, 1)
		displayAllSearchesToUser(bot, update.Message.Chat.ID, db)
		return
	}

	search_term := parser.SearchTerm{
		Location:  userStates[update.Message.Chat.ID].city,
		Price_min: float64(minPrice),
		Price_max: float64(maxPrice),
	}

	url, error := parser.CreateUrl(search_term)
	if error != nil {
		log.Println(err)

		msg.Text = "❌ Failed to create a url. Please try again."
		sendMessage(bot, msg)

		delete(userStates, update.Message.Chat.ID)

		// Remove the previous message and display all searches again
		removeUpdateMessageRelative(bot, update, 1)
		displayAllSearchesToUser(bot, update.Message.Chat.ID, db)
		return
	}

	// Add search to database
	err = database.AddSearch(db, update.Message.Chat.ID, url)

	if err != nil {
		log.Println(err)

		msg.Text = "❌ Failed to add to database. Please try again."
		sendMessage(bot, msg)
	}

	// Remove user from userStates
	delete(userStates, update.Message.Chat.ID)

	// Remove the last bot's message
	removeUpdateMessageRelative(bot, update, 1)
	displayAllSearchesToUser(bot, update.Message.Chat.ID, db)
}
