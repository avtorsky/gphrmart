package handlers

import (
	"net/http"
	"time"

	"github.com/avtorsky/gphrmart/internal/auth"
	"github.com/avtorsky/gphrmart/internal/models"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/go-chi/jwtauth/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	db        *storage.MockStorager
	handler   http.Handler
	tokenSign string
	ctrl      *gomock.Controller
}

func (suite *AuthHandlerTestSuite) setupAuth(
	handler http.HandlerFunc,
) {
	signingKey := []byte("secret42")
	token := auth.GenerateJWTToken(
		&models.User{ID: 1, Username: "gopher"},
		signingKey,
		time.Hour,
	)
	tokenSign, _ := token.SignedString(signingKey)
	suite.tokenSign = tokenSign
	tokenAuth := jwtauth.New("HS256", signingKey, nil)
	verifier := jwtauth.Verifier(tokenAuth)
	authentifier := jwtauth.Authenticator
	suite.handler = verifier(authentifier(handler))
}
