package ratelimit

import (
	"context"
	"ddd_demo/internal/service/sms"
	smsmocks "ddd_demo/internal/service/sms/mocks"
	"ddd_demo/pkg/limiter"
	limitermocks "ddd_demo/pkg/limiter/mocks"
	"errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRateLimitSMSService_Send(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter)

		// 一个输入都没有

		// 预期输出
		wantErr error
	}{
		{
			name: "不限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)
				return svc, l
			},
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				return svc, l
			},
			wantErr: errLimited,
		},
		{
			name: "限流器错误",
			mock: func(ctrl *gomock.Controller) (sms.Service, limiter.Limiter) {
				svc := smsmocks.NewMockService(ctrl)
				l := limitermocks.NewMockLimiter(ctrl)
				l.EXPECT().Limit(gomock.Any(), gomock.Any()).
					Return(false, errors.New("redis限流器错误"))
				return svc, l
			},
			wantErr: errors.New("redis限流器错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			smsSvc, l := tc.mock(ctrl)
			svc := NewRateLimitSMSService(smsSvc, l)
			err := svc.Send(context.Background(), "abc",
				[]string{"123"}, "123456")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
