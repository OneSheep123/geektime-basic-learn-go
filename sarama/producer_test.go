package sarama

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

var addr = []string{"localhost:9094"}

func TestSyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner

	// 添加时间戳相关配置
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	// 设置时间戳类型
	cfg.Version = sarama.V3_6_0_0

	producer, err := sarama.NewSyncProducer(addr, cfg)
	assert.NoError(t, err)

	for i := 0; i < 100; i++ {
		_, _, er := producer.SendMessage(&sarama.ProducerMessage{
			Topic:     "test_topic",
			Value:     sarama.StringEncoder("这是一条消息"),
			Timestamp: time.Now(), // 显式设置时间戳
			// 会在生产者和消费者之间传递的
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("key1"),
					Value: []byte("value1"),
				},
			},
			Metadata: "这是 metadata",
		})
		require.NoError(t, er)
	}
}

func TestAsyncProducer(t *testing.T) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true
	cfg.Version = sarama.V3_6_0_0 // 设置版本

	producer, err := sarama.NewAsyncProducer(addr, cfg)
	assert.NoError(t, err)
	msgs := producer.Input()
	msgs <- &sarama.ProducerMessage{
		Topic:     "test_topic",
		Value:     sarama.StringEncoder("这是一条消息"),
		Timestamp: time.Now(), // 显式设置时间戳
		// 会在生产者和消费者之间传递的
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),
				Value: []byte("value1"),
			},
		},
		Metadata: "这是 metadata",
	}

	select {
	case msg := <-producer.Successes():
		t.Log("发送成功", string(msg.Value.(sarama.StringEncoder)))
	case err := <-producer.Errors():
		t.Log("发送失败", err.Err, err.Msg)
	}
}
