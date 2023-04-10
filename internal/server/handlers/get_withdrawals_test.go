package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/avtorsky/gphrmart/internal/models"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WithdrawalsTestSuite struct {
	suite.Suite
	AuthHandlerTestSuite
}

func (suite *WithdrawalsTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorager(suite.ctrl)
	withdrawals := Withdrawals{db: suite.db}
	handler := http.HandlerFunc(withdrawals.Handler)
	suite.setupAuth(handler)
}

func (suite *WithdrawalsTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *WithdrawalsTestSuite) makeRequest(
	testName string,
	auth bool,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/user/widtdrawals", nil)
	if auth {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer: %v", suite.tokenSign))
	}
	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *WithdrawalsTestSuite) TestUnauthorized() {
	rr := suite.makeRequest("TestUnauthorized", false)
	suite.Equal(http.StatusUnauthorized, rr.Code)
}

func (suite *WithdrawalsTestSuite) TestNoContent() {
	suite.db.EXPECT().
		GetWithdrawals(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(nil, storage.ErrEmptyResult)

	rr := suite.makeRequest("TestNoContent", true)
	suite.Equal(http.StatusNoContent, rr.Code)
}

func (suite *WithdrawalsTestSuite) TestQueue() {
	start := time.Now()
	withdrawals := []models.Withdrawal{
		{
			Order:       "651725122560505",
			Sum:         442,
			ProcessedAt: start,
		},
		{
			Order:       "12345678903",
			Sum:         42,
			ProcessedAt: start.Add(5 * time.Second),
		},
	}

	suite.db.EXPECT().
		GetWithdrawals(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(withdrawals, nil)

	rr := suite.makeRequest("TestQueue", true)
	suite.Equal(http.StatusOK, rr.Code)

	expected := fmt.Sprintf(
		`[{"order":"651725122560505","sum":442,"processed_at":"%v"},
		{"order":"12345678903","sum":42,"processed_at":"%v"}]`,
		start.Format(time.RFC3339),
		start.Add(5*time.Second).Format(time.RFC3339),
	)
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func TestWithdrawalsTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawalsTestSuite))
}
