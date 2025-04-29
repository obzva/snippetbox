package model

import "errors"

var (
	ErrNoRecord           = errors.New("model: no matching record found")
	ErrInvalidCredentials = errors.New("model: invalid credentials")
	ErrDuplicateEmail     = errors.New("model: duplicate email")
)
