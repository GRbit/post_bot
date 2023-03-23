package model

type State struct {
	Base

	ChatID               string
	Requested            bool
	CheckedIn            bool
	SearchCmdWasPrevious bool
	FileIDs              map[string]string
	Users                []*User
}
