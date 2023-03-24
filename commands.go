package main

import (
	"strconv"
	"strings"

	"github.com/grbit/post_bot/db"
	"github.com/grbit/post_bot/model"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/xerrors"
)

const (
	giveMeCmdName = "give_me_some"

	// my personal chat id for debug purposes
	myChatID = 805666391
)

type updateHandleFunc func(update tgbotapi.Update, state *model.State) (tgbotapi.Chattable, error)

type commandHandler struct {
	name       string
	desc       string
	handleFunc updateHandleFunc
}

func newCommandHandler(name, desc string, handleFunc updateHandleFunc) *commandHandler {
	return &commandHandler{name: name, desc: desc, handleFunc: handleFunc}
}

func makeHandlers() (handlers []*commandHandler) {
	handlers = append(handlers,
		newCommandHandler("help", "Что почём",
			stringHandler("Этот бот может выдать тебе почтовых адресов чтобы ты порадовал людей открыточками. "+
				`Отправь команду "/`+giveMeCmdName+`" чтобы получить адрес. `+
				"Можешь добавить ник в телефграме после команды если хочешь адрес кого-то особенного.")),
		newCommandHandler("start", "Что почём",
			stringHandler(
				"Добро пожаловать домой! Этот бот создан для того, чтобы ты не заблудился в волшебном мире Холодка. "+
					`Отправь команду "/`+giveMeCmdName+`" чтобы получить адрес. `+
					"Можешь добавить ник в телефграме после команды если хочешь адрес кого-то особенного.")),
		newCommandHandler(giveMeCmdName, "Взять адрес", giveMeHandler()),
	)

	return handlers
}

func stringHandler(s string) updateHandleFunc {
	return func(upd tgbotapi.Update, _ *model.State) (tgbotapi.Chattable, error) {
		return tgbotapi.NewMessage(upd.Message.Chat.ID, s), nil
	}
}

func giveMeHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.Text == "/"+giveMeCmdName {
			s.SearchCmdWasPrevious = true

			msg.Text = "Отлично! Теперь напиши ник в Telegram/Instagram чтобы я мог найти адрес."
			msg.Text += `Или, если хочешь случайный адрес, то просто напиши "ok".`

			return msg, nil
		}

		searchReq := strings.Replace(update.Message.Text, "/"+giveMeCmdName, "", 1)
		searchReq = strings.TrimSpace(searchReq)

		if strings.EqualFold(searchReq, "ok") || strings.EqualFold(searchReq, "ок") {
			msg.Text = "Корейский рандом сказал дать тебе это:\n" + db.Random().String()

			return msg, nil
		}

		res, err := db.Search(searchReq)
		if err != nil {
			return nil, xerrors.Errorf("searching (req=%q): %w", searchReq, err)
		}

		switch len(res) {
		case 0:
			msg.Text = "Я ничего не нашёл =("
		case 1:
			msg.Text = "Я нашёл!\n" + res[0]
		default:
			msg.Text = "Ого, да тут вас много...\n"

			for i, s := range res {
				msg.Text += "Номер " + strconv.Itoa(i) + ":\n" + s
			}
		}

		return msg, nil
	}
}
