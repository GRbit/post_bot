package db

import (
	"context"
	"math/rand"

	"github.com/grbit/post_bot/internal/model"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
	"golang.org/x/xerrors"
)

func FindByTg(ctx context.Context, tg string) (*model.Address, error) {
	addr, err := globalRepo.getOneAddress(ctx, tg)
	if err != nil {
		return nil, xerrors.Errorf("searching in DB: %w", err)
	}

	return addr, nil
}

func Search(req string) ([]string, error) {
	if req == "" {
		return nil, nil
	}

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
		Interface("persons", cache.Persons).
		Msg("Random call")

	log.Debug().
		Interface("len(persons)", len(cache.Persons)).
		Msg("Random call")

	tgs := maps.Keys(cache.Persons)
	k := rand.Intn(len(tgs))
	log.Debug().Str("random key", tgs[k]).Send()

	tg := tgs[k]
	log.Debug().Str("random telegram", tg).Send()

	return cache.Persons[tg]
}
