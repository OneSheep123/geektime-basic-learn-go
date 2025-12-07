package startup

import "ddd_demo/pkg/logger"

func InitLogger() logger.LoggerV1 {
	return logger.NewNopLogger()
}
