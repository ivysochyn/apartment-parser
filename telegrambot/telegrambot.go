package telegrambot

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var keyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Searches 🔍"),
	),
)

func handleMessages(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message.IsCommand() {
		handleCommand(bot, update)
	}

	if update.Message.Text == "Searches 🔍" {
		handleSearches(bot, update)
	}

	// Remove last user's message
	deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
	_, err := bot.Send(deleteMessage)
	if err != nil {
		log.Println(err)
	}
}

func handleSearches(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.ReplyMarkup = keyboard
	// TODO: List all active searches
	msg.Text = "❌ You have 0 active searches"

	// Add button to create new search
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🟢 Create new search", "create"),
		),
	)

	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func handleCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyMarkup = keyboard
		msg.Text = "Welcome to the " + bot.Self.UserName + "🏠"
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

func handleCreateSearch(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	// TODO: Implement
}

func createBot(debug bool) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	bot.Debug = debug
	return bot, err
}

func StartBot(debug bool) {
	bot, err := createBot(debug)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Handle updates
	for update := range updates {
		if update.CallbackQuery != nil {
			// Check if user clicked on "Create new search" button
			if update.CallbackQuery.Data == "create" {
				// TODO: Create new search
				handleCreateSearch(bot, update)
			}
			// Delete callback query
			deleteMessage := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
			_, err := bot.Send(deleteMessage)
			if err != nil {
				log.Println(err)
			}
		}

		if update.Message != nil {
			handleMessages(bot, update)
		}
	}
}
