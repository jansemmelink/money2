package db

import "github.com/go-msvc/errors"

type Claim struct {
	UserID *string `json:"user_id"`
}

func ValidateID(id *string) error {
	if id == nil {
		return errors.Errorf("id=nil")
	}
	if len(*id) != 36 {
		return errors.Errorf("len(id=%s)=%d != 36", *id, len(*id))
	}
	return nil
}

var (
	ErrNotAllowed error
)

func init() {
	ErrNotAllowed = errors.Errorf("Not allowed")
}
