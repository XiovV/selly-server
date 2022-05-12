package jwt

import "errors"

var (
	ErrInvalidSignature = errors.New("signature is invalid")
	ErrTokenExpired     = errors.New("token has expired")
)
