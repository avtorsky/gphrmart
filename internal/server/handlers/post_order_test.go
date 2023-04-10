package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/avtorsky/gphrmart/internal/accrual"
	"github.com/avtorsky/gphrmart/internal/storage"
	"github.com/avtorsky/gphrmart/internal/storage/queue"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type PostOrderTestSuite struct {
	suite.Suite
	AuthHandlerTestSuite
	accrualService *accrual.MockAccrualer
	queue          *queue.MockOrderStatusNotifier
}

func (suite *PostOrderTestSuite) SetupSuite() {
	suite.ctrl = gomock.NewController(suite.T())
	suite.db = storage.NewMockStorager(suite.ctrl)
	suite.queue = queue.NewMockOrderStatusNotifier(suite.ctrl)

	suite.accrualService = accrual.NewMockAccrualer(suite.ctrl)
	suite.db.EXPECT().
		Queue().
		Return(suite.queue)

	postOrder := NewPostOrder(
		context.Background(),
		suite.db,
		2,
		suite.accrualService,
	)
	handler := http.HandlerFunc(postOrder.Handler)
	suite.setupAuth(handler)
}

func (suite *PostOrderTestSuite) TearDownSuite() {
	suite.ctrl.Finish()
}

func (suite *PostOrderTestSuite) makeRequest(
	testName string,
	auth bool,
	body io.Reader,
) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/user/orders", body)
	req.Header.Set("Content-Type", "text/plain")
	if auth {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer: %v", suite.tokenSign))
	}
	suite.handler.ServeHTTP(rr, req)
	log.Printf("[%v]: %v", testName, rr.Body.String())
	return rr
}

func (suite *PostOrderTestSuite) TestUnauthorized() {
	rr := suite.makeRequest("TestUnauthorized", false, bytes.NewBuffer([]byte(`651725122560505`)))
	suite.Equal(http.StatusUnauthorized, rr.Code)
}

func (suite *PostOrderTestSuite) TestOrderAccepted() {
	suite.db.EXPECT().
		AddOrder(gomock.Any(), gomock.Eq("651725122560505"), gomock.Eq(1)).
		Times(1).
		Return(nil)
	suite.queue.EXPECT().
		Add(gomock.Any(), gomock.Eq("651725122560505")).
		Times(1).
		Return(nil)

	rr := suite.makeRequest("TestOrderAccepted", true, bytes.NewBuffer([]byte(`651725122560505`)))
	suite.Equal(http.StatusAccepted, rr.Code)
}

func (suite *PostOrderTestSuite) TestOrderAlreadyAddedByThisUser() {
	suite.db.EXPECT().
		AddOrder(gomock.Any(), gomock.Eq("651725122560505"), gomock.Eq(1)).
		Times(1).
		Return(fmt.Errorf(
			"%w: orderID=%v userID=%v",
			storage.ErrOrderAlreadyAddedByThisUser,
			"651725122560505",
			1,
		))
	rr := suite.makeRequest("TestOrderAlreadyAddedByThisUser", true, bytes.NewBuffer([]byte(`651725122560505`)))
	suite.Equal(http.StatusOK, rr.Code)
}

func (suite *PostOrderTestSuite) TestOrderAlreadyAddedByOtherUser() {
	suite.db.EXPECT().
		AddOrder(gomock.Any(), gomock.Eq("651725122560505"), gomock.Eq(1)).
		Times(1).
		Return(fmt.Errorf("%w: orderID=%v userID=%v", storage.ErrOrderAlreadyAddedByOtherUser, "651725122560505", 1))
	rr := suite.makeRequest("TestOrderAlreadyAddedByOtherUser", true, bytes.NewBuffer([]byte(`651725122560505`)))
	suite.Equal(http.StatusConflict, rr.Code)
}

func (suite *PostOrderTestSuite) TestInvalidPayload() {
	rr := suite.makeRequest("TestInvalidPayload", true, bytes.NewBuffer([]byte(``)))
	suite.Equal(http.StatusUnprocessableEntity, rr.Code)
}

func (suite *PostOrderTestSuite) TestLuhnValidation() {
	rr := suite.makeRequest("TestLuhnValidation", true, bytes.NewBuffer([]byte(`12345678904`)))
	suite.Equal(http.StatusUnprocessableEntity, rr.Code)
}

func TestPostOrderTestSuite(t *testing.T) {
	suite.Run(t, new(PostOrderTestSuite))
}
