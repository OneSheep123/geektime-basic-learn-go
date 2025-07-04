package ioc

import (
	"ddd_demo/internal/service/oauth2/wechat"
	"ddd_demo/pkg/logger"
	"go.uber.org/zap"
	"os"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		zap.L().Error("找不到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		zap.L().Error("找不到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appID, appSecret, l)
}
