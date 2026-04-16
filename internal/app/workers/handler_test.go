package workers_test

import (
	"context"
	"delay/internal/app/workers"
	"delay/internal/domain"
	"delay/internal/service/notifications"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wb-go/wbf/logger"
)

// MockStatusUpdater is a mock of the statusUpdater interface.
type MockStatusUpdater struct {
	mock.Mock
}

func (m *MockStatusUpdater) Status(ctx context.Context, id uuid.UUID) (string, error) {
	args := m.Called(ctx, id)
	return args.String(0), args.Error(1)
}

func (m *MockStatusUpdater) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

// MockMailer is a mock of the mailer interface.
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) Send(ctx context.Context, message, destination string) error {
	args := m.Called(ctx, message, destination)
	return args.Error(0)
}

// MockTelegram is a mock of the telegram interface.
type MockTelegram struct {
	mock.Mock
}

func (m *MockTelegram) Send(ctx context.Context, message, destination string) error {
	args := m.Called(ctx, message, destination)
	return args.Error(0)
}

func TestNotificationHandler_Handle(t *testing.T) {
	log := logger.NewZerologAdapter("test", "test")
	ctx := context.Background()
	id := uuid.New()

	n := domain.Notification{
		ID:          id,
		Message:     "Hello",
		Destination: "test@example.com",
		Channel:     domain.ChannelEmail,
	}

	t.Run("Success Delivery", func(t *testing.T) {
		mockStatus := new(MockStatusUpdater)
		mockMail := new(MockMailer)
		
		handler := workers.NewNotificationHandler(mockStatus, mockMail, nil, log)

		mockStatus.On("Status", mock.Anything, id).Return(string(domain.StatusCreated), nil)
		mockMail.On("Send", mock.Anything, "Hello", "test@example.com").Return(nil)
		mockStatus.On("UpdateStatus", mock.Anything, id, string(domain.StatusSent)).Return(nil)

		err := handler.Handle(ctx, n)
		assert.NoError(t, err)
		mockStatus.AssertExpectations(t)
		mockMail.AssertExpectations(t)
	})

	t.Run("Skip if record not found (Cancelled)", func(t *testing.T) {
		mockStatus := new(MockStatusUpdater)
		handler := workers.NewNotificationHandler(mockStatus, nil, nil, log)

		mockStatus.On("Status", mock.Anything, id).Return("", notifications.ErrNotFound)

		err := handler.Handle(ctx, n)
		assert.NoError(t, err) // Should skip without error
		mockStatus.AssertExpectations(t)
	})

	t.Run("Skip if status is canceled", func(t *testing.T) {
		mockStatus := new(MockStatusUpdater)
		handler := workers.NewNotificationHandler(mockStatus, nil, nil, log)

		mockStatus.On("Status", mock.Anything, id).Return(string(domain.StatusCancelled), nil)

		err := handler.Handle(ctx, n)
		assert.NoError(t, err) // Should skip without error
		mockStatus.AssertExpectations(t)
	})
}
