package storage

import "errors"

var ErrUserAlreadyExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrOrderAlreadyAddedByThisUser = errors.New("order already added by this user")
var ErrOrderAlreadyAddedByOtherUser = errors.New("order already added by other user")
var ErrEmptyResult = errors.New("query returned empty result")
var ErrOrderNotFound = errors.New("order with specified orderID not found")
