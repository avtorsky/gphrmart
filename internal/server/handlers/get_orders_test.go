package handlers

import (
	"database/sql"
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

type GetOrdersTestSuite struct {
	AuthHandlerTestSuite
}

func (suite *GetOrdersTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorager(suite.ctrl)
	getOrder := GetOrder{db: suite.db}
	handler := http.HandlerFunc(getOrder.Handler)
	suite.setupAuth(handler)
}

func (suite *GetOrdersTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *GetOrdersTestSuite) makeRequest(
	testName string,
	auth bool,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/user/orders", nil)
	req.Header.Set("Content-Type", "text/plain")
	if auth {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer: %v", suite.tokenSign))
	}
	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *GetOrdersTestSuite) TestUnauthorized() {
	rr := suite.makeRequest("TestUnauthorized", false)
	suite.Equal(http.StatusUnauthorized, rr.Code)
}

func (suite *GetOrdersTestSuite) TestNoContent() {
	suite.db.EXPECT().
		FindOrdersByUserID(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(nil, storage.ErrEmptyResult)

	rr := suite.makeRequest("TestNoContent", true)
	suite.Equal(http.StatusNoContent, rr.Code)
}

func (suite *GetOrdersTestSuite) TestOrderNew() {
	start := time.Now()
	orders := []models.Order{
		{
			ID:         "651725122560505",
			Status:     "NEW",
			UploadedAt: start,
			Accrual:    sql.NullFloat64{Float64: 0, Valid: false},
		},
	}

	suite.db.EXPECT().
		FindOrdersByUserID(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(orders, nil)

	rr := suite.makeRequest("TestOrderNew", true)

	suite.Equal(http.StatusOK, rr.Code)
	expected := fmt.Sprintf(
		`[{"number":"651725122560505","status":"NEW","uploaded_at":"%v"}]`,
		start.Format(time.RFC3339),
	)
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func (suite *GetOrdersTestSuite) TestOrderProcessing() {
	start := time.Now()
	orders := []models.Order{
		{
			ID:         "651725122560505",
			Status:     "PROCESSING",
			UploadedAt: start,
			Accrual:    sql.NullFloat64{Float64: 0, Valid: false},
		},
	}

	suite.db.EXPECT().
		FindOrdersByUserID(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(orders, nil)

	rr := suite.makeRequest("TestOrderProcessing", true)

	suite.Equal(http.StatusOK, rr.Code)
	expected := fmt.Sprintf(
		`[{"number":"651725122560505","status":"PROCESSING","uploaded_at":"%v"}]`,
		start.Format(time.RFC3339),
	)
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func (suite *GetOrdersTestSuite) TestOrderInvalid() {
	start := time.Now()
	orders := []models.Order{
		{
			ID:         "651725122560505",
			Status:     "INVALID",
			UploadedAt: start,
			Accrual:    sql.NullFloat64{Float64: 0, Valid: false},
		},
	}

	suite.db.EXPECT().
		FindOrdersByUserID(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(orders, nil)

	rr := suite.makeRequest("TestOrderInvalid", true)

	suite.Equal(http.StatusOK, rr.Code)
	expected := fmt.Sprintf(
		`[{"number":"651725122560505","status":"INVALID","uploaded_at":"%v"}]`,
		start.Format(time.RFC3339),
	)
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func (suite *GetOrdersTestSuite) TestOrderProcessed() {
	start := time.Now()
	orders := []models.Order{
		{
			ID:         "651725122560505",
			Status:     "PROCESSED",
			UploadedAt: start,
			Accrual:    sql.NullFloat64{Float64: 420.42, Valid: true},
		},
	}

	suite.db.EXPECT().
		FindOrdersByUserID(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(orders, nil)

	rr := suite.makeRequest("TestOrderProcessed", true)

	suite.Equal(http.StatusOK, rr.Code)
	expected := fmt.Sprintf(
		`[{"number":"651725122560505","status":"PROCESSED","uploaded_at":"%v", "accrual": 420.42}]`,
		start.Format(time.RFC3339),
	)
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func (suite *GetOrdersTestSuite) TestQueue() {
	start := time.Now()
	orders := []models.Order{
		{
			ID:         "651725122560505",
			Status:     "PROCESSED",
			UploadedAt: start,
			Accrual:    sql.NullFloat64{Float64: 42.42, Valid: true},
		},
		{
			ID:         "12345678903",
			Status:     "NEW",
			UploadedAt: start.Add(5 * time.Second),
			Accrual:    sql.NullFloat64{Float64: 0, Valid: false},
		},
	}

	suite.db.EXPECT().
		FindOrdersByUserID(gomock.Any(), gomock.Eq(1)).
		Times(1).
		Return(orders, nil)

	rr := suite.makeRequest("TestQueue", true)

	suite.Equal(http.StatusOK, rr.Code)
	expected := fmt.Sprintf(
		`[{"number":"651725122560505","status":"PROCESSED","uploaded_at":"%v", "accrual": 42.42},
		{"number":"12345678903","status":"NEW","uploaded_at":"%v"}]`,
		start.Format(time.RFC3339),
		start.Add(5*time.Second).Format(time.RFC3339),
	)
	assert.JSONEq(suite.T(), expected, rr.Body.String())
}

func TestGetOrdersTestSuite(t *testing.T) {
	suite.Run(t, new(GetOrdersTestSuite))
}
