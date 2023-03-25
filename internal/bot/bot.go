package bot

import (
	"runtime/debug"

	"github.com/grbit/post_bot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
	"github.com/grbit/post_bot/internal/state"
	"github.com/grbit/post_bot/internal/model"
)

const (
	sendMsgRetries = 10
)

type MyBot struct {
	*tgbotapi.BotAPI
	Retries  int
	Handlers map[string]*commandHandler
}

func New(cfg config.Values) (*MyBot, error) {
	tgBot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, xerrors.Errorf("creating Bot API: %w", err)
	}

	tgBot.Debug = cfg.Debug

	b := &MyBot{
		BotAPI:  tgBot,
		Retries: sendMsgRetries,
	}

	b.configure(makeHandlers())

	return b, nil
}

func (b *MyBot) StartBot() error {
	log.Info().Msgf("Authorized on account %s", b.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.GetUpdatesChan(u)

	for update := range updates {
		log.Debug().
			Interface("update", update).
			Interface("update.Message", update.Message).
			Msg("log")

		go func(update tgbotapi.Update) {
			if err := b.handleUpdate(update); err != nil {
				log.Error().
					Interface("update", update).
					Err(err).
					Msg("handling update")
			}
		}(update)
	}

	return xerrors.Errorf("something went wrong, out of cycle")
}

func (b *MyBot) handleUpdate(update tgbotapi.Update) error {
	defer func() {
		if rec := recover(); rec != nil {
			switch err := rec.(type) {
			case error:
				s := string(debug.Stack())

				log.Error().
					Err(err).
					Str("stack", s).
					Msgf("recovering panic: %s", s)
			default:
				log.Debug().Interface("recover", rec).Msgf("some unknown not nil recover: %+v", rec)
			}
		}
	}()

	if update.Message == nil { // ignore any non-Message updates
		return nil
	}

	st := state.Get(update.Message.Chat.ID)
	st.Telegram = update.Message.From.UserName

	// Create a new MessageConfig. We don't have text yet, so we leave it empty.
	var msg tgbotapi.Chattable
	cmd := update.Message.Command()
	chatID := update.Message.Chat.ID
	var err error

	switch {
	case !update.Message.IsCommand():

		switch st.PreviousCmd {
		case model.CmdGiveMeSome:
			msg, err = giveMeHandler()(update, st)
		case model.CmdAddAddress:
			msg, err = addAddressHandler()(update, st)
		case model.CmdAddInstagram:
			msg, err = addInstagramHandler()(update, st)
		case model.CmdAddWishes:
			msg, err = addWishesHandler()(update, st)
		case model.CmdAddPersonName:
			msg, err = addPersonNameHandler()(update, st)
		default: // ignore any other non-command Messages
		}

	case update.Message == nil: // ignore any non-Message updates
	default:
		if h, ok := b.Handlers[cmd]; ok && h != nil {
			msg, err = h.handleFunc(update, st)
		}
	}

	if update.Message.IsCommand() {
		st.PreviousCmd = cmd
	} else {
		st.PreviousCmd = ""
	}

	if err != nil {
		log.Error().Err(err).Msg("handler error")
		msg = tgbotapi.NewMessage(chatID,
			"Тут какая-то ошибка произошла... Напишите прогеру t.me/grbit, пусть починит.")
	}

	if msg == nil {
		msg = tgbotapi.NewMessage(chatID,
			"Ну я хз. Не понимаю что от меня хотят. Может `/help`?")
	}

	m, err := b.sendWithRetries(msg)
	if err != nil {
		return xerrors.Errorf("sending a message %+v: %w", msg, err)
	}

	if m.Document != nil {
		st.FileIDs[m.Document.FileName] = m.Document.FileID
	}

	return nil
}

func (b *MyBot) configure(handlers []*commandHandler) {
	b.Handlers = make(map[string]*commandHandler)
	cmds := make([]tgbotapi.BotCommand, len(handlers))
	for i, h := range handlers {
		cmds[i] = tgbotapi.BotCommand{Command: h.name, Description: h.desc}
		b.Handlers[h.name] = h
	}

	commands := tgbotapi.NewSetMyCommands(cmds...)

	msg, err := b.Send(commands)
	log.Info().Err(err).Interface("msg", msg).Msg("set cmds")

	cc, err := b.BotAPI.GetMyCommands()
	log.Info().Err(err).Interface("commands", cc).Msg("got cmds")
}

func (b *MyBot) sendWithRetries(message tgbotapi.Chattable) (m tgbotapi.Message, err error) {
	for i := 1; i < b.Retries; i++ {
		m, err = b.Send(message)
		if err != nil {
			log.Warn().Err(err).Interface("message_config", message).Msg("failed to send a message")
		} else {
			return m, nil
		}
	}

	return m, xerrors.Errorf("sending a message %+v: %w", message, err)
}
