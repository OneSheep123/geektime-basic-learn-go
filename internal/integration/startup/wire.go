//go:build wireinject

package startup

import (
	repository2 "ddd_demo/interactive/repository"
	cache2 "ddd_demo/interactive/repository/cache"
	dao2 "ddd_demo/interactive/repository/dao"
	service2 "ddd_demo/interactive/service"
	"ddd_demo/internal/events/article"
	"ddd_demo/internal/repository"
	"ddd_demo/internal/repository/cache"
	"ddd_demo/internal/repository/dao"
	"ddd_demo/internal/service"
	"ddd_demo/internal/service/sms"
	"ddd_demo/internal/service/sms/async"
	"ddd_demo/internal/web"
	ijwt "ddd_demo/internal/web/jwt"
	"ddd_demo/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitSaramaClient,
	InitSyncProducer,
	InitLogger)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	repository.NewCachedArticleRepository,
	cache.NewArticleRedisCache,
	dao.NewArticleGORMDAO,
	service.NewArticleService)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		articlSvcProvider,
		interactiveSvcSet,
		// cache 部分
		cache.NewCodeCache,

		// repository 部分
		repository.NewCodeRepository,

		article.NewSaramaSyncProducer,

		// Service 部分
		ioc.InitSMSService,
		service.NewCodeService,
		InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitAsyncSmsService(svc sms.Service) *async.Service {
	wire.Build(thirdPartySet, repository.NewAsyncSMSRepository,
		dao.NewGORMAsyncSmsDAO,
		async.NewService,
	)
	return &async.Service{}
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		userSvcProvider,
		interactiveSvcSet,
		repository.NewCachedArticleRepository,
		cache.NewArticleRedisCache,
		service.NewArticleService,
		article.NewSaramaSyncProducer,
		web.NewArticleHandler)
	return &web.ArticleHandler{}
}
