package otpbot

import "errors"

var (
	ErrNotLinked      = errors.New("telegram not linked")
	ErrBadRequest     = errors.New("otp bot bad request")
	ErrUnauthorized   = errors.New("otp bot unauthorized")
	ErrRateLimited    = errors.New("otp bot rate limited")
	ErrDeliveryFailed = errors.New("otp delivery failed")
)
