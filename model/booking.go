package model

type Address struct {
	Base

	Telegram   string
	Instagram  string
	PersonName string
	Address    string
	Wishes     string
}

func (p Address) String() string {
	msg := p.PersonName + ". Адрес: " + p.Address
	if p.Wishes != "" {
		msg += ". Пожелания: " + p.Wishes
	}

	return msg
}
