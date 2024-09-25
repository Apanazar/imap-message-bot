package handlers

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

// Connecting to the IMAP server
func ConnectToIMAP(server string) (*client.Client, error) {
	c, err := client.DialTLS(server, &tls.Config{})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Authorization on the mail server
func Login(c *client.Client, email, password string) error {
	return c.Login(email, password)
}

// Receiving mail
func FetchMail(c *client.Client, email, password string) {
	_, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, err := c.Select("INBOX", false)
		if err != nil {
			log.Printf("Error selecting INBOX: %v", err)
			continue
		}

		FetchMessages(c)
		time.Sleep(5 * time.Second)
	}
}

// Receiving new unread messages
func FetchMessages(c *client.Client) {
	seqSet := new(imap.SeqSet)

	searchCriteria := imap.NewSearchCriteria()
	searchCriteria.WithoutFlags = []string{imap.SeenFlag}
	ids, err := c.Search(searchCriteria)
	if err != nil {
		log.Printf("Error searching for new messages: %v", err)
		return
	}

	if len(ids) == 0 {
		return
	}

	seqSet.AddNum(ids...)
	section := &imap.BodySectionName{}
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}, messages)
	}()

	message_text := ""
	for msg := range messages {
		log.Printf("New message received: %v", msg.Envelope.Subject)
		msgBody := msg.GetBody(section)
		if msgBody != nil {
			reader, err := mail.CreateReader(msgBody)
			if err != nil {
				log.Printf("Error creating message reader: %v", err)
				continue
			}

			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error reading message part: %v", err)
					break
				}

				body, _ := io.ReadAll(part.Body)
				re := regexp.MustCompile("<[^>]*>")
				plainText := re.ReplaceAllString(string(body), "")
				plainText = strings.Trim(plainText, "\n")
				message_text = fmt.Sprintf("New message: %s\nText: %s", msg.Envelope.Subject, plainText)
			}

			SendMessageToUser(message_text)
		}

		item := imap.FormatFlagsOp(imap.AddFlags, true)
		flags := []interface{}{imap.SeenFlag}
		if err := c.Store(seqSet, item, flags, nil); err != nil {
			log.Printf("Error marking message as seen: %v", err)
		}
	}

	if err := <-done; err != nil {
		log.Printf("Error fetching messages: %v", err)
	}
}
