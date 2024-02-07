package subscription

import (
	"fmt"
	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"strconv"
	"telegram-ki-maya/models"
	"telegram-ki-maya/pkg"
	"time"
)

// Server is the abstraction for whatsapp or telegram etc.
// this interface handles all the updates that comes from server
// and should handle all the updates that needs to be sent to the server
type Server interface {
	Update() chan models.Message
	Listen(service Service)
	Serve()
}

type server struct {
	updates chan models.Message
	conn    *tgBotAPI.BotAPI
}

func NewServer(conn *tgBotAPI.BotAPI) Server {
	return &server{
		updates: make(chan models.Message, 100),
		conn:    conn,
	}
}

func (s *server) Update() chan models.Message {
	return s.updates
}

func (s *server) Listen(service Service) {
	updatesConfig := tgBotAPI.NewUpdate(0)
	updatesConfig.Timeout = 30
	updatesChan := s.conn.GetUpdatesChan(updatesConfig)
	for update := range updatesChan {
		fmt.Println(pkg.PrintUpdate(&update))
		// downloading should only occur when there is someone subscribed to this chat
		if update.Message == nil {
			continue
		}
		if update.Message.Text == "/id" {
			msg := tgBotAPI.NewMessage(update.Message.Chat.ID, fmt.Sprint(update.Message.Chat.ID))
			_, err := s.conn.Send(msg)
			if err != nil {
				log.Println("error sending message: ", err)
			}
			continue
		}
		if !service.HasSubscribers(fmt.Sprint(update.Message.Chat.ID)) {
			log.Println("no subscribers")
			continue
		}
		mess := update.Message
		go func(mess *tgBotAPI.Message) {
			message := models.Message{}
			message.ChatId = fmt.Sprint(mess.Chat.ID)
			message.Sender = mess.From.UserName
			message.Caption = &mess.Caption
			if mess.ReplyToMessage != nil {
				quotedText := "*" + mess.ReplyToMessage.From.UserName + "*: " + mess.ReplyToMessage.Text
				if mess.ReplyToMessage.From.ID == s.conn.Self.ID {
					quotedText = mess.ReplyToMessage.Text
				}
				message.QuotedText = &quotedText
			}
			var fileUrl string
			var err error
			if len(mess.Photo) != 0 {
				fileUrl, err = s.conn.GetFileDirectURL(mess.Photo[len(mess.Photo)-1].FileID)
			} else if mess.Video != nil {
				fileUrl, err = s.conn.GetFileDirectURL(mess.Video.FileID)
			} else if mess.Document != nil {
				fileUrl, err = s.conn.GetFileDirectURL(mess.Document.FileID)
			} else if mess.Audio != nil {
				fileUrl, err = s.conn.GetFileDirectURL(mess.Audio.FileID)
			} else if mess.Sticker != nil {
				fileUrl, err = s.conn.GetFileDirectURL(mess.Sticker.FileID)
			} else if mess.Text != "" {
				message.Text = &mess.Text
			}
			if err != nil {
				log.Println("Error getting direct url.")
				return
			}
			if fileUrl != "" {
				httpClient := &http.Client{Timeout: time.Minute * 60}
				resp, err := httpClient.Get(fileUrl)
				if err != nil {
					log.Println("Error getting file.")
					return
				}
				if len(mess.Photo) != 0 {
					message.Image, err = io.ReadAll(resp.Body)
				} else if mess.Video != nil {
					message.Video, err = io.ReadAll(resp.Body)
				} else if mess.Document != nil {
					message.Document, err = io.ReadAll(resp.Body)
				} else if mess.Audio != nil {
					message.Audio, err = io.ReadAll(resp.Body)
				} else if mess.Sticker != nil {
					message.Sticker, err = io.ReadAll(resp.Body)
				}
				if err != nil {
					fmt.Println("Error while downloading data")
					return
				}
			}
			service.SendToClients() <- message
		}(mess)
	}
}

// Serve methods sends the message to the server
func (s *server) Serve() {
	for mess := range s.updates {
		var msg tgBotAPI.Chattable
		var chatId, err = strconv.ParseInt(mess.ChatId, 10, 64)
		if err != nil {
			continue
		}

		var caption string
		if mess.Caption != nil {
			caption = "[" + mess.Sender + "](tg://user?id=6972063311): " + *mess.Caption
		}

		text := ""

		if mess.QuotedText != nil {
			text += "> " + *mess.QuotedText + "\n"
		}

		text += "[" + mess.Sender + "](tg://user?id=6972063311): "

		switch {
		case mess.Text != nil:
			text += *mess.Text
			m := tgBotAPI.NewMessage(chatId, text)
			m.ParseMode = "MarkdownV2"
			msg = m
		case mess.Image != nil && len(mess.Image) > 0:
			m := tgBotAPI.NewPhoto(chatId, tgBotAPI.FileBytes{
				Name:  "Photo",
				Bytes: mess.Image,
			})
			m.ParseMode = "MarkdownV2"
			if caption != "" {
				m.Caption = caption
			} else {
				m.Caption = text + "Send a photo."
			}
			msg = m
		case mess.Video != nil && len(mess.Video) > 0:
			m := tgBotAPI.NewVideo(chatId, tgBotAPI.FileBytes{
				Name:  "Video",
				Bytes: mess.Video,
			})
			m.ParseMode = "MarkdownV2"
			if caption != "" {
				m.Caption = caption
			} else {
				m.Caption = text + "Send a video."
			}
			msg = m
		case mess.Audio != nil && len(mess.Audio) > 0:
			m := tgBotAPI.NewAudio(chatId, tgBotAPI.FileBytes{
				Name:  "Audio",
				Bytes: mess.Audio,
			})
			m.ParseMode = "MarkdownV2"
			if caption != "" {
				m.Caption = caption
			} else {
				m.Caption = text + "Send a audio."
			}
			msg = m
		case mess.Document != nil && len(mess.Document) > 0:
			var name string
			if mess.Filename != nil {
				name = *mess.Filename
			} else {
				name = "Document"
			}
			m := tgBotAPI.NewDocument(chatId, tgBotAPI.FileBytes{
				Name:  name,
				Bytes: mess.Document,
			})
			m.ParseMode = "MarkdownV2"
			if caption != "" {
				m.Caption = caption
			} else {
				m.Caption = text + "Send a document."
			}
			msg = m
		case mess.Sticker != nil && len(mess.Sticker) > 0:
			msg = tgBotAPI.NewSticker(chatId, tgBotAPI.FileBytes{
				Name:  "Sticker",
				Bytes: mess.Sticker,
			})
		case mess.Caption != nil:
			m := tgBotAPI.NewMessage(chatId, caption)
			m.ParseMode = "MarkdownV2"
			msg = m
		default:
			return
		}

		rsp, err := s.conn.Send(msg)
		if err != nil {
			log.Println("error sending message: ", err)
		} else {
			log.Println("message send with id: ", rsp.MessageID)
		}
	}
}