package telegrambot

import (
	"apartment-parser/database"
	"apartment-parser/parser"

	"database/sql"
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
	text += "💵 " + offer.Price + "\n"

	if offer.AdditionalPayment != "" {
		text += "📝 " + offer.AdditionalPayment + "\n"
	}
	if offer.Area != "" {
		text += "📐 " + offer.Area + "\n"
	}
	if offer.Rooms != "" {
		text += "🛏 " + offer.Rooms + "\n"
	}
	if offer.Floor != "" {
		text += "🏢 " + offer.Floor + "\n"
	}

	text += "\n📅 " + offer.Time + "\n"
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
			processAllOffersFromSearch(bot, search, offers_db)
			time.Sleep(10 * time.Second)
		}
	}
}

// Parse all offers from given search and send them to user.
//
// Parameters:
//
//	bot: Telegram bot instance.
//	search: Search to parse offers from.
//	offers_db: Database with offers.
func processAllOffersFromSearch(bot *tgbotapi.BotAPI, search database.Search, offers_db *sql.DB) {
	page, err := parser.FetchHTMLPage(search.URL)
	if err != nil {
		panic(err)
	}

	offers := parser.ParseHtml(page)

	for _, offer := range offers {
		exists, err := database.OfferExists(offers_db, offer, search.UserID)
		if err != nil {
			panic(err)
		}
		if !exists {
			offer = parser.ParseOffer(offer)
			err := database.AddOffer(offers_db, offer, search.UserID)
			if err != nil {
				panic(err)
			}

			// TODO: Refactor this
			// if has 'Dzisiaj' in time as a first word, send offer
			if len(offer.Time) > 7 && offer.Time[:7] == "Dzisiaj" {
				sendOfferToUser(bot, offer, search.UserID)
			}
		}
	}
}
