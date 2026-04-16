package notifications_test

import (
	"context"
	"delay/internal/config"
	"delay/internal/domain"
	"delay/internal/service/notifications"
	mockservice "delay/internal/service/notifications/mocks"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wb-go/wbf/logger"
)

func newTestLogger() *logger.ZerologAdapter {
	return logger.NewZerologAdapter("test", "test")
}

func newTestService(
	t *testing.T,
) (*notifications.Service, *mockservice.MockRepository, *mockservice.MockCache, *mockservice.MockPublisher) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockRepo := mockservice.NewMockRepository(ctrl)
	mockCache := mockservice.NewMockCache(ctrl)
	mockPub := mockservice.NewMockPublisher(ctrl)

	cfg := &config.Config{}
	log := newTestLogger()

	s := notifications.New(mockRepo, mockPub, mockCache, cfg, log)
	return s, mockRepo, mockCache, mockPub
}

func TestService_Create_Success(t *testing.T) {
	s, mockRepo, mockCache, mockPub := newTestService(t)

	notification := &domain.Notification{
		Message:     "Test create",
		DataToSent:  time.Now().Add(2 * time.Hour),
		Channel:     domain.ChannelEmail,
		Destination: "test@example.com",
	}
	id := uuid.New()

	mockRepo.EXPECT().Create(gomock.Any(), notification).Return(id, nil)
	mockRepo.EXPECT().UpdateStatus(gomock.Any(), id, "created").Return(nil)
	mockRepo.EXPECT().Get(gomock.Any(), id).Return(notification, nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:status:"+id.String(), "created", 5*time.Minute).Return(nil)
	mockCache.EXPECT().Delete(gomock.Any(), "notification:getall").Return(nil)
	mockPub.EXPECT().Publish(gomock.Any(), notification).Return(nil)

	gotID, err := s.Create(context.Background(), notification)
	assert.NoError(t, err)
	assert.Equal(t, id, gotID)
}

func TestService_Create_FailCreate(t *testing.T) {
	s, mockRepo, _, _ := newTestService(t)

	notification := &domain.Notification{
		Message: "Fail create",
	}

	mockRepo.EXPECT().Create(gomock.Any(), notification).Return(uuid.Nil, errors.New("create fail"))

	gotID, err := s.Create(context.Background(), notification)
	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, gotID)
}

func TestService_Status_CacheHit(t *testing.T) {
	s, mockRepo, mockCache, _ := newTestService(t)

	id := uuid.New()
	status := "sent"

	mockCache.EXPECT().
		GetString(gomock.Any(), "notification:status:"+id.String()).
		Return("sent", nil)

	gotStatus, err := s.Status(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, status, gotStatus)

	_ = mockRepo
}

func TestService_Status_CacheMiss(t *testing.T) {
	s, mockRepo, mockCache, _ := newTestService(t)

	id := uuid.New()
	status := "created"

	mockCache.EXPECT().
		GetString(gomock.Any(), "notification:status:"+id.String()).
		Return("", errors.New("cache miss"))
	mockRepo.EXPECT().Status(gomock.Any(), id).Return(status, nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:status:"+id.String(), status, 5*time.Minute).Return(nil)

	gotStatus, err := s.Status(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, status, gotStatus)
}

func TestService_Cancel_Success(t *testing.T) {
	s, mockRepo, mockCache, _ := newTestService(t)

	id := uuid.New()

	mockRepo.EXPECT().Cancel(gomock.Any(), id).Return(nil)
	mockCache.EXPECT().Delete(gomock.Any(), "notification:getall").Return(nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:status:"+id.String(), "", time.Duration(0)).Return(nil)

	err := s.Cancel(context.Background(), id)
	assert.NoError(t, err)
}

func TestService_Cancel_Fail(t *testing.T) {
	s, mockRepo, _, _ := newTestService(t)

	id := uuid.New()

	mockRepo.EXPECT().Cancel(gomock.Any(), id).Return(errors.New("cancel error"))

	err := s.Cancel(context.Background(), id)
	assert.Error(t, err)
}

func TestService_GetAll_Success(t *testing.T) {
	s, mockRepo, mockCache, _ := newTestService(t)

	notes := []domain.Notification{
		{ID: uuid.New(), Message: "msg1"},
		{ID: uuid.New(), Message: "msg2"},
	}

	mockCache.EXPECT().GetJSON(gomock.Any(), "notification:getall", gomock.Any()).DoAndReturn(
		func(ctx context.Context, key string, dest any) error {
			ptr := dest.(*[]domain.Notification)
			*ptr = notes
			return nil
		},
	)

	gotNotifications, err := s.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, *gotNotifications, 2)

	_ = mockRepo
}

func TestService_GetAll_CacheMiss(t *testing.T) {
	s, mockRepo, mockCache, _ := newTestService(t)

	notes := []domain.Notification{
		{ID: uuid.New(), Message: "msg1"},
		{ID: uuid.New(), Message: "msg2"},
	}

	mockCache.EXPECT().GetJSON(gomock.Any(), "notification:getall", gomock.Any()).Return(errors.New("cache miss"))
	mockRepo.EXPECT().GetAll(gomock.Any()).Return(&notes, nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:getall", gomock.Any(), 5*time.Minute).Return(nil)

	gotNotifications, err := s.GetAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, *gotNotifications, 2)
}

func TestService_UpdateStatus(t *testing.T) {
	s, mockRepo, mockCache, _ := newTestService(t)

	id := uuid.New()
	status := "done"

	mockRepo.EXPECT().UpdateStatus(gomock.Any(), id, status).Return(nil)
	mockCache.EXPECT().Delete(gomock.Any(), "notification:getall").Return(nil)
	mockCache.EXPECT().Set(gomock.Any(), "notification:status:"+id.String(), status, 5*time.Minute).Return(nil)

	err := s.UpdateStatus(context.Background(), id, status)
	assert.NoError(t, err)
}
