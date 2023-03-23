package model

type User struct {
	Base
	BookingID *int64
	Requested *bool
	CheckedIn *bool
}
