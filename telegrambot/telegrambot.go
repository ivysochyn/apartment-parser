package telegrambot

import (
	"apartment-parser/database"

	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Stores user states for new searches
var userStates = make(map[int64]UserNewSearch)

// Keyboard for the bot
var keyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Searches 🔍"),
	),
)

// Create a new bot.
//
// Parameters:
//
//	debug: If true, the bot will print debug messages.
//
// Returns:
//
//	bot: A pointer to the bot object.
//	err: An error if the bot could not be created.
func createBot(debug bool) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	bot.Debug = debug
	return bot, err
}

// Start the telegram bot.
// Opens the database and starts listening for updates.
//
// Parameters:
//
//	debug: If true, the bot will print debug messages.
func StartBot(debug bool) {
	bot, err := createBot(debug)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	search_db, err := database.OpenSearchesDatabase("searches.db")
	if err != nil {
		log.Println(err)
	}

	offers_db, err := database.OpenOffersDatabase("offers.db")
	if err != nil {
		log.Println(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	go parseOffers(bot, offers_db, search_db)

	// Handle updates
	for update := range updates {
		if update.CallbackQuery != nil {
			processCallbackQuery(bot, update, search_db)
		}

		if update.Message != nil {
			processMessage(bot, update, search_db)
		}
	}
}
