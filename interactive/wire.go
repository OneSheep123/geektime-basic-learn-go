//go:build wireinject

package main

import (
	"ddd_demo/interactive/events"
	"ddd_demo/interactive/grpc"
	"ddd_demo/interactive/ioc"
	repository2 "ddd_demo/interactive/repository"
	cache2 "ddd_demo/interactive/repository/cache"
	dao2 "ddd_demo/interactive/repository/dao"
	service2 "ddd_demo/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(ioc.InitDB,
	ioc.InitLogger,
	ioc.InitSaramaClient,
	ioc.InitRedis)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(thirdPartySet,
		interactiveSvcSet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
