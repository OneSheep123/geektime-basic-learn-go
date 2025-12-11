package events

import (
	"context"
	"ddd_demo/interactive/repository"
	"ddd_demo/internal/events/article"
	"ddd_demo/pkg/logger"
	"ddd_demo/pkg/samarax"
	"github.com/IBM/sarama"
	"time"
)

type InteractiveReadEventConsumer struct {
	repo   repository.InteractiveRepository
	client sarama.Client
	l      logger.LoggerV1
}

func NewInteractiveReadEventConsumer(repo repository.InteractiveRepository,
	client sarama.Client, l logger.LoggerV1) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{repo: repo, client: client, l: l}
}

//func (i *InteractiveReadEventConsumer) Start() error {
//	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
//	if err != nil {
//		return err
//	}
//	go func() {
//		er := cg.Consume(context.Background(),
//			[]string{TopicReadEvent},
//			samarax.NewBatchHandler[ReadEvent](i.l, i.BatchConsume))
//		if er != nil {
//			i.l.Error("退出消费", logger.Error(er))
//		}
//	}()
//	return err
//}

func (i *InteractiveReadEventConsumer) Start() error {
	// 使用带重试机制的消费者组创建
	var cg sarama.ConsumerGroup
	var err error

	// 重试创建消费者组
	for attempts := 0; attempts < 3; attempts++ {
		cg, err = sarama.NewConsumerGroupFromClient("interactive", i.client)
		if err == nil {
			break
		}
		i.l.Error("创建消费者组失败，重试中...",
			logger.Error(err),
			logger.Int("attempt", attempts+1))
		time.Sleep(time.Second * time.Duration(attempts+1))
	}

	if err != nil {
		return err
	}

	go func() {
		// 持续消费循环，处理消费过程中的错误
		for {
			i.l.Info("开始消费消息...")
			err := cg.Consume(context.Background(),
				[]string{article.TopicReadEvent},
				samarax.NewHandler[article.ReadEvent](i.l, i.Consume))

			if err != nil {
				i.l.Error("消费过程中出现错误", logger.Error(err))
				// 如果是协调器不可用等临时性错误，等待一段时间后重试
				if err.Error() == "kafka server: The coordinator is not available" {
					i.l.Info("检测到协调器不可用，等待后重试...")
					time.Sleep(5 * time.Second)
					continue
				}
				// 对于其他错误，记录并退出
				i.l.Error("退出消费", logger.Error(err))
				return
			}

			// 检查上下文是否被取消
			select {
			case <-context.Background().Done():
				i.l.Info("消费上下文已取消，退出消费")
				return
			default:
			}
		}
	}()
	return nil
}

func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage,
	event article.ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(ctx, "article", event.Aid)
}
