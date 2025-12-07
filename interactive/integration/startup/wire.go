//go:build wireinject

package startup

import (
	"ddd_demo/interactive/grpc"
	repository2 "ddd_demo/interactive/repository"
	cache2 "ddd_demo/interactive/repository/cache"
	dao2 "ddd_demo/interactive/repository/dao"
	service2 "ddd_demo/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	//InitSaramaClient,
	//InitSyncProducer,
	InitLogger,
)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveSvcSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
}
