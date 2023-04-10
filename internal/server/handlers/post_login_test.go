package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avtorsky/gphrmart/internal/auth"
	"github.com/avtorsky/gphrmart/internal/models"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type LoginTestSuite struct {
	suite.Suite
	db      *storage.MockStorager
	handler http.HandlerFunc
	ctrl    *gomock.Controller
}

func (suite *LoginTestSuite) makeRequest(
	testName string,
	testContentType bool,
	body io.Reader,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/user/login", body)
	if testContentType {
		req.Header.Set("Content-Type", "application/json")
	}
	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr

}

func (suite *LoginTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorager(suite.ctrl)
	login := Login{
		Session: Session{
			db:             suite.db,
			signingKey:     []byte("secret42"),
			expireDuration: 1 * time.Hour,
		},
	}
	suite.handler = http.HandlerFunc(login.Handler)
}

func (suite *LoginTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *LoginTestSuite) TestOk() {
	jsonStr := []byte(`{"login":"gopher", "password": "mart123"}`)
	suite.db.EXPECT().
		FindUser(gomock.Any(), gomock.Eq("gopher"), gomock.Eq("mart123")).
		Times(1).
		Return(&models.User{
			ID:             1,
			Username:       "gopher",
			HashedPassword: auth.GenerateHash("mart123", "456"),
			Salt:           "456",
		}, nil)

	rr := suite.makeRequest("TestOk", true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *LoginTestSuite) TestUnauthorized() {
	jsonStr := []byte(`{"login":"gopher", "password": "mart124"}`)
	suite.db.EXPECT().
		FindUser(gomock.Any(), gomock.Eq("gopher"), gomock.Eq("mart124")).
		Times(1).
		Return(&models.User{
			ID:             1,
			Username:       "gopher",
			HashedPassword: auth.GenerateHash("mart123", "456"),
			Salt:           "456",
		}, nil)

	resp := suite.makeRequest("TestUnauthorized", true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusUnauthorized, resp.Code)
}

func (suite *LoginTestSuite) TestInvalidPayload() {
	jsonStr := []byte(`{"login":"gopher", "pass`)
	resp := suite.makeRequest("TestInvalidPayload", true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusBadRequest, resp.Code)
}

func (suite *LoginTestSuite) TestInvalidContentType() {
	resp := suite.makeRequest("TestInvalidContentType", false, nil)
	suite.Equal(http.StatusBadRequest, resp.Code)
}

func TestLoginTestSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}
