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

var thirdPartySet = wire.NewSet(ioc.InitSrcDB,
	ioc.InitDstDB,
	ioc.InitDoubleWritePool,
	// 由单写替换为双写的DB
	ioc.InitBizDB,
	ioc.InitLogger,
	ioc.InitSaramaClient,
	ioc.InitSaramaSyncProducer,
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
		ioc.InitInteractiveProducer,
		ioc.InitFixerConsumer,
		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		ioc.InitGinxServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
