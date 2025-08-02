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
	// If the message has reply markup, remove that message first
	if update.CallbackQuery.Message.ReplyMarkup != nil {
		// Check if previous message is a media group
		if update.CallbackQuery.Message.ReplyToMessage != nil && update.CallbackQuery.Message.ReplyToMessage.MediaGroupID != "" {
			// Remove all the messages in between
			start_id := update.CallbackQuery.Message.ReplyToMessage.MessageID
			end_id := update.CallbackQuery.Message.MessageID - 1
			for i := start_id; i <= end_id; i++ {
				deleteMediaGroup := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, i)
				sendDeleteMessage(bot, deleteMediaGroup)
			}
		}
	}
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
func removeUpdateMessageRelative(bot *tgbotapi.BotAPI, message *tgbotapi.Message, relativeMessageID int) {
	deleteMessage := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID-relativeMessageID)
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
func sendDeleteMessage(bot *tgbotapi.BotAPI, msg tgbotapi.Chattable) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

// processPriceStr processes a price range string and returns the min and max price.
// Supports formats:
//   - "1000-2000" - price range from 1000 to 2000
//   - "1000+" or "1000-" - minimum price of 1000, no maximum
//   - "-2000" or "2000" - maximum price of 2000, no minimum
//
// Parameters:
//
//	price_range_str: Price range string to process.
//
// Returns:
//
//	minPrice: Minimum price of the range (0 if not specified).
//	maxPrice: Maximum price of the range (0 if not specified).
//	error: Error if the price range string is not valid.
func processPriceStr(price_range_str string) (int, int, error) {
	// Trim spaces
	price_range_str = strings.TrimSpace(price_range_str)

	// Check for "+" suffix (e.g., "1000+")
	if strings.HasSuffix(price_range_str, "+") {
		minPriceStr := strings.TrimSuffix(price_range_str, "+")
		minPrice, err := strconv.Atoi(strings.TrimSpace(minPriceStr))
		if err != nil {
			return 0, 0, errors.New("invalid minimum price format")
		}
		if minPrice < 0 {
			return 0, 0, errors.New("minimum price cannot be negative")
		}
		return minPrice, 0, nil
	}

	// Check for "-" prefix (e.g., "-2000" meaning "up to 2000")
	// But we need to distinguish from negative numbers
	if strings.HasPrefix(price_range_str, "-") && !strings.Contains(price_range_str[1:], "-") {
		// Try to parse as a number first to check if it's negative
		if num, err := strconv.Atoi(price_range_str[1:]); err == nil && num < 0 {
			return 0, 0, errors.New("maximum price must be positive")
		}

		maxPriceStr := strings.TrimPrefix(price_range_str, "-")
		maxPrice, err := strconv.Atoi(strings.TrimSpace(maxPriceStr))
		if err != nil {
			return 0, 0, errors.New("invalid maximum price format")
		}
		if maxPrice <= 0 {
			return 0, 0, errors.New("maximum price must be positive")
		}
		return 0, maxPrice, nil
	}

	// Check for range format (e.g., "1000-2000" or "1000-")
	if strings.Contains(price_range_str, "-") {
		parts := strings.Split(price_range_str, "-")
		if len(parts) != 2 {
			return 0, 0, errors.New("invalid price range format")
		}

		// Process minimum price
		minPrice := 0
		if parts[0] != "" {
			var err error
			minPrice, err = strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, 0, errors.New("invalid minimum price")
			}
			if minPrice < 0 {
				return 0, 0, errors.New("minimum price cannot be negative")
			}
		}

		// Process maximum price
		maxPrice := 0
		if parts[1] != "" {
			var err error
			maxPrice, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return 0, 0, errors.New("invalid maximum price")
			}
			if maxPrice <= 0 {
				return 0, 0, errors.New("maximum price must be positive")
			}
		}

		// Validate range
		if minPrice > 0 && maxPrice > 0 && minPrice > maxPrice {
			return 0, 0, errors.New("minimum price cannot be higher than maximum price")
		}

		return minPrice, maxPrice, nil
	}

	// Check if it's just a number (treat as maximum price)
	if price, err := strconv.Atoi(price_range_str); err == nil {
		if price <= 0 {
			return 0, 0, errors.New("price must be positive")
		}
		return 0, price, nil
	}

	return 0, 0, errors.New("invalid price format")
}
