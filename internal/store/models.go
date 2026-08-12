package store

import "strings"

type Account struct {
	ID                int64
	Name              string
	Email             string
	Password          string
	ProfileID         string
	AutomationFlow    string
	AutomationParams  map[string]interface{}
	AutomationEnabled bool
	Fanpages          []Fanpage
}

type Fanpage struct {
	ID       int64
	FBPageID string
	Name     string
}

// LoginID returns the identifier for the Facebook login screen.
func (a Account) LoginID() string {
	email := strings.TrimSpace(a.Email)
	profile := strings.TrimSpace(a.ProfileID)
	if email != "" && strings.Contains(email, "@") {
		return email
	}
	if profile != "" {
		return profile
	}
	return email
}

type Device struct {
	ID           int64
	Serial       string
	Label        string
	AdbIndex     int
	MaxAccounts  int
	Status       string
}

type Assignment struct {
	ID         int64
	DeviceID   int64
	AccountID  int64
	SlotNo     int
	Active     bool
	LastUsedAt *string
}
