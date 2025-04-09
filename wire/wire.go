//go:build wireinject

package wire

import (
	"ddd_demo/wire/repository"
	"ddd_demo/wire/repository/dao"
	"github.com/google/wire"
)

func InitUserRepository() *repository.UserRepository {
	wire.Build(repository.NewUserRepository, dao.NewUserDao, InitDB)
	return &repository.UserRepository{}
}
