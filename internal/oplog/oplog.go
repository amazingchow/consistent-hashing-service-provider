package oplog

import (
	"context"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	consistenthashing "github.com/amazingchow/photon-dance-consistent-hashing/internal/ch"
	pb_api "github.com/amazingchow/photon-dance-consistent-hashing/pb/api"
)

type OpLogger struct {
	ctx    context.Context
	oplogs chan *pb_api.OpLogEntry
	topic  string

	consumerGroup sarama.ConsumerGroup
	consumer      *ConsumerImpl
	producer      sarama.AsyncProducer

	overfeed int64

	executor *consistenthashing.Executor
}

func NewOpLogger(ctx context.Context, topic string, executor *consistenthashing.Executor) *OpLogger {
	oplogger := OpLogger{
		ctx:      ctx,
		oplogs:   make(chan *pb_api.OpLogEntry, 4096),
		topic:    topic,
		overfeed: 0,
		executor: executor,
	}
	log.Info().Msgf("init oplogger successfully, topic: %s", topic)
	return &oplogger
}

func (olr *OpLogger) Close() {
	close(olr.oplogs)
	if olr.consumerGroup != nil {
		olr.consumerGroup.Close()
	}
	log.Info().Msgf("oplogger quits, topic: %s", olr.topic)
}

func (olr *OpLogger) SetUpConsumer(endpoints []string, groupID string) {
	var err error

	kafkaConfig := sarama.NewConfig()
	GetKafkaAccessEnv(kafkaConfig)
	hn, _ := os.Hostname()
	kafkaConfig.ClientID = hn
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	kafkaConfig.Consumer.Group.Session.Timeout = 20 * time.Second
	kafkaConfig.Consumer.Group.Heartbeat.Interval = 6 * time.Second
	kafkaConfig.Consumer.MaxProcessingTime = 500 * time.Millisecond

	olr.consumerGroup, err = sarama.NewConsumerGroup(endpoints, groupID, kafkaConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup oplogger's consumer group")
	}
	olr.consumer = &ConsumerImpl{oplogger: olr}
	go func() {
		olr.consume()
	}()

	log.Info().Msg("setup oplogger's consumer group successfully")
}

func (olr *OpLogger) consume() {
CONSUME_LOOP:
	for {
		if err := olr.consumerGroup.Consume(olr.ctx, []string{olr.topic}, olr.consumer); err != nil {
			log.Error().Err(err).Msg("oplogger failed to consume msg")
			time.Sleep(time.Second * 5)
		}
		select {
		case <-olr.ctx.Done():
			{
				log.Warn().Msg("oplogger's consume loop has been closed")
				break CONSUME_LOOP
			}
		case err, ok := <-olr.consumerGroup.Errors():
			{
				if ok {
					log.Error().Err(err).Msg("oplogger failed to consume msg")
					time.Sleep(time.Second * 5)
				}
			}
		default:
			{
				log.Warn().Msg("oplogger's comsumer group rebalances")
			}
		}
	}
}

func (olr *OpLogger) cfunc(data []byte, topic string) error {
	var oplog = new(pb_api.OpLogEntry)
	proto.Unmarshal(data, oplog) // nolint

	switch oplog.GetOperationType() {
	case pb_api.OperationType_OPERATION_TYPE_ADD:
		{
			log.Debug().Msg("consume oplog OPERATION_TYPE_ADD_SHARD")
			var node = new(pb_api.Node)
			proto.Unmarshal(oplog.GetPayload(), node) // nolint
			olr.executor.Join(node)                   // nolint
		}
	case pb_api.OperationType_OPERATION_TYPE_REMOVE:
		{
			log.Debug().Msg("consume oplog OPERATION_TYPE_REMOVE_SHARD")
			olr.executor.Leave(string(oplog.GetPayload())) // nolint
		}
	default:
		{
			log.Error().Msgf("invalid oplog type %v", oplog.GetOperationType())
		}
	}
	return nil
}

