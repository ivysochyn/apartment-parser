package telegrambot

import (
	"database/sql"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Process callback query.
// Verifies the callback query data and calls the appropriate function.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Telegram update instance.
//	search_db: Search database instance.
func processCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, search_db *sql.DB) {
	data := strings.Split(update.CallbackQuery.Data, "|")
	switch data[0] {

	case "remove_msg":
		removeUpdateQueryMessage(bot, update)
		return

	case "search":
		processSearchAction(bot, update, search_db)

	default:
		log.Println("Unknown callback query data: ", data[0])
	}

	removeUpdateQueryMessage(bot, update)
}
