package types

// UserEconomy represents user's economic data
type UserEconomy struct {
	UserID           int64 `json:"userId"`
	Balance          int64 `json:"balance"`
	VerificationLevel int   `json:"verificationLevel"`
	Energy            int32 `json:"energy"`
	MaxEnergy         int32 `json:"maxEnergy"`
	TapValue          int   `json:"tapValue"`
	RegenSpeed        int   `json:"regenSpeed"`
	LastEnergyTime    int64 `json:"lastEnergyTime"`
	PendingBalance    int64 `json:"pendingBalance"`
	ReferralID        *int64 `json:"referralId"`
	ReferralCount     int   `json:"referralCount"`
	LastTapTime       int64 `json:"lastTapTime"`
	IsBanned          bool  `json:"isBanned"`
}
