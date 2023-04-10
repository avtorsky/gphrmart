package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/avtorsky/gphrmart/internal/models"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BalanceTestSuite struct {
	AuthHandlerTestSuite
}

func (suite *BalanceTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorager(suite.ctrl)
	balance := Balance{db: suite.db}
	handler := http.HandlerFunc(balance.Handler)
	suite.setupAuth(handler)
}

func (suite *BalanceTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *BalanceTestSuite) makeRequest(testName string, auth bool) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/user/balance", nil)
	req.Header.Set("Content-Type", "text/plain")
	if auth {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer: %v", suite.tokenSign))
	}
	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *BalanceTestSuite) TestUnauthorized() {
	rr := suite.makeRequest("TestUnauthorized", false)
	suite.Equal(http.StatusUnauthorized, rr.Code)
}

func (suite *BalanceTestSuite) TestZeroTransactions() {
	suite.db.EXPECT().
		GetBalance(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(&models.Balance{
			Current:   sql.NullFloat64{Float64: 0, Valid: false},
			Withdrawn: sql.NullFloat64{Float64: 0, Valid: false},
		}, nil)

	rr := suite.makeRequest("TestZeroTransactions", true)
	suite.Equal(http.StatusOK, rr.Code)
	expected := `{
		"current": 0,
		"withdrawn": 0
	}`
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func (suite *BalanceTestSuite) TestNonZeroTransactions() {
	suite.db.EXPECT().
		GetBalance(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(&models.Balance{
			Current:   sql.NullFloat64{Float64: 420.42, Valid: true},
			Withdrawn: sql.NullFloat64{Float64: 42.42, Valid: true},
		}, nil)

	rr := suite.makeRequest("TestNonZeroTransactions", true)
	suite.Equal(http.StatusOK, rr.Code)
	expected := `{
		"current": 420.42,
		"withdrawn": 42.42
	}`
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func TestBalanceTestSuite(t *testing.T) {
	suite.Run(t, new(BalanceTestSuite))
}
