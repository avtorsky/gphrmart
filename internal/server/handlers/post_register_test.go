package handlers

import (
	"bytes"
	"fmt"
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

type RegisterTestSuite struct {
	suite.Suite
	db      *storage.MockStorager
	handler http.HandlerFunc
	ctrl    *gomock.Controller
}

func (suite *RegisterTestSuite) makeRequest(
	testName string,
	testContentType bool,
	body io.Reader,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/user/register", body)
	if testContentType {
		req.Header.Set("Content-Type", "application/json")
	}
	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *RegisterTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorager(suite.ctrl)

	register := Register{
		Session: Session{
			db:             suite.db,
			signingKey:     []byte("secret42"),
			expireDuration: 1 * time.Hour,
		},
	}
	suite.handler = http.HandlerFunc(register.Handler)
}

func (suite *RegisterTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *RegisterTestSuite) TestOK() {
	jsonStr := []byte(`{"login":"gopher", "password": "mart123"}`)
	suite.db.EXPECT().
		AddUser(gomock.Any(), gomock.Eq("gopher"), gomock.Eq("mart123")).
		Times(1).
		Return(&models.User{
			ID:             1,
			Username:       "gopher",
			HashedPassword: auth.GenerateHash("mart123", "456"),
			Salt:           "456",
		}, nil)

	rr := suite.makeRequest("TestOK", true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *RegisterTestSuite) TestUserAlreadyExists() {
	jsonStr := []byte(`{"login":"gopher", "password": "mart123"}`)
	suite.db.EXPECT().
		AddUser(gomock.Any(), gomock.Eq("gopher"), gomock.Eq("mart123")).
		Times(1).
		Return(&models.User{
			ID:             1,
			Username:       "gopher",
			HashedPassword: auth.GenerateHash("mart123", "456"),
			Salt:           "456",
		}, nil)
	resp1 := suite.makeRequest("TestUserAlreadyExists", true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusOK, resp1.Code)
	suite.db.EXPECT().
		AddUser(gomock.Any(), gomock.Eq("gopher"), gomock.Eq("mart123")).
		Times(1).
		Return(nil, fmt.Errorf("%w: %v", storage.ErrUserAlreadyExists, "gopher"))
	resp2 := suite.makeRequest("TestUserAlreadyExists", true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusConflict, resp2.Code)
}

func (suite *RegisterTestSuite) TestInvalidPayload() {
	jsonStr := []byte(`{"login":"gopher", "pass`)
	resp1 := suite.makeRequest("TestInvalidPayload", true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusBadRequest, resp1.Code)
}

func (suite *RegisterTestSuite) TestInvalidContentType() {
	jsonStr := []byte(`{"login":"gopher", "pass`)
	resp1 := suite.makeRequest("TestInvalidContentType", false, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusBadRequest, resp1.Code)

}

func TestRegisterTestSuite(t *testing.T) {
	suite.Run(t, new(RegisterTestSuite))
}
