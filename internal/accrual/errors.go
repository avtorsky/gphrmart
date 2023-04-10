package accrual

import "errors"

var ErrTooManyRequests = errors.New("accrual endpoint rate-limit reached")