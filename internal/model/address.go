package model

type Address struct {
	Base

	Telegram   string
	Instagram  string
	PersonName string
	Address    string
	Wishes     string

	Approved bool

	// Not added yet
	Email string
	Phone string
}

func (p Address) String() string {
	msg := ""
	if p.PersonName != "" {
		msg += p.PersonName + "."
	}

	if p.Address != "" {
		msg += " Адрес: " + p.Address + "."
	}

	if p.Wishes != "" {
		msg += " Пожелания: " + p.Wishes + "."
	}

	return msg
}

func (p Address) IsEmpty() bool {
	if p.Instagram == "" && p.PersonName == "" && p.Address == "" && p.Wishes == "" {
		return true
	}

	return false
}
