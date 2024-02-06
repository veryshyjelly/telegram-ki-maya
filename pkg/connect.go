package pkg

import tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func Connect(apiToken string, debug bool) *tgBotAPI.BotAPI {
	bot, err := tgBotAPI.NewBotAPI(apiToken)
	if err != nil {
		panic(err)
	}
	bot.Debug = debug
	return bot
}