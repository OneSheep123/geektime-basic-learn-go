package startup

import (
	"ddd_demo/internal/service/oauth2/wechat"
	"ddd_demo/pkg/logger"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	return wechat.NewService("", "", l)
}
