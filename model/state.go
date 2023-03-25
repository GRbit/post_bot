package model

const (
	CmdGiveMeSome   = "give_me_some"
	CmdAddAddress   = "add_address"
	CmdAddInstagram = "add_instagram"
	CmdAddWishes    = "add_wishes"
	CmdAddName      = "add_name"
	CmdMyData       = "my_data"
)

type State struct {
	Base

	ChatID            int64
	GivenAddressesCtr int
	PreviousCmd       string
	FileIDs           map[string]string
	Telegram          string
	Users             []*User
}
