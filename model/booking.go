package model

import (
	"fmt"
	"regexp"
	"strconv"
)

type Address struct {
	Base

	Telegram   string
	Instagram  string
	PersonName string
	Address    string
	Wishes     string
}

func (p Address) String() string {
	return p.Address
}

var bRe = regexp.MustCompile(`(К([0-9][0-9]?)|ЗГ).([0-9])`)

func adaptBuilding(b string) string {
	mm := bRe.FindAllStringSubmatch(b, -1)
	for _, m := range mm {
		if len(m) != 4 {
			continue
		}

		if m[1] == "ЗГ" {
			return fmt.Sprintf("Зеленый городок, %s этаж", m[3])
		}

		n, _ := strconv.Atoi(m[2])

		return fmt.Sprintf("Корпус №%d, %s этаж", n, m[3])
	}

	return b
}
