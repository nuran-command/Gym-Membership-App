package domain

import "time"

type User struct {
	ID        string
	Name      string
	Email     string
	Credits   int
	CreatedAt time.Time
}

type Membership struct {
	ID        string
	UserID    string
	Type      string
	Status    string
	ExpiresAt time.Time
}

type Booking struct {
	ID        string
	UserID    string
	AssetID   string
	StartTime time.Time
	EndTime   time.Time
	Status    string
}

type CreditTransaction struct {
	ID        string
	UserID    string
	Amount    int
	Type      string
	Reason    string
	CreatedAt time.Time
}
