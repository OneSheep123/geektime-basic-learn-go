package service

import (
	"context"
	"ddd_demo/interactive/domain"
	"ddd_demo/interactive/repository"
	repomocks "ddd_demo/interactive/repository/mocks"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInteractiveService_IncrReadCnt(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.InteractiveRepository

		biz   string
		bizId int64

		wantErr error
	}{
		{
			name: "增加阅读计数成功",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().IncrReadCnt(gomock.Any(), "article", int64(1)).Return(nil)
				return repo
			},
			biz:     "article",
			bizId:   1,
			wantErr: nil,
		},
		{
			name: "增加阅读计数失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().IncrReadCnt(gomock.Any(), "article", int64(1)).
					Return(errors.New("db error"))
				return repo
			},
			biz:     "article",
			bizId:   1,
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewInteractiveService(repo)
			err := svc.IncrReadCnt(context.Background(), tc.biz, tc.bizId)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestInteractiveService_Like(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.InteractiveRepository

		biz string
		id  int64
		uid int64

		wantErr error
	}{
		{
			name: "点赞成功",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().IncrLike(gomock.Any(), "article", int64(1), int64(123)).Return(nil)
				return repo
			},
			biz:     "article",
			id:      1,
			uid:     123,
			wantErr: nil,
		},
		{
			name: "点赞失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().IncrLike(gomock.Any(), "article", int64(1), int64(123)).
					Return(errors.New("db error"))
				return repo
			},
			biz:     "article",
			id:      1,
			uid:     123,
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewInteractiveService(repo)
			err := svc.Like(context.Background(), tc.biz, tc.id, tc.uid)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestInteractiveService_CancelLike(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.InteractiveRepository

		biz string
		id  int64
		uid int64

		wantErr error
	}{
		{
			name: "取消点赞成功",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().DecrLike(gomock.Any(), "article", int64(1), int64(123)).Return(nil)
				return repo
			},
			biz:     "article",
			id:      1,
			uid:     123,
			wantErr: nil,
		},
		{
			name: "取消点赞失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().DecrLike(gomock.Any(), "article", int64(1), int64(123)).
					Return(errors.New("db error"))
				return repo
			},
			biz:     "article",
			id:      1,
			uid:     123,
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewInteractiveService(repo)
			err := svc.CancelLike(context.Background(), tc.biz, tc.id, tc.uid)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestInteractiveService_Collect(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.InteractiveRepository

		biz   string
		bizId int64
		cid   int64
		uid   int64

		wantErr error
	}{
		{
			name: "收藏成功",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().AddCollectionItem(gomock.Any(), "article", int64(1), int64(2), int64(123)).Return(nil)
				return repo
			},
			biz:     "article",
			bizId:   1,
			cid:     2,
			uid:     123,
			wantErr: nil,
		},
		{
			name: "收藏失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().AddCollectionItem(gomock.Any(), "article", int64(1), int64(2), int64(123)).
					Return(errors.New("db error"))
				return repo
			},
			biz:     "article",
			bizId:   1,
			cid:     2,
			uid:     123,
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewInteractiveService(repo)
			err := svc.Collect(context.Background(), tc.biz, tc.bizId, tc.cid, tc.uid)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestInteractiveService_Get(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.InteractiveRepository

		biz string
		id  int64
		uid int64

		wantInteractive domain.Interactive
		wantErr         error
	}{
		{
			name: "获取互动数据成功",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().Get(gomock.Any(), "article", int64(1)).
					Return(domain.Interactive{
						BizId:      1,
						ReadCnt:    10,
						LikeCnt:    5,
						CollectCnt: 3,
					}, nil)
				repo.EXPECT().Liked(gomock.Any(), "article", int64(1), int64(123)).
					Return(true, nil)
				repo.EXPECT().Collected(gomock.Any(), "article", int64(1), int64(123)).
					Return(false, nil)
				return repo
			},
			biz: "article",
			id:  1,
			uid: 123,
			wantInteractive: domain.Interactive{
				BizId:      1,
				ReadCnt:    10,
				LikeCnt:    5,
				CollectCnt: 3,
				Liked:      true,
				Collected:  false,
			},
			wantErr: nil,
		},
		{
			name: "获取基础互动数据失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().Get(gomock.Any(), "article", int64(1)).
					Return(domain.Interactive{}, errors.New("db error"))
				return repo
			},
			biz:             "article",
			id:              1,
			uid:             123,
			wantInteractive: domain.Interactive{},
			wantErr:         errors.New("db error"),
		},
		{
			name: "获取点赞状态失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().Get(gomock.Any(), "article", int64(1)).
					Return(domain.Interactive{
						BizId:      1,
						ReadCnt:    10,
						LikeCnt:    5,
						CollectCnt: 3,
					}, nil)
				repo.EXPECT().Liked(gomock.Any(), "article", int64(1), int64(123)).
					Return(false, errors.New("db error"))
				repo.EXPECT().Collected(gomock.Any(), "article", int64(1), int64(123)).
					Return(false, nil)
				return repo
			},
			biz: "article",
			id:  1,
			uid: 123,
			wantInteractive: domain.Interactive{
				BizId:      1,
				ReadCnt:    10,
				LikeCnt:    5,
				CollectCnt: 3,
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "获取收藏状态失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().Get(gomock.Any(), "article", int64(1)).
					Return(domain.Interactive{
						BizId:      1,
						ReadCnt:    10,
						LikeCnt:    5,
						CollectCnt: 3,
					}, nil)
				repo.EXPECT().Liked(gomock.Any(), "article", int64(1), int64(123)).
					Return(true, nil)
				repo.EXPECT().Collected(gomock.Any(), "article", int64(1), int64(123)).
					Return(false, errors.New("db error"))
				return repo
			},
			biz: "article",
			id:  1,
			uid: 123,
			wantInteractive: domain.Interactive{
				BizId:      1,
				ReadCnt:    10,
				LikeCnt:    5,
				CollectCnt: 3,
				Liked:      true,
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewInteractiveService(repo)
			interactive, err := svc.Get(context.Background(), tc.biz, tc.id, tc.uid)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantInteractive, interactive)
		})
	}
}

