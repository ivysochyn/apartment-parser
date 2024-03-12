# ApartiBot - Your personal apartment search assistant

ApartiBot is a Telegram bot that aggregates apartment listings from different sources and sends them to you.
The bot is currently in development and only supports the following sources:
* [OLX](https://www.olx.pl/)
* [Otodom](https://www.otodom.pl/)

## How to use

To start using the bot, you need to have a Telegram Bot Token.
The token is fetched from the environment variable `TELEGRAM_APITOKEN`:
```bash
TELEGRAM_APITOKEN=your_token_here go run apartment-parser
```