type ConsumerImpl struct {
	oplogger *OpLogger
}

// Setup runs at the beginning of a session before ConsumeClaim.
func (consumer *ConsumerImpl) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup runs at the end of a session, once all ConsumeClaim goroutines have exited.
func (consumer *ConsumerImpl) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *ConsumerImpl) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
CONSUME_LOOP:
	for {
		select {
		case msg, ok := <-claim.Messages():
			{
				if !ok {
					break CONSUME_LOOP
				}
				if err := consumer.oplogger.cfunc(msg.Value, consumer.oplogger.topic); err != nil {
					return err
				} else {
					session.MarkMessage(msg, "")
				}
			}
		case <-consumer.oplogger.ctx.Done():
			{
				break CONSUME_LOOP
			}
		}
	}
	return nil
}

func (olr *OpLogger) SetUpProducer(endpoints []string) {
	var err error

	kafkaConfig := sarama.NewConfig()
	GetKafkaAccessEnv(kafkaConfig)
	kafkaConfig.Producer.Flush.MaxMessages = 100
	kafkaConfig.Producer.Flush.Frequency = time.Millisecond * 500
	kafkaConfig.Producer.Partitioner = sarama.NewHashPartitioner

	olr.producer, err = sarama.NewAsyncProducer(endpoints, kafkaConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup oplogger's producer")
	}
	go func() {
		olr.produce()
	}()

	log.Info().Msg("setup oplogger's producer successfully")
}

func (olr *OpLogger) produce() {
PRODUCE_LOOP:
	for {
		select {
		case msg, ok := <-olr.oplogs:
			{
				if !ok {
					log.Error().Msg("oplogger's msg channel has been closed")
					break PRODUCE_LOOP
				}
				if err := olr.pfunc(olr.producer.Input(), msg, olr.topic); err != nil {
					log.Error().Err(err).Msg("oplogger failed to produce msg")
				}
			}
		case err, ok := <-olr.producer.Errors():
			{
				if ok {
					log.Error().Err(err).Msg("oplogger failed to produce msg")
				}
			}
		case <-olr.ctx.Done():
			{
				log.Warn().Msg("oplogger's produce loop has been closed")
				break PRODUCE_LOOP
			}
		}
	}
}

func (olr *OpLogger) pfunc(msgChan chan<- *sarama.ProducerMessage, oplog *pb_api.OpLogEntry, topic string) error {
	bytes, _ := proto.Marshal(oplog)
	switch oplog.GetOperationType() {
	case pb_api.OperationType_OPERATION_TYPE_ADD:
		{
			msgChan <- &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(strconv.Itoa(int(pb_api.OperationType_OPERATION_TYPE_ADD))),
				Value: sarama.ByteEncoder(bytes),
			}
			log.Debug().Msg("produce oplog OPERATION_TYPE_ADD_SHARD")
		}
	case pb_api.OperationType_OPERATION_TYPE_REMOVE:
		{
			msgChan <- &sarama.ProducerMessage{
				Topic: topic,
				Key:   sarama.StringEncoder(strconv.Itoa(int(pb_api.OperationType_OPERATION_TYPE_REMOVE))),
				Value: sarama.ByteEncoder(bytes),
			}
			log.Debug().Msg("produce oplog OPERATION_TYPE_REMOVE_SHARD")
		}
	default:
		{
			log.Error().Msgf("invalid oplog type %v", oplog.GetOperationType())
		}
	}
	return nil
}

func (olr *OpLogger) SyncAddOpLog(oplog *pb_api.OpLogEntry) {
	select {
	case olr.oplogs <- oplog:
	default:
		atomic.AddInt64(&(olr.overfeed), 1)
		log.Warn().Msgf("oplogger's producer works too fast, overflow msg: %d", atomic.LoadInt64(&(olr.overfeed)))
	}
}
