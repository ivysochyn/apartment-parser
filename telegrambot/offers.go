package telegrambot

import (
	"apartment-parser/database"
	"apartment-parser/parser"

	"database/sql"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Send offer to user with given id.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	offer: Offer to send.
//	UserId: Id of user to send offer to.
func sendOfferToUser(bot *tgbotapi.BotAPI, offer parser.Offer, UserId int64) {
	message_string := offerToText(offer)
	msg := tgbotapi.NewMessage(UserId, message_string)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = false

	var images []interface{} = make([]interface{}, 0)

	// Download all the images and send them to user
	if len(offer.Images) > 0 {
		for _, image_url := range offer.Images {
			var image []byte
			image, err := parser.DownloadImage(image_url)
			if err != nil {
				panic(err)
			}

			images = append(images, tgbotapi.NewInputMediaPhoto(tgbotapi.FileBytes{Name: "image.jpg", Bytes: image}))
		}
	}

	if len(images) > 1 {
		// If there are more than 9 images, send only 9
		if len(images) > 9 {
			images = images[:9]
		}
		// Create a media group
		media_group := tgbotapi.NewMediaGroup(UserId, images)
		media_group_msg, err := bot.SendMediaGroup(media_group)
		if err != nil {
			panic(err)
		}
		// Add replyto message id to the first message
		msg.ReplyToMessageID = media_group_msg[0].MessageID

	} else if len(images) == 1 {
		// Create a photo message
		photo_msg := tgbotapi.NewPhoto(UserId, images[0].(tgbotapi.InputMediaPhoto).Media)
		photo_msg_sent, err := bot.Send(photo_msg)
		if err != nil {
			panic(err)
		}
		// Add replyto message id to the first message
		msg.ReplyToMessageID = photo_msg_sent.MessageID
	}

	reply_markup := tgbotapi.NewInlineKeyboardMarkup()
	reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🗑️ Remove", "remove_msg|"),
	))
	msg.ReplyMarkup = reply_markup
	sendMessage(bot, msg)
}

// Convert offer to text.
//
// Parameters:
//
//	offer: Offer to convert.
//
// Returns:
//
//	Text representation of offer.
func offerToText(offer parser.Offer) string {
	text := "<a href=\"" + offer.Url + "\">" + offer.Title + "</a>\n\n"
	text += "📍 " + offer.Location + "\n"
	text += "💵 " + strconv.Itoa(offer.Price+offer.AdditionalPayment) + " zł"
	if offer.AdditionalPayment != 0 {
		text += " (" + strconv.Itoa(offer.Price) + " + " + strconv.Itoa(offer.AdditionalPayment) + ")"
	}
	text += "\n"

	if offer.Area != "" {
		text += "📐 " + offer.Area + "\n"
	}
	if offer.Rooms != "" {
		text += "🛏 " + offer.Rooms + "\n"
	}
	if offer.Floor != "" {
		text += "🏢 " + offer.Floor + "\n"
	}

	text += "\n📅 Dzisiaj o " + offer.Time + "\n"
	return text
}

// Parse all offers from all searches and send them to users in a loop.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	offers_db: Database with offers.
//	search_db: Database with searches.
func parseOffers(bot *tgbotapi.BotAPI, offers_db *sql.DB, search_db *sql.DB) {
	for {
		searches, err := database.GetAllSearches(search_db)
		if err != nil {
			panic(err)
		}

		for _, search := range searches {
			processAllOffersFromSearch(bot, search, offers_db, search_db)
		}
	}
}

// Parse all offers from given search and send them to user.
//
// Parameters:
//
//		bot: Telegram bot instance.
//		search: Search to parse offers from.
//		offers_db: Database with offers.
//	 search_db: Database with searches.
func processAllOffersFromSearch(bot *tgbotapi.BotAPI, search database.Search, offers_db *sql.DB, search_db *sql.DB) {
	page, err := parser.FetchHTMLPage(search.URL)
	if err != nil {
		log.Printf("Error fetching page: %v", err)
		return
	}

	offers := parser.ParseHtml(page)

	for _, offer := range offers {
		search_exists, err := database.SearchExists(search_db, search)
		if !search_exists {
			return
		}

		exists, err := database.OfferExists(offers_db, offer, search.UserID)
		if err != nil {
			log.Printf("Error checking if offer exists: %v", err)
			return
		}
		if !exists {
			offer = parser.ParseOffer(offer)
			err := database.AddOffer(offers_db, offer, search.UserID)
			if err != nil {
				log.Printf("Error adding offer to database: %v", err)
				return
			}

			// if has 'Dzisiaj' in time and images, send offer
			if len(offer.Images) > 0 {
				sendOfferToUser(bot, offer, search.UserID)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
