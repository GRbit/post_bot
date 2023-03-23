package main

import (
	"runtime/debug"

	"github.com/grbit/post_bot/state"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

type myBot struct {
	*tgbotapi.BotAPI
	Retries  int
	Handlers map[string]*commandHandler
}

func newBot() (*myBot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, xerrors.Errorf("creating Bot API: %w", err)
	}

	bot.Debug = cfg.Debug

	return &myBot{
		BotAPI:  bot,
		Retries: senMsgRetries,
	}, nil
}

func (b *myBot) startBot() error {
	log.Printf("Authorized on account %s", b.Self.UserName)

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

func (b *myBot) handleUpdate(update tgbotapi.Update) error {
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

	// Create a new MessageConfig. We don't have text yet, so we leave it empty.
	var msg tgbotapi.Chattable
	cmd := update.Message.Command()
	chatID := update.Message.Chat.ID
	var err error

	switch {
	case st.SearchCmdWasPrevious && !update.Message.IsCommand():
		msg, err = giveMeHandler()(update, st)
		st.SearchCmdWasPrevious = false
	case update.Message == nil: // ignore any non-Message updates
	case !update.Message.IsCommand(): // ignore any non-command Messages
	default:
		st.SearchCmdWasPrevious = false

		if h, ok := b.Handlers[cmd]; ok && h != nil {
			msg, err = h.handleFunc(update, st)
		}
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

	m, err := b.sendWithretries(msg)
	if err != nil {
		return xerrors.Errorf("sending a message %+v: %w", msg, err)
	}

	if m.Document != nil {
		st.FileIDs[m.Document.FileName] = m.Document.FileID
	}

	return nil
}

func (b *myBot) configure(handlers []*commandHandler) {
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

func (b *myBot) sendWithretries(message tgbotapi.Chattable) (m tgbotapi.Message, err error) {
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
