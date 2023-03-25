package db

import (
	"context"

	"github.com/grbit/post_bot/model"

	"golang.org/x/xerrors"
)

func AddAddress(ctx context.Context, tg, address string) error {
	tg = prepareTelegram(tg)

	a, err := globalRepo.getOneAddress(ctx, tg)
	if err != nil {
		return xerrors.Errorf("searching address in DB: %w", err)
	}

	a.Address = address

	if err = globalRepo.upsertAddresses([]*model.Address{a}); err != nil {
		return xerrors.Errorf("upserting address: %w", err)
	}

	cache.Add(a)

	return nil
}

func AddInstagram(ctx context.Context, tg, instagram string) error {
	tg = prepareTelegram(tg)
	instagram = prepareInstagram(instagram)

	a, err := globalRepo.getOneAddress(ctx, tg)
	if err != nil {
		return xerrors.Errorf("searching address in DB: %w", err)
	}

	a.Instagram = instagram

	if err = globalRepo.upsertAddresses([]*model.Address{a}); err != nil {
		return xerrors.Errorf("upserting address: %w", err)
	}

	cache.Add(a)

	return nil
}

func AddWishes(ctx context.Context, tg, wishes string) error {
	tg = prepareTelegram(tg)

	a, err := globalRepo.getOneAddress(ctx, tg)
	if err != nil {
		return xerrors.Errorf("searching address in DB: %w", err)
	}

	a.Wishes = wishes

	if err = globalRepo.upsertAddresses([]*model.Address{a}); err != nil {
		return xerrors.Errorf("upserting address: %w", err)
	}

	cache.Add(a)

	return nil
}

func AddPersonName(ctx context.Context, tg, name string) error {
	tg = prepareTelegram(tg)

	a, err := globalRepo.getOneAddress(ctx, tg)
	if err != nil {
		return xerrors.Errorf("searching address in DB: %w", err)
	}

	a.PersonName = name

	if err = globalRepo.upsertAddresses([]*model.Address{a}); err != nil {
		return xerrors.Errorf("upserting address: %w", err)
	}

	cache.Add(a)

	return nil
}

func (r *Repo) getOneAddress(ctx context.Context, tg string) (*model.Address, error) {
	tg = prepareTelegram(tg)

	aa := []*model.Address{}
	if err := r.WithContext(ctx).Find(&aa, "telegram = ?", tg).Error; err != nil {
		return nil, err
	}

	if len(aa) > 1 {
		return nil, xerrors.Errorf("found to many (%d) addresses", len(aa))
	}

	if len(aa) == 1 {
		return aa[0], nil
	}

	return &model.Address{Telegram: tg}, nil
}
