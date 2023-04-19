package telegrambot

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"

	"apartment-parser/database"
	"apartment-parser/parser"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserNewSearch struct {
	user_id int64
	state   string
	city    string
}

type City struct {
	Name string
	Code string
}

var cities = []City{
	{
		Name: "Białystok",
		Code: "bialystok",
	},
	{
		Name: "Bydgoszcz",
		Code: "bydgoszcz",
	},
	{
		Name: "Gdańsk",
		Code: "gdansk",
	},
	{
		Name: "Gdynia",
		Code: "gdynia",
	},
	{
		Name: "Katowice",
		Code: "katowice",
	},
	{
		Name: "Kielce",
		Code: "kielce",
	},
	{
		Name: "Kraków",
		Code: "krakow",
	},
	{
		Name: "Lublin",
		Code: "lublin",
	},
	{
		Name: "Łódź",
		Code: "lodz",
	},
	{
		Name: "Poznań",
		Code: "poznan",
	},
	{
		Name: "Radom",
		Code: "radom",
	},
	{
		Name: "Rzeszów",
		Code: "rzeszow",
	},
	{
		Name: "Szczecin",
		Code: "szczecin",
	},
	{
		Name: "Wrocław",
		Code: "wroclaw",
	},
	{
		Name: "Warszawa",
		Code: "warszawa",
	}}

var userStates = make(map[int64]UserNewSearch)

var keyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Searches 🔍"),
	),
)

func handleMessages(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	if update.Message.IsCommand() {
		handleCommand(bot, update)
	}

	if update.Message.Text == "Searches 🔍" {
		handleSearches(bot, update, db)
	}

	// If user exists in userStates
	if userState, ok := userStates[update.Message.Chat.ID]; ok {
		if userState.state == "search|price" {
			handleNewSearchPrice(bot, update, db)
		}
	}

	// Remove last user's message
	deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
	_, err := bot.Send(deleteMessage)
	if err != nil {
		log.Println(err)
	}
}

func handleNewSearchPrice(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	// Check if price is valid
	// Separate price range
	price := update.Message.Text
	priceRange := strings.Split(price, "-")
	if len(priceRange) != 2 {
		msg.Text = "❌ Invalid price range"
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}

		// Remove the last bot's message
		deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID-1)
		_, err = bot.Send(deleteMessage)
		if err != nil {
			log.Println(err)
		}
		// Remove user from userStates
		delete(userStates, update.Message.Chat.ID)
		return
	}

	// Check if price range is valid
	minPrice, err := strconv.Atoi(priceRange[0])
	if err != nil {
		msg.Text = "❌ Invalid price range"
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}

		// Remove the last bot's message
		deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID-1)
		_, err = bot.Send(deleteMessage)
		if err != nil {
			log.Println(err)
		}
		// Remove user from userStates
		delete(userStates, update.Message.Chat.ID)
		return
	}

	maxPrice, err := strconv.Atoi(priceRange[1])
	if err != nil {
		msg.Text = "❌ Invalid price range"
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}

		// Remove the last bot's message
		deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID-1)
		_, err = bot.Send(deleteMessage)
		if err != nil {
			log.Println(err)
		}
		// Remove user from userStates
		delete(userStates, update.Message.Chat.ID)
		return
	}

	// Check if minPrice is lower than maxPrice
	if minPrice > maxPrice {
		msg.Text = "❌ Invalid price range"
		_, err := bot.Send(msg)
		if err != nil {
			log.Println(err)
		}

		// Remove the last bot's message
		deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID-1)
		_, err = bot.Send(deleteMessage)
		if err != nil {
			log.Println(err)
		}
		// Remove user from userStates
		delete(userStates, update.Message.Chat.ID)
		return
	}

	search_term := parser.SearchTerm{
		Location:  userStates[update.Message.Chat.ID].city,
		Price_min: float64(minPrice),
		Price_max: float64(maxPrice),
	}
	url, error := parser.CreateUrl(search_term)
	if error != nil {
		log.Println(error)
	}

	// Add search to database
	err = database.AddSearch(db, update.Message.Chat.ID, url)

	if err != nil {
		log.Println(err)
	}

	// Remove user from userStates
	delete(userStates, update.Message.Chat.ID)

	// Remove the last bot's message
	deleteMessage := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID-1)
	_, err = bot.Send(deleteMessage)
	if err != nil {
		log.Println(err)
	}

	handleSearches(bot, update, db)
}

func handleSearches(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

	searches, err := database.ListSearches(db, update.Message.Chat.ID)
	if err != nil {
		log.Println(err)
	}
	reply_markup := tgbotapi.NewInlineKeyboardMarkup()
	if len(searches) == 0 {
		msg.Text = "❌ You have 0 active searches"
	} else {
		msg.Text = "🔍 You have " + strconv.Itoa(len(searches)) + " active searches"
		for _, search := range searches {
			// New line
			search_info, err := parser.GetSearchInfo(search.URL)
			if err != nil {
				log.Println(err)
			} else {
				// Add button to reply_markup
				reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("💵 "+search_info, "search|"+strconv.Itoa(int(search.ID))),
				))
			}
		}
		msg.ReplyMarkup = reply_markup
	}

	// Add button to create new search
	reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cleanup"),
		tgbotapi.NewInlineKeyboardButtonData("🟢 Create new search", "create"),
	))
	msg.ReplyMarkup = reply_markup
	_, err = bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

func listSearchInfo(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB, search_id int) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.Text)
	search, err := database.GetSearch(db, int64(search_id))
	if err != nil {
		log.Println(err)
	}
	search_info, err := parser.GetSearchInfo(search.URL)
	if err != nil {
		log.Println(err)
	}
	msg.Text = search_info
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cleanup"),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Delete search", "delete|"+strconv.Itoa(int(search.ID))),
		),
	)
	_, err = bot.Send(msg)
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

func ListCities(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.Text)
	msg.Text = "🌇 Choose the city you want to search in"

	reply_markup := tgbotapi.NewInlineKeyboardMarkup()

	for i := 0; i < len(cities); i += 3 {
		row := tgbotapi.NewInlineKeyboardRow()

		for j := 0; j < 3; j++ {
			if i+j < len(cities) {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(cities[i+j].Name, "new-search|"+cities[i+j].Code))
			}
		}
		reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, row)
	}

	reply_markup.InlineKeyboard = append(reply_markup.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cleanup"),
	))

	msg.ReplyMarkup = reply_markup
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)

	}
}

func handleNewSearchCity(bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB, city string) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.Text)
	msg.Text = "Write the price range in PLN (e.g. 1000-2000)"

	// Add user to userStates
	userStates[update.CallbackQuery.Message.Chat.ID] = UserNewSearch{
		user_id: update.CallbackQuery.Message.Chat.ID,
		state:   "search|price",
		city:    city,
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err)

	}
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

	search_db, err := database.OpenSearchesDatabase("searches.db")
	if err != nil {
		log.Println(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Handle updates
	for update := range updates {
		if update.CallbackQuery != nil {
			ProcessCallbackQuery(bot, update, search_db)
		}

		if update.Message != nil {
			handleMessages(bot, update, search_db)
		}
	}
}
