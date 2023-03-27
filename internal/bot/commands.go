package bot

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/grbit/post_bot/internal/model"
	"github.com/grbit/post_bot/internal/repo"

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
	help := "Добро пожаловать домой! Этот бот создан для посткроссинга бёрнеров по всему миру. " +
		"Здесь ты можешь оставить свой адрес для писем, а можешь получить адрес кого-нибудь из друзей. " +
		"Отправь команду /" + model.CmdGiveMeSome + " чтобы получить рандомный адрес получателя. " +
		"Если ты хочешь получить адрес кого-то особенного, то можешь добавить его ник в телеграме после команды.\n\n" +
		"Чтобы добавить свои адрес, ник в инстаграме, ФИО или пожелания для отправителя, напиши соответствующие команды:\n\n" +
		"/" + model.CmdAddAddress + " - добавить адрес\n" +
		"/" + model.CmdAddInstagram + " - добавить инстаграм\n" +
		"/" + model.CmdAddPersonName + " - добавить ФИО\n" +
		"/" + model.CmdAddWishes + " - добавить пожелания (что ты хочешь получить по почте)\n\n" +
		"Если хочешь посмотреть свои данные, то отправь команду /" + model.CmdMyData

	handlers = append(handlers,
		&commandHandler{
			name:       "help",
			desc:       "Что почём",
			handleFunc: stringHandler(help),
		},
		&commandHandler{
			name:       "start",
			desc:       "Что почём",
			handleFunc: stringHandler(help),
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
			name:       model.CmdAddPersonName,
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

		a, err := db.FindByTg(defaultCtx(), s.Telegram)
		if err != nil {
			return nil, xerrors.Errorf("finding by telegram '%s': %w", s.Telegram, err)
		}

		switch {
		case a.Address == "":
			msg.Text = "Ты не добавил адрес. Напиши /" + model.CmdAddAddress + " чтобы добавить."

			return msg, nil
		case !a.Approved:
			msg.Text = "Модераторы ещё не одобрили твои данные. Подожди немного или напиши @rain_aroma." +
				"Мы стараемся давать доступ только проверенным людям."

			return msg, nil
		case update.Message.Text == "/"+model.CmdGiveMeSome:
			msg.Text = "Отлично! Теперь напиши ник в Telegram/Instagram чтобы я мог найти адрес."
			msg.Text += `Или, если хочешь случайный адрес, то просто напиши "ok".`

			return msg, nil
		case update.Message.Text == "":
			msg.Text = "Ты не написал ничего. Я понимаю только текст. Напиши ник в Telegram например."

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
		addr = strings.ReplaceAll(addr, "  ", " ")

		if len(addr) < 10 {
			msg.Text = "Сомневаюсь что это твой адрес, какой-то он короткий. Попробуй ещё раз."

			return msg, nil
		}

		if err := db.AddAddress(defaultCtx(), s.Telegram, addr); err != nil {
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

		if err := db.AddInstagram(defaultCtx(), s.Telegram, inst); err != nil {
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

		if err := db.AddWishes(defaultCtx(), s.Telegram, wishes); err != nil {
			return nil, xerrors.Errorf("adding wishes (req=%q): %w", wishes, err)
		}

		msg.Text = "Пожелания добавлены!"

		return msg, nil
	}
}

func addPersonNameHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.Text == "/"+model.CmdAddPersonName {
			msg.Text = "Давай добавим ФИО! Просто напиши их в следующем сообщении."

			return msg, nil
		}

		name := strings.Replace(update.Message.Text, "/"+model.CmdAddPersonName, "", 1)
		name = strings.TrimSpace(name)

		if err := db.AddPersonName(defaultCtx(), s.Telegram, name); err != nil {
			return nil, xerrors.Errorf("adding name (req=%q): %w", name, err)
		}

		msg.Text = "ФИО добавлены!"

		return msg, nil
	}
}

func myDataHandler() updateHandleFunc {
	return func(update tgbotapi.Update, s *model.State) (tgbotapi.Chattable, error) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		addr, err := db.FindByTg(defaultCtx(), s.Telegram)
		if err != nil {
			return nil, xerrors.Errorf("searching address by (tg=%q): %w", s.Telegram, err)
		}

		if addr.IsEmpty() {
			msg.Text = "Ты ещё не добавил свои данные. Начни с адреса /" + model.CmdAddAddress

			return msg, nil
		}

		msg.Text = "Вот твои данные:\n" + addr.String()
		if addr.Instagram != "" {
			msg.Text += "\nInstagram: " + addr.Instagram + "."
		}

		return msg, nil
	}
}

const defaultCtxTimeout = 10 * time.Second

func defaultCtx() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCtxTimeout)

	go func() {
		<-ctx.Done()
		cancel()
	}()

	return ctx
}
