package app

import (
	"context"

	v "github.com/core-go/core/v10"
	"github.com/core-go/health"
	"github.com/core-go/log/zap"
	q "github.com/core-go/sql"

	"go-service/internal/user/adapter/handler"
	"go-service/internal/user/adapter/repository"
	"go-service/internal/user/port"
	"go-service/internal/user/service"
)

type ApplicationContext struct {
	Health *health.Handler
	User   port.UserPort
}

func NewApp(ctx context.Context, cfg Config) (*ApplicationContext, error) {
	db, err := q.OpenByConfig(cfg.Sql)
	if err != nil {
		return nil, err
	}
	logError := log.LogError
	validator, err := v.NewValidator()
	if err != nil {
		return nil, err
	}

	userRepository, err := repository.NewUserAdapter(db, repository.BuildQuery)
	userService := service.NewUserService(db, userRepository)
	userHandler := handler.NewUserHandler(userService, validator.Validate, logError)

	sqlChecker := q.NewHealthChecker(db)
	healthHandler := health.NewHandler(sqlChecker)

	return &ApplicationContext{
		Health: healthHandler,
		User:   userHandler,
	}, nil
}
