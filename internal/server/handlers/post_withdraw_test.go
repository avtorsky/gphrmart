package handlers

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/avtorsky/gphrmart/internal/models"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type WithdrawTestSuite struct {
	suite.Suite
	AuthHandlerTestSuite
}

func (suite *WithdrawTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorager(suite.ctrl)
	withdraw := Withdraw{db: suite.db}
	handler := http.HandlerFunc(withdraw.Handler)
	suite.setupAuth(handler)
}

func (suite *WithdrawTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *WithdrawTestSuite) makeRequest(
	testName string,
	auth,
	testContentType bool,
	body io.Reader,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/user/balance/withdraw", body)
	if testContentType {
		req.Header.Set("Content-Type", "application/json")
	}
	if auth {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer: %v", suite.tokenSign))
	}
	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *WithdrawTestSuite) TestUnauthorized() {
	rr := suite.makeRequest("TestUnauthorized", false, false, nil)
	suite.Equal(http.StatusUnauthorized, rr.Code)
}

func (suite *WithdrawTestSuite) TestInvalidContentType() {
	rr := suite.makeRequest("TestInvalidContentType", true, false, nil)
	suite.Equal(http.StatusBadRequest, rr.Code)
}

func (suite *WithdrawTestSuite) TestNotEnoughBalance() {
	jsonStr := []byte(`{"order":"651725122560505", "sum": 442}`)

	suite.db.EXPECT().
		GetBalance(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(&models.Balance{
			Current:   sql.NullFloat64{Float64: 100, Valid: true},
			Withdrawn: sql.NullFloat64{Float64: 100, Valid: true},
		}, nil)

	rr := suite.makeRequest("TestNotEnoughBalance", true, true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusPaymentRequired, rr.Code)
}

func (suite *WithdrawTestSuite) TestOK() {
	jsonStr := []byte(`{"order":"651725122560505", "sum": 42}`)

	suite.db.EXPECT().
		GetBalance(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(&models.Balance{
			Current:   sql.NullFloat64{Float64: 100, Valid: true},
			Withdrawn: sql.NullFloat64{Float64: 100, Valid: true},
		}, nil)

	suite.db.EXPECT().
		AddWithdrawalRecord(gomock.Any(), gomock.Eq("651725122560505"), gomock.Eq(42.0), gomock.Eq(1)).
		Times(1).
		Return(nil)

	rr := suite.makeRequest("TestOK", true, true, bytes.NewBuffer(jsonStr))
	suite.Equal(http.StatusOK, rr.Code)
}

func TestWithdrawTestSuite(t *testing.T) {
	suite.Run(t, new(WithdrawTestSuite))
}
