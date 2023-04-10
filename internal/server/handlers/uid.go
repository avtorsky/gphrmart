package handlers

import (
	"context"
	"fmt"

	"github.com/go-chi/jwtauth/v5"
)

func GetUID(ctx context.Context) (int, error) {
	_, claims, _ := jwtauth.FromContext(ctx)
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("%w: %+v", ErrInvalidClaims, claims)
	}
	return int(userID), nil
}
