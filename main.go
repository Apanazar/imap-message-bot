package main

import (
	"fmt"
	"imap-go/handlers"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("config/config.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	imapDomain := os.Getenv("IMAP_DOMAIN")
	imapPortStr := os.Getenv("IMAP_PORT")
	imapPassword := os.Getenv("IMAP_PASSWORD")
	email := os.Getenv("IMAP_MAIL")

	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramUserID := os.Getenv("TELEGRAM_USER_ID")
	userID, _ := strconv.ParseInt(telegramUserID, 10, 64)

	handlers.InitBot(telegramToken, userID)

	imapPort, err := strconv.Atoi(imapPortStr)
	if err != nil {
		log.Fatalf("Error converting IMAP_PORT to integer: %v", err)
	}
	imapServer := fmt.Sprintf("%s:%d", imapDomain, imapPort)

	c, err := handlers.ConnectToIMAP(imapServer)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	log.Println("Connected to IMAP server")
	if err := handlers.Login(c, email, imapPassword); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in to IMAP server")

	go handlers.FetchMail(c, email, imapPassword)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := handlers.Bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message != nil && update.Message.Text == "/start" {
			handlers.StartHandler(update)
		}
	}
}
