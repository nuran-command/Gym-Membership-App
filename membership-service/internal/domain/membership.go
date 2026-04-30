package domain

type Membership struct {
	ID        string
	UserID    string
	AssetID   string
	Status    string
	StartDate string
	EndDate   string
}

type MembershipRepository interface {
	Create(m *Membership) error
	GetByID(id string) (*Membership, error)
	Cancel(id string) (*Membership, error)
	ValidateAccess(userID, assetID string) (bool, error)
}

type MembershipUsecase interface {
	CreateMembership(userID, assetID, start, end string) (*Membership, error)
	GetMembership(id string) (*Membership, error)
	CancelMembership(id string) (*Membership, error)
	ValidateAccess(userID, assetID string) (bool, error)
}
