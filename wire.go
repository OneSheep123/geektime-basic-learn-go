//go:build wireinject

package main

import (
	"ddd_demo/interactive/events"
	repository2 "ddd_demo/interactive/repository"
	cache2 "ddd_demo/interactive/repository/cache"
	dao2 "ddd_demo/interactive/repository/dao"
	service2 "ddd_demo/interactive/service"
	"ddd_demo/internal/events/article"
	"ddd_demo/internal/repository"
	"ddd_demo/internal/repository/cache"
	"ddd_demo/internal/repository/dao"
	"ddd_demo/internal/service"
	"ddd_demo/internal/web"
	ijwt "ddd_demo/internal/web/jwt"
	"ddd_demo/ioc"

	"github.com/google/wire"
)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		ioc.InitRlockClient,
		// DAO 部分
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,

		interactiveSvcSet,
		rankingSvcSet,
		ioc.InitRankingJob,
		ioc.InitJobs,

		article.NewSaramaSyncProducer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,
		cache.NewArticleRedisCache,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		//告诉 Wire 将所有依赖注入到 App 结构体的字段中
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
