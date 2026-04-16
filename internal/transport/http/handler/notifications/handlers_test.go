package notifications_test

import (
	"bytes"
	"delay/internal/domain"
	"delay/internal/transport/http/handler/notifications"
	"delay/internal/transport/http/handler/notifications/mocks"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wb-go/wbf/logger"
)

func newTestLogger() *logger.ZerologAdapter {
	return logger.NewZerologAdapter("test", "test")
}

func TestCreateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	notification := domain.Notification{
		Message:     "Test create",
		DataToSent:  time.Now().Add(time.Hour),
		Channel:     domain.ChannelEmail,
		Destination: "test@example.com",
	}
	id := uuid.New()

	mockHandler.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(id, nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	reqBody, _ := json.Marshal(map[string]any{
		"message":      notification.Message,
		"data_sent_at": notification.DataToSent.Format(time.RFC3339),
		"channel":      string(notification.Channel),
		"destination":  notification.Destination,
	})
	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp uuid.UUID
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, id, resp)
}

func TestCreateHandler_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader([]byte("{invalid-json")))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateHandler_MissingMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	mockHandler.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(uuid.Nil, errors.New("message is required")).
		Times(1)

	postData := map[string]any{
		"message":      "",
		"data_sent_at": time.Now().Add(time.Hour).Format(time.RFC3339),
		"channel":      string(domain.ChannelEmail),
		"destination":  "test@example.com",
	}
	reqBody, _ := json.Marshal(postData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateHandler_PastDataToSent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	mockHandler.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(uuid.Nil, errors.New("data sent must be in the future")).
		Times(1)

	postData := map[string]any{
		"message":      "test",
		"data_sent_at": time.Now().Add(-time.Hour).Format(time.RFC3339),
		"channel":      string(domain.ChannelEmail),
		"destination":  "test@example.com",
	}
	reqBody, _ := json.Marshal(postData)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/notify/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Create(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetAllHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	notes := []domain.Notification{
		{ID: uuid.New(), Message: "Test1", Status: "created", DataToSent: time.Now(), Channel: domain.ChannelEmail, Destination: "test1@example.com"},
	}

	mockHandler.EXPECT().
		GetAll(gomock.Any()).
		Return(&notes, nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/all", nil)

	h.GetAll(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []domain.Notification
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
}

func TestGetAllHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	mockHandler.EXPECT().
		GetAll(gomock.Any()).
		Return(nil, errors.New("db error")).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/all", nil)

	h.GetAll(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetStatusHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	id := uuid.New()
	mockHandler.EXPECT().
		Status(gomock.Any(), id).
		Return("created", nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/status/"+id.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: id.String()}}

	h.GetStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "created", resp)
}

func TestGetStatusHandler_BadUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	badID := "bad-uuid"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notify/status/"+badID, nil)
	c.Params = []gin.Param{{Key: "id", Value: badID}}

	h.GetStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	id := uuid.New()
	mockHandler.EXPECT().
		Cancel(gomock.Any(), id).
		Return(nil).
		Times(1)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/notify/cancel/"+id.String(), nil)
	c.Params = []gin.Param{{Key: "id", Value: id.String()}}

	h.Cancel(c)
	c.Writer.WriteHeaderNow()

	res := w.Result()
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.Empty(t, w.Body.String())
}

func TestCancelHandler_BadUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := newTestLogger()
	mockHandler := mocks.NewMockNotificationHandler(ctrl)
	h := notifications.New(mockHandler, log)

	badID := "bad-uuid"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/notify/cancel/"+badID, nil)
	c.Params = []gin.Param{{Key: "id", Value: badID}}

	h.Cancel(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"invalid request body"}`, w.Body.String())
}
