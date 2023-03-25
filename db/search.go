package db

import (
	"math/rand"
	"strings"

	"github.com/grbit/post_bot/model"

	"github.com/rs/zerolog/log"
	"golang.org/x/xerrors"
)

func FindByTg(tg string) (*model.Address, error) {
	addr, err := globalRepo.getOneAddress(tg)
	if err != nil {
		return nil, xerrors.Errorf("searching in DB: %w", err)
	}

	return addr, nil
}

func Search(req string) ([]string, error) {
	req = strings.ToLower(req)

	bb, err := globalRepo.searchAddress(req)
	if err != nil {
		return nil, xerrors.Errorf("searching in DB: %w", err)
	}

	var res []string
	for _, b := range bb {
		res = append(res, b.String())
	}

	return res, nil
}

func Random() *model.Address {
	cache.RLock()
	defer cache.RUnlock()

	log.Trace().
		Interface("telegrams", cache.Tgs).
		Interface("persons", cache.Persons).
		Msg("Random call")

	log.Debug().
		Interface("len(telegrams)", len(cache.Tgs)).
		Interface("len(persons)", len(cache.Persons)).
		Msg("Random call")

	k := rand.Intn(len(cache.Tgs))
	log.Debug().Str("random key", cache.Tgs[k]).Send()
	tg := cache.Tgs[k]
	log.Debug().Str("random telegram", tg).Send()

	return cache.Persons[tg]
}

func (r *Repo) searchAddress(req string) ([]*model.Address, error) {
	phone := preparePhone(req)
	tg := prepareTelegram(req)

	aa := []*model.Address{}
	if err := r.
		Find(&aa,
			"phone = ? OR email = ? OR telegram = ? OR instagram = ?",
			phone, req, tg, req).Error; err != nil {
		return nil, err
	}

	return aa, nil
}

func preparePhone(b string) string {
	l := len("79998148871")

	b = strings.ReplaceAll(b, `"`, ``)
	b = strings.ReplaceAll(b, " ", "")
	b = strings.ReplaceAll(b, "(", "")
	b = strings.ReplaceAll(b, ")", "")
	b = strings.ReplaceAll(b, "+", "")
	b = strings.ReplaceAll(b, "-", "")

	if len(b) == l && strings.HasPrefix(b, "89") {
		b = strings.Replace(b, "8", "7", 1)
	} else if len(b) == l-1 && strings.HasPrefix(b, "9") {
		b = "7" + b
	}

	return b
}

func prepareTelegram(b string) string {
	b = strings.ReplaceAll(b, "@", "")
	b = strings.ReplaceAll(b, " ", "")
	b = strings.ReplaceAll(b, "http://t.me", "")
	b = strings.ReplaceAll(b, "https://t.me", "")
	b = strings.ReplaceAll(b, "http://www.t.me", "")
	b = strings.ReplaceAll(b, "https://www.t.me", "")
	b = strings.Trim(b, "/")

	return b
}

func prepareInstagram(b string) string {
	b = strings.ReplaceAll(b, "@", "")
	b = strings.ReplaceAll(b, " ", "")
	b = strings.ReplaceAll(b, "http://www.instagram.com", "")
	b = strings.ReplaceAll(b, "https://www.instagram.com", "")
	b = strings.ReplaceAll(b, "http://instagram.com", "")
	b = strings.ReplaceAll(b, "https://instagram.com", "")
	b = strings.Trim(b, "/")

	return b
}
