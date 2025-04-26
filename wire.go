//go:build wireinject

package main

import (
	"ddd_demo/internal/repository"
	"ddd_demo/internal/repository/cache"
	"ddd_demo/internal/repository/dao"
	"ddd_demo/internal/service"
	"ddd_demo/internal/web"
	"ddd_demo/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		dao.NewUserDAO,

		cache.NewCodeCache, cache.NewUserCache,

		repository.NewCodeRepository,
		repository.NewCachedUserRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,

		// handler 部分
		web.NewUserHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}
