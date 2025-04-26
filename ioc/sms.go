package ioc

import (
	"ddd_demo/internal/service/sms"
	"ddd_demo/internal/service/sms/localsms"
)

func InitSMSService() sms.Service {
	return localsms.NewService()
	// 如果有需要，就可以用这个
	//return initTencentSMSService()
}
