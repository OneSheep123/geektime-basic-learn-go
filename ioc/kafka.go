package ioc

import (
	events2 "ddd_demo/interactive/events"
	"ddd_demo/internal/events"
	"github.com/IBM/sarama"
	"github.com/spf13/viper"
	"time"
)

func InitSaramaClient() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	scfg := sarama.NewConfig()
	scfg.Version = sarama.V3_6_0_0 // 设置 Kafka 版本
	scfg.Producer.Return.Successes = true
	// 添加消费者组配置
	scfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	scfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	// 添加更多消费者组相关的配置
	scfg.Consumer.Group.Session.Timeout = 30 * time.Second
	scfg.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	scfg.Consumer.Group.Rebalance.Timeout = 60 * time.Second
	scfg.Consumer.Group.Rebalance.Retry.Max = 4
	scfg.Consumer.Group.Rebalance.Retry.Backoff = 2 * time.Second
	scfg.Metadata.Retry.Max = 4
	scfg.Metadata.Retry.Backoff = 500 * time.Millisecond
	scfg.Net.DialTimeout = 30 * time.Second
	scfg.Net.ReadTimeout = 30 * time.Second
	scfg.Net.WriteTimeout = 30 * time.Second

	client, err := sarama.NewClient(cfg.Addr, scfg)
	if err != nil {
		panic(err)
	}
	return client
}

func InitSyncProducer(c sarama.Client) sarama.SyncProducer {
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		panic(err)
	}
	return p
}

func InitConsumers(c1 *events2.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
