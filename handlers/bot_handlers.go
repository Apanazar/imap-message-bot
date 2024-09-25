package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var Bot *tgbotapi.BotAPI
var UserID int64

// Telegram Bot Initialization
func InitBot(token string, userID int64) {
	var err error
	Bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	UserID = userID
	log.Println("Bot is started...")
	log.Printf("Authorized on account %s", Bot.Self.UserName)
}

// Processing the start command
func StartHandler(update tgbotapi.Update) {
	if update.Message != nil {
		msg := tgbotapi.NewMessage(UserID, "Listening started...")
		Bot.Send(msg)
	}
}

// Sending a message
func SendMessageToUser(text string) {
	msg := tgbotapi.NewMessage(UserID, text)
	Bot.Send(msg)
}
