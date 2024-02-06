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
		if !service.HasSubscribers(fmt.Sprint(update.Message.Chat.ID)) {
			log.Println("no subscribers")
			continue
		}
		mess := update.Message
		message := models.Message{}
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
			continue
		}
		if fileUrl != "" {
			httpClient := &http.Client{Timeout: time.Minute * 60}
			resp, err := httpClient.Get(fileUrl)
			if err != nil {
				log.Println("Error getting file.")
				continue
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
				continue
			}
		}
		s.updates <- message
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
			caption = mess.Sender + ": " + *mess.Caption
		}

		switch {
		case mess.Text != nil:

		case mess.Image != nil && len(mess.Image) > 0:
			m := tgBotAPI.NewPhoto(chatId, tgBotAPI.FileBytes{
				Name:  "Photo",
				Bytes: mess.Image,
			})
			m.Caption = caption
			msg = m
		case mess.Video != nil && len(mess.Video) > 0:
			m := tgBotAPI.NewVideo(chatId, tgBotAPI.FileBytes{
				Name:  "Video",
				Bytes: mess.Video,
			})
			m.Caption = caption
			msg = m
		case mess.Audio != nil && len(mess.Audio) > 0:
			m := tgBotAPI.NewAudio(chatId, tgBotAPI.FileBytes{
				Name:  "Video",
				Bytes: mess.Audio,
			})
			m.Caption = caption
			msg = m
		case mess.Document != nil && len(mess.Document) > 0:
			m := tgBotAPI.NewDocument(chatId, tgBotAPI.FileBytes{
				Name:  "Video",
				Bytes: mess.Document,
			})
			m.Caption = caption
			msg = m
		case mess.Sticker != nil && len(mess.Sticker) > 0:
			msg = tgBotAPI.NewSticker(chatId, tgBotAPI.FileBytes{
				Name:  "Video",
				Bytes: mess.Sticker,
			})
		case mess.Caption != nil:

		}

		rsp, err := s.conn.Send(msg)
		if err != nil {
			log.Println("error sending message: ", err)
		}
		log.Println("message send with id: ", rsp.MessageID)
	}
}