package telegrambot

import (
	"apartment-parser/database"
	"database/sql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

func ProcessCallbackQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update, search_db *sql.DB) {
	data := strings.Split(update.CallbackQuery.Data, "|")

	if len(data) == 1 {
		switch update.CallbackQuery.Data {
		case "create":
			ListCities(bot, update, search_db)

		case "cleanup":
			deleteMessage := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
			_, err := bot.Send(deleteMessage)
			if err != nil {
				log.Println(err)
			}
		}
	} else if len(data) == 2 {
		switch data[0] {
		case "search":
			search_id, err := strconv.Atoi(data[1])
			if err != nil {
				log.Println(err)
			} else {
				listSearchInfo(bot, update, search_db, search_id)
			}

		case "delete":
			search_id, err := strconv.Atoi(data[1])
			if err != nil {
				log.Println(err)
			}
			err = database.DeleteSearch(search_db, int64(search_id))
			if err != nil {
				log.Println(err)
			}

		case "new-search":
			handleNewSearchCity(bot, update, search_db, data[1])
		}
	}

	// Delete callback query
	deleteMessage := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	_, err := bot.Send(deleteMessage)
	if err != nil {
		log.Println(err)
	}
}
