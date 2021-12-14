package db

import "time"

// BaseDto
type BaseDto struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// StringToNull transforms empty string to nil string, so that gorm stores it as NULL
func StringToNull(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NullToString transforms NULL to empty string
func NullToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
