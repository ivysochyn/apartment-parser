package telegrambot

import (
	"database/sql"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Process message.
// Processes the message and calls the appropriate function.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Telegram update.
//	db: Database instance of the search database.
func processMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	if update.Message.IsCommand() {
		processCommand(bot, update)
	}

	if update.Message.Text == "Searches 🔍" {
		displayAllSearchesToUser(bot, update.Message.Chat.ID, db)
	}

	// If user exists in userStates
	if userState, ok := userStates[update.Message.Chat.ID]; ok {
		if userState.state == "search|price" {
			newSearchProcessPrice(bot, update, db)
		}
	}

	// Remove last user's message
	removeUpdateMessage(bot, update)
}

// Process command.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Telegram update.
func processCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyMarkup = keyboard
		msg.Text = "Welcome to the " + bot.Self.UserName + "🏠"
		sendMessage(bot, msg)
	}
}
