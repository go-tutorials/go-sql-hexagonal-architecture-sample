package port

import (
	"context"

	"go-service/internal/user/domain"
)

type UserRepository interface {
	Load(ctx context.Context, id string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) (int64, error)
	Update(ctx context.Context, user *domain.User) (int64, error)
	Patch(ctx context.Context, user map[string]interface{}) (int64, error)
	Delete(ctx context.Context, id string) (int64, error)
	Search(ctx context.Context, filter *domain.UserFilter) ([]domain.User, int64, error)
}
