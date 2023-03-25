package db

import "strings"

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
