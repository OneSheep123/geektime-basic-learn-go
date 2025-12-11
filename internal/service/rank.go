package service

import (
	"context"
	intrv1 "ddd_demo/api/proto/gen/intr/v1"
	"ddd_demo/internal/domain"
	"ddd_demo/internal/repository"
	"errors"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
)

//go:generate mockgen -source=./rank.go -package=svcmocks -destination=./mocks/rank.mock.go RankService
type RankService interface {
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	intrSvc   intrv1.InteractiveServiceClient
	artSvc    ArticleService
	batchSize int
	scoreFunc func(likeCnt int64, utime time.Time) float64
	n         int // topN堆容量

	repo repository.RankingRepository
}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}

func NewBatchRankingService(intrSvc intrv1.InteractiveServiceClient, artSvc ArticleService, repo repository.RankingRepository) RankService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			// 时间
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 最终是要放到缓存里面的
	// 存到缓存里面
	return b.repo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}

	// 小顶堆
	topN := queue.NewPriorityQueue[Score](b.n, func(a, b Score) int {
		if a.score > b.score {
			return 1
		}
		if a.score < b.score {
			return -1
		}
		return 0
	})

	for {
		// 取数据
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		//if len(arts) == 0 {
		//	break
		//}
		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		// 取点赞数
		intrResp, err := b.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article", Ids: ids,
		})
		if err != nil {
			return nil, err
		}
		intrMap := intrResp.Intrs
		for _, art := range arts {
			intr := intrMap[art.Id]
			//intr, ok := intrMap[art.Id]
			//if !ok {
			//	continue
			//}
			score := b.scoreFunc(intr.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			err = topN.Enqueue(ele)
			if errors.Is(err, queue.ErrOutOfCapacity) {
				// 这个也是满了
				// 拿出最小的元素
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(minEle)
				}
			}
		}
		offset = offset + len(arts)
		// 没有取够一批，我们就直接中断执行
		// 没有下一批了
		if len(arts) < b.batchSize ||
			// 这个是一个优化
			arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}

	// 这边 topN 里面就是最终结果
	res := make([]domain.Article, topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
