package telegrambot

import (
    "log"
    "os"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleNewOffer(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "New offer")
    bot.Send(msg)
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
        if update.Message == nil {
            continue
        }

        // ignore any non-commands
        if !update.Message.IsCommand() {
            continue
        }

        switch update.Message.Command() {
            case "new_offer":
                handleNewOffer(bot, update)
        }
    }
}
