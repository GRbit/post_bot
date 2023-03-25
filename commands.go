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
	// my personal chat id for debug purposes
	myChatID = 805666391
)

type updateHandleFunc func(update tgbotapi.Update, state *model.State) (tgbotapi.Chattable, error)

type commandHandler struct {
	name       string
	desc       string
	handleFunc updateHandleFunc
}

func makeHandlers() (handlers []*commandHandler) {
	handlers = append(handlers,
		&commandHandler{
			name: "help",
			desc: "Что почём",
			handleFunc: stringHandler("Этот бот может выдать тебе почтовых адресов чтобы ты порадовал людей открыточками. " +
				`Отправь команду "/` + model.CmdGiveMeSome + `" чтобы получить адрес. ` +
				"Можешь добавить ник в телефграме после команды если хочешь адрес кого-то особенного. " +
				"Если хочешь добавить свой адрес, то отправь команду /" + model.CmdAddAddress),
		},
		&commandHandler{
			name: "start",
			desc: "Что почём",
			handleFunc: stringHandler(
				"Добро пожаловать домой! Этот бот создан для того, чтобы ты не заблудился в волшебном мире Холодка. " +
					`Отправь команду "/` + model.CmdGiveMeSome + `" чтобы получить адрес. ` +
					"Можешь добавить ник в телефграме после команды если хочешь адрес кого-то особенного.\n\n" +
					"Чтобы добавить свои адрес, ник в инстаграме, ФИО или пожелания для отправителя, " +
					"напиши соответствующие команды:\n\n" +
					"/" + model.CmdAddAddress + " - добавить адрес\n" +
					"/" + model.CmdAddInstagram + " - добавить инстаграм\n" +
					"/" + model.CmdAddName + " - добавить ФИО\n" +
					"/" + model.CmdAddWishes + " - добавить пожелания\n\n" +
					"Если хочешь посмотреть свои данные, то отправь команду /" + model.CmdMyData),
		},
		&commandHandler{
			name:       model.CmdGiveMeSome,
			desc:       "Взять адрес",
			handleFunc: giveMeHandler(),
		},
		&commandHandler{
			name:       model.CmdAddAddress,
			desc:       "Добавить адрес",
			handleFunc: addAddressHandler(),
		},
		&commandHandler{
			name:       model.CmdAddInstagram,
			desc:       "Добавить Instagram",
			handleFunc: addInstagramHandler(),
		},
		&commandHandler{
			name:       model.CmdAddWishes,
			desc:       "Добавить пожелания",
			handleFunc: addWishesHandler(),
		},
		&commandHandler{
			name:       model.CmdAddName,
			desc:       "Добавить ФИО",
			handleFunc: addPersonNameHandler(),
		},
		&commandHandler{
			name:       model.CmdMyData,
			desc:       "Посмотреть свои данные",
			handleFunc: myDataHandler(),
		},
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

		if update.Message.Text == "/"+model.CmdGiveMeSome {
			msg.Text = "Отлично! Теперь напиши ник в Telegram/Instagram чтобы я мог найти адрес."
			msg.Text += `Или, если хочешь случайный адрес, то просто напиши "ok".`

			return msg, nil
		}

		searchReq := strings.Replace(update.Message.Text, "/"+model.CmdGiveMeSome, "", 1)
		searchReq = strings.TrimSpace(searchReq)

		if strings.EqualFold(searchReq, "ok") || strings.EqualFold(searchReq, "ок") {
			msg.Text = "Корейский рандом сказал дать тебе это:\n" + db.Random().String()
			s.GivenAddressesCtr++

			return msg, nil
		}

		res, err := db.Search(searchReq)
		if err != nil {
			return nil, xerrors.Errorf("searching (req=%q): %w", searchReq, err)
		}

		s.GivenAddressesCtr += len(res)

		switch len(res) {
		case 0:
			msg.Text = "Я ничего не нашёл =("
		case 1:
			msg.Text = "Я нашёл!\n" + res[0]
		default:
			msg.Text = "Ого, да тут много адресов...\n"

			for i, s := range res {
				msg.Text += "Номер " + strconv.Itoa(i) + ":\n" + s
			}
		}

		return msg, nil
	}
}

func addAddressHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.Text == "/"+model.CmdAddAddress {
			msg.Text = "Давай добавим адрес! Просто напиши его в следующем сообщении."

			return msg, nil
		}

		addr := strings.Replace(update.Message.Text, "/"+model.CmdAddAddress, "", 1)
		addr = strings.TrimSpace(addr)

		if err := db.AddAddress(s.Telegram, addr); err != nil {
			return nil, xerrors.Errorf("adding address (req=%q): %w", addr, err)
		}

		msg.Text = "Адрес добавлен!"

		return msg, nil
	}
}

func addInstagramHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.Text == "/"+model.CmdAddInstagram {
			msg.Text = "Давай добавим Instagram! Просто напиши свой ник в следующем сообщении."

			return msg, nil
		}

		inst := strings.Replace(update.Message.Text, "/"+model.CmdAddInstagram, "", 1)
		inst = strings.TrimSpace(inst)

		if err := db.AddInstagram(s.Telegram, inst); err != nil {
			return nil, xerrors.Errorf("adding instagram (req=%q): %w", inst, err)
		}

		msg.Text = "Instagram добавлен!"

		return msg, nil
	}
}

func addWishesHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.Text == "/"+model.CmdAddWishes {
			msg.Text = "Давай добавим пожелания! Просто напиши их в следующем сообщении."

			return msg, nil
		}

		wishes := strings.Replace(update.Message.Text, "/"+model.CmdAddWishes, "", 1)
		wishes = strings.TrimSpace(wishes)

		if err := db.AddWishes(s.Telegram, wishes); err != nil {
			return nil, xerrors.Errorf("adding wishes (req=%q): %w", wishes, err)
		}

		msg.Text = "Пожелания добавлены!"

		return msg, nil
	}
}

func addPersonNameHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.Text == "/"+model.CmdAddName {
			msg.Text = "Давай добавим ФИО! Просто напиши их в следующем сообщении."

			return msg, nil
		}

		name := strings.Replace(update.Message.Text, "/"+model.CmdAddName, "", 1)
		name = strings.TrimSpace(name)

		if err := db.AddPersonName(s.Telegram, name); err != nil {
			return nil, xerrors.Errorf("adding name (req=%q): %w", name, err)
		}

		msg.Text = "ФИО добавлены!"

		return msg, nil
	}
}

func myDataHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		addr, err := db.FindByTg(s.Telegram)
		if err != nil {
			return nil, xerrors.Errorf("searching address by (tg=%q): %w", s.Telegram, err)
		}

		if update.Message.Text == "/"+model.CmdMyData {
			msg.Text = "Вот твои данные:\n" + addr.String()
			msg.Text += "\nInstagram: " + addr.Instagram

			return msg, nil
		}

		return nil, nil
	}
}
