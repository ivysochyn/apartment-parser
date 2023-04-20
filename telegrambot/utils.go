package telegrambot

import (
	"errors"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Structure for storing user search parameters
//
// Attributes:
//
//	user_id: ID of the user that creates the search.
//	state: State of the search creation process.
//	city: City to create the search for.
type UserNewSearch struct {
	user_id int64
	state   string
	city    string
}

// Structure for representing a city.
//
// Attributes:
//
//	Name: Name of the city to represent.
//	Code: Encoded name of the city used in the URL.
type City struct {
	Name string
	Code string
}

// Remove the update message using a callback query.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Update message to remove.
func removeUpdateQueryMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	deleteMessage := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	sendDeleteMessage(bot, deleteMessage)
}

// Remove the update message using a message.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Update message to remove.
func removeUpdateMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
	sendDeleteMessage(bot, deleteMessage)
}

// Remove the update message using a message and a relative message ID.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	update: Update message to use as a relative point.
//	relativeMessageID: Amount of messages back to move from the update message.
func removeUpdateMessageRelative(bot *tgbotapi.BotAPI, update tgbotapi.Update, relativeMessageID int) {
	deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID-relativeMessageID)
	sendDeleteMessage(bot, deleteMessage)
}

// Send a message using a MessageConfig.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	msg: Message to send.
func sendMessage(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

// Send a delete message using a DeleteMessageConfig.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	msg: DeleteMessageConfig to send.
func sendDeleteMessage(bot *tgbotapi.BotAPI, msg tgbotapi.DeleteMessageConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

// Process a price range string and return the min and max price.
//
// Parameters:
//
//	price_range_str: Price range string to process.
//
// Returns:
//
//	minPrice: Minimum price of the range.
//	maxPrice: Maximum price of the range.
//	error: Error if the price range string is not valid.
func processPriceStr(price_range_str string) (int, int, error) {
	price_range := strings.Split(price_range_str, "-")

	if len(price_range) != 2 {
		return 0, 0, errors.New("Price range is not valid")
	}

	minPrice, err := strconv.Atoi(price_range[0])
	if err != nil {
		return 0, 0, errors.New("minPrice is not valid")
	}

	maxPrice, err := strconv.Atoi(price_range[1])
	if err != nil {
		return 0, 0, errors.New("maxPrice is not valid")
	}

	if minPrice > maxPrice {
		return 0, 0, errors.New("minPrice is higher than maxPrice")
	}

	return minPrice, maxPrice, nil
}