func TestInteractiveService_GetByIds(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.InteractiveRepository

		biz string
		ids []int64

		wantMap map[int64]domain.Interactive
		wantErr error
	}{
		{
			name: "批量获取互动数据成功",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).
					Return([]domain.Interactive{
						{
							BizId:      1,
							ReadCnt:    10,
							LikeCnt:    5,
							CollectCnt: 3,
						},
						{
							BizId:      2,
							ReadCnt:    8,
							LikeCnt:    2,
							CollectCnt: 1,
						},
						{
							BizId:      3,
							ReadCnt:    15,
							LikeCnt:    7,
							CollectCnt: 4,
						},
					}, nil)
				return repo
			},
			biz: "article",
			ids: []int64{1, 2, 3},
			wantMap: map[int64]domain.Interactive{
				1: {
					BizId:      1,
					ReadCnt:    10,
					LikeCnt:    5,
					CollectCnt: 3,
				},
				2: {
					BizId:      2,
					ReadCnt:    8,
					LikeCnt:    2,
					CollectCnt: 1,
				},
				3: {
					BizId:      3,
					ReadCnt:    15,
					LikeCnt:    7,
					CollectCnt: 4,
				},
			},
			wantErr: nil,
		},
		{
			name: "批量获取互动数据失败",
			mock: func(ctrl *gomock.Controller) repository.InteractiveRepository {
				repo := repomocks.NewMockInteractiveRepository(ctrl)
				repo.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2, 3}).
					Return(nil, errors.New("db error"))
				return repo
			},
			biz:     "article",
			ids:     []int64{1, 2, 3},
			wantMap: nil,
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repo := tc.mock(ctrl)
			svc := NewInteractiveService(repo)
			result, err := svc.GetByIds(context.Background(), tc.biz, tc.ids)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantMap, result)
		})
	}
}
