package domain

import "time"

type User struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Credits   int       `db:"credits" json:"credits"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Membership struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Type      string    `db:"type" json:"type"`
	Status    string    `db:"status" json:"status"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
}

type Booking struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	AssetID   string    `db:"asset_id" json:"asset_id"`
	StartTime time.Time `db:"start_time" json:"start_time"`
	EndTime   time.Time `db:"end_time" json:"end_time"`
	Status    string    `db:"status" json:"status"`
}

type CreditTransaction struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Amount    int       `db:"amount" json:"amount"`
	Type      string    `db:"type" json:"type"`
	Reason    string    `db:"reason" json:"reason"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
