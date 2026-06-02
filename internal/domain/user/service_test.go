package user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/hafis915/fintrack/internal/domain/user"
)

type repoMock struct{ mock.Mock }

func (m *repoMock) GetByUserID(ctx context.Context, id uuid.UUID) (*user.Profile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *repoMock) Create(ctx context.Context, in user.CreateProfileInput) (*user.Profile, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *repoMock) UpdateLifestyle(ctx context.Context, id uuid.UUID, ls *string, em *int) (*user.Profile, error) {
	args := m.Called(ctx, id, ls, em)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.Profile), args.Error(1)
}
func (m *repoMock) UpdateIncome(ctx context.Context, id uuid.UUID, enc, hint string) (string, error) {
	args := m.Called(ctx, id, enc, hint)
	return args.String(0), args.Error(1)
}

type encMock struct{ mock.Mock }

func (m *encMock) EncryptIncome(amount int64) (string, error) {
	args := m.Called(amount)
	return args.String(0), args.Error(1)
}

func TestUpdateIncome_EncryptsAndReturnsHint(t *testing.T) {
	repo := &repoMock{}
	enc := &encMock{}
	uid := uuid.New()

	enc.On("EncryptIncome", int64(8_000_000)).Return("CIPHER", nil)
	repo.On("UpdateIncome", mock.Anything, uid, "CIPHER", "Rp 8jt").Return("Rp 8jt", nil)

	svc := user.NewService(repo, enc)
	hint, err := svc.UpdateIncome(context.Background(), uid, 8_000_000)
	require.NoError(t, err)
	require.Equal(t, "Rp 8jt", hint)
}
