package handlers

import "errors"

var ErrInvalidRequest = errors.New("invalid request")
var ErrInvalidContentType = errors.New("invalid request content-type header")
var ErrInvalidPayload = errors.New("invalid request payload")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidBalance = errors.New("insufficient balance sum")
var ErrInvalidClaims = errors.New("invalid claims")
var ErrServerShutdown = errors.New("server is shutting down")
