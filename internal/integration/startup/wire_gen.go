// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"ddd_demo/internal/repository"
	"ddd_demo/internal/repository/cache"
	"ddd_demo/internal/repository/dao"
	"ddd_demo/internal/service"
	"ddd_demo/internal/web"
	"ddd_demo/internal/web/jwt"
	"ddd_demo/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	db := InitDB()
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, handler, codeService)
	articleDAO := dao.NewArticleGORMDAO(db)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, userRepository, articleCache)
	articleService := service.NewArticleService(articleRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService)
	wechatService := InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, handler, userService)
	engine := ioc.InitWebServer(v, userHandler, articleHandler, oAuth2WechatHandler)
	return engine
}

func InitArticleHandler(dao2 dao.ArticleDAO) *web.ArticleHandler {
	loggerV1 := InitLogger()
	db := InitDB()
	userDAO := dao.NewUserDAO(db)
	cmdable := InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(dao2, userRepository, articleCache)
	articleService := service.NewArticleService(articleRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService)
	return articleHandler
}

// wire.go:

var thirdPartySet = wire.NewSet(
	InitRedis, InitDB,
	InitLogger)

var userSvcProvider = wire.NewSet(dao.NewUserDAO, cache.NewUserCache, repository.NewCachedUserRepository, service.NewUserService)

var articlSvcProvider = wire.NewSet(repository.NewCachedArticleRepository, cache.NewArticleRedisCache, dao.NewArticleGORMDAO, service.NewArticleService)
