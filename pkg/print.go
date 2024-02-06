package pkg

import (
	"fmt"
	tgBotAPI "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"time"
)

var (
	colorReset = "\033[0m"
	bold       = "\u001b[1m"

	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	//colorWhite  = "\033[37m"
)

func PrintUser(u *tgBotAPI.User) string {
	var res = colorRed
	if u.IsBot {
		res += fmt.Sprint(" BOT ")
	}
	res += fmt.Sprint(u.FirstName, " ")
	if u.LastName != "" {
		res += fmt.Sprint(u.LastName, " ")
	}
	if u.UserName != "" {
		res += fmt.Sprint(u.UserName, " ")
	}
	res += colorReset
	return res
}

func PrintUpdate(u *tgBotAPI.Update) string {
	var res string

	if u.Message != nil {
		res += PrintMessage(u.Message)
	}
	if u.CallbackQuery != nil {
		res += PrintCallback(u.CallbackQuery)
	}

	return res
}

func PrintChat(c *tgBotAPI.Chat) string {
	var res string
	res += fmt.Sprint(c.Title, " ")
	return res
}

func PrintMessage(m *tgBotAPI.Message) string {
	var res string
	// “private”, “group”, “supergroup” or “channel”
	res += bold
	if m.Chat.Type == "group" {
		res += colorPurple
	} else if m.Chat.Type == "supergroup" {
		res += colorCyan
	} else if m.Chat.Type == "private" {
		res += colorBlue
	} else {
		res += colorYellow
	}
	res += fmt.Sprint(m.MessageID, " ")
	res += fmt.Sprint("[", time.Unix(int64(m.Date), 0).String()[:19], "] ")
	res += PrintChat(m.Chat)
	res += PrintUser(m.From)
	if m.ForwardFromMessageID != 0 {
		res += fmt.Sprint("[forwarded from ", PrintChat(m.ForwardFromChat), "] ")
	}
	res += bold + colorBlue + "»»» " + colorGreen
	if m.ReplyToMessage != nil {
		res += fmt.Sprint("[reply to ", m.ReplyToMessage.MessageID, "] ")
	}

	if m.Animation != nil {
		res += fmt.Sprint("[animation ", m.Animation.FileName, " size=", m.Animation.FileSize, "] ")
	} else if m.Audio != nil {
		res += fmt.Sprint("[audio ", m.Audio.FileName, " size=", m.Document.FileSize, "] ")
	} else if m.Document != nil {
		res += fmt.Sprint("[document ", m.Document.FileName, " size=", m.Document.FileSize, "] ")
	} else if len(m.Photo) != 0 {
		res += fmt.Sprint("[photo] ")
	} else if m.Sticker != nil {
		res += fmt.Sprint("[sticker ", m.Sticker.Emoji, " size=", m.Sticker.FileSize, "] ")
	} else if m.Video != nil {
		res += fmt.Sprint("[video] ")
	} else if m.Voice != nil {
		res += fmt.Sprint("[voice size=", m.Voice.FileSize, "] ")
	}

	if len(m.NewChatMembers) != 0 {
		res += fmt.Sprint("[new Members: ")
		for _, v := range m.NewChatMembers {
			res += PrintUser(&v) + " "
		}
		res += "] "
	}

	if m.GroupChatCreated {
		res += "[group Chat Created] "
	}
	if m.SuperGroupChatCreated {
		res += "[super Group Chat Created] "
	}
	if m.ChannelChatCreated {
		res += "[channel Chat created] "
	}
	if len(m.Entities) != 0 {
		for _, v := range m.Entities {
			res += PrintEntity(&v, m.Text)
		}
	}
	if len(m.CaptionEntities) != 0 {
		for _, v := range m.CaptionEntities {
			res += PrintEntity(&v, m.Caption)
		}
	}

	res += fmt.Sprint(m.Text, m.Caption) + colorReset
	return res
}

func PrintEntity(e *tgBotAPI.MessageEntity, text string) string {
	var res string
	if e.Type == "mention" {
		res += "[mention] "
	} else {
		res += "[" + e.Type + " " + text[e.Offset+1:e.Offset+e.Length] + "] "
	}
	return res
}

func PrintCallback(b *tgBotAPI.CallbackQuery) string {
	var res string

	res += bold
	res += colorYellow
	res += fmt.Sprint(b.Message.MessageID, " ")
	res += PrintChat(b.Message.Chat)
	res += PrintUser(b.From) + " "
	res += bold + colorBlue + "»»» " + colorGreen
	res += b.Data
	res += colorReset

	return res
}