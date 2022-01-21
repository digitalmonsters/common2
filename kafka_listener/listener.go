package kafka_listener

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.elastic.co/apm"
	"io"
	"sync"
	"time"
)

var readerMutex sync.Mutex

type kafkaListener struct {
	cfg                 boilerplate.KafkaListenerConfiguration
	ctx                 context.Context
	readers             map[int]*kafka.Reader // key is partition; 0 - for GroupId
	targetTopic         string
	command             ICommand
	listenerName        string
	cancelFn            context.CancelFunc
	hasRunningRequest   bool
	dialer              *kafka.Dialer
	isConsumerGroupMode bool
}

func newKafkaListener(config boilerplate.KafkaListenerConfiguration, ctx context.Context, command ICommand) *kafkaListener {
	if len(config.Topic) == 0 {
		panic("kafka topic should not be empty")
	}

	if config.MaxBytes == 0 {
		config.MaxBytes = 10e6 // 10 MB
	}

	if config.MaxBackOffTimeMilliseconds <= 0 {
		config.MaxBackOffTimeMilliseconds = 60000 // 60sec
	}

	if config.BackOffTimeIntervalMilliseconds <= 0 {
		config.BackOffTimeIntervalMilliseconds = 1000 // 1s
	}

	if config.KafkaAuth == nil {
		config.KafkaAuth = &boilerplate.KafkaAuth{}
	}

	dialer, err := getKafkaDialer(config.Tls, *config.KafkaAuth)

	if err != nil {
		panic(err)
	}

	localCtx, cancelFn := context.WithCancel(ctx)

	return &kafkaListener{
		cfg:                 config,
		ctx:                 localCtx,
		cancelFn:            cancelFn,
		targetTopic:         config.Topic,
		command:             command,
		dialer:              dialer,
		isConsumerGroupMode: len(config.GroupId) > 0,
		readers:             map[int]*kafka.Reader{},
		listenerName:        fmt.Sprintf("kafka_listener_%v", config.Topic),
	}
}

func (k kafkaListener) GetTopic() string {
	return k.targetTopic
}

func (k *kafkaListener) getPartitionsForTopic() ([]int, error) {
	if k.isConsumerGroupMode {
		return []int{0}, nil // 0 means that we dont care as we have GroupId
	}

	var finalPartitions []int

	for _, host := range boilerplate.SplitHostsToSlice(k.cfg.Hosts) {
		con, err := k.dialer.Dial("tcp", host)

		if err != nil {
			log.Err(err).Msgf("can not get connection to calculate partitions for topic %v", k.cfg.Topic)
			continue
		}

		partitions, err := con.ReadPartitions(k.cfg.Topic)

		if err != nil {
			log.Err(err).Msgf("can not get partitions for topic %v", k.cfg.Topic)
			_ = con.Close()
			continue
		}

		for _, p := range partitions {
			finalPartitions = append(finalPartitions, p.ID)
		}

		_ = con.Close()
	}

	if len(finalPartitions) == 0 {
		return nil, errors.New(fmt.Sprintf("no partitions found for topic %v", k.cfg.Topic))
	}

	return finalPartitions, nil
}

func (k *kafkaListener) checkIfTopicExists(topic string) error {
	transport := kafka.DefaultTransport

	if k.cfg.Tls {
		dialer := &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		}

		dialer.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}

		transport = &kafka.Transport{
			TLS: &tls.Config{
				InsecureSkipVerify: true,
			},
			Dial: dialer.DialFunc,
		}
	}

	client := &kafka.Client{
		Transport: transport,
	}

	tcp := kafka.TCP(boilerplate.SplitHostsToSlice(k.cfg.Hosts)...)

	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)

	defer func() {
		cancel()
	}()

	meta, err := client.Metadata(ctx, &kafka.MetadataRequest{
		Addr: tcp,
	})

	if err != nil {
		return err
	}

	var exists bool
	for _, t := range meta.Topics {
		if t.Name == topic {
			exists = true
			break
		}
	}

	if !exists {
		return errors.New(fmt.Sprintf("topic [%v] doesn't exist", topic))
	}

	return nil
}

func (k *kafkaListener) getReaderForPartition(partition int) (*kafka.Reader, error) {
	readerMutex.Lock()
	defer readerMutex.Unlock()

	if v, ok := k.readers[partition]; ok {
		return v, nil
	}

	var auth boilerplate.KafkaAuth

	if k.cfg.KafkaAuth != nil {
		auth = *k.cfg.KafkaAuth
	}

	dialer, err := getKafkaDialer(k.cfg.Tls, auth)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	var kafkaCfg = kafka.ReaderConfig{
		Brokers:        boilerplate.SplitHostsToSlice(k.cfg.Hosts),
		GroupID:        k.cfg.GroupId,
		Partition:      partition, // if GroupId
		Topic:          k.targetTopic,
		MinBytes:       k.cfg.MinBytes,
		MaxBytes:       k.cfg.MaxBytes,
		CommitInterval: time.Millisecond,
		Dialer:         dialer,
	}

	r := kafka.NewReader(kafkaCfg)

	k.readers[partition] = r

	return r, nil
}

func (k *kafkaListener) ListenInBatches(maxBatchSize int, maxDuration time.Duration) {
	var partitions []int
	var err error
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = time.Duration(k.cfg.MaxBackOffTimeMilliseconds) * time.Millisecond
	b.InitialInterval = time.Duration(k.cfg.BackOffTimeIntervalMilliseconds) * time.Millisecond

	for k.ctx.Err() == nil {
		if err := k.checkIfTopicExists(k.targetTopic); err != nil {
			log.Err(err).Msgf("listener [%v]. Topic [%v] does not exists. waiting for topic to be available with interval 10s",
				k.listenerName, k.targetTopic)

			transaction := apm_helper.StartNewApmTransaction("start-kafka-listener", "kafka", nil, nil)
			apm_helper.AddApmLabel(transaction, "topic", k.targetTopic)
			apm_helper.CaptureApmError(err, transaction)

			transaction.End()
			duration := b.NextBackOff()
			if duration == b.Stop {
				break
			}
			time.Sleep(duration)
			continue
		}

		partitions, err = k.getPartitionsForTopic()

		if err != nil {
			log.Err(err).Send()

			time.Sleep(b.NextBackOff())
		}

		if true { // fck linter
			break
		}
	}

	if k.isConsumerGroupMode {
		partitions = []int{0}
	}

	for _, partition := range partitions {
		p := partition
		bPartitions := backoff.NewExponentialBackOff()
		bPartitions.MaxElapsedTime = time.Duration(k.cfg.MaxBackOffTimeMilliseconds) * time.Millisecond
		bPartitions.InitialInterval = time.Duration(k.cfg.BackOffTimeIntervalMilliseconds) * time.Millisecond
		go func() {
			firstRun := true
			for k.ctx.Err() == nil {
				reader, err := k.getReaderForPartition(p)

				if err != nil {
					log.Err(err).Send()
					duration := bPartitions.NextBackOff()
					if duration == bPartitions.Stop {
						break
					}
					time.Sleep(duration)
					continue
				}

				if !k.isConsumerGroupMode && firstRun { // then lets read only new messages from this point
					if err := reader.SetOffsetAt(k.ctx, time.Now().UTC()); err != nil {
						log.Err(err).Send()
					}
				}

				firstRun = false

				if err := k.listen(maxBatchSize, maxDuration, reader); err != nil {
					//if len(k.cfg.GroupId) > 0 {
					//	k.closeReader(p) // reset to last position
					//}

					tx := apm_helper.StartNewApmTransaction(k.listenerName, "kafka_listener", nil, nil)

					apm_helper.CaptureApmError(err, tx)
					log.Err(err).Send()

					tx.End()
					duration := bPartitions.NextBackOff()
					if duration == bPartitions.Stop {
						break
					}
					time.Sleep(duration)
				}
			}
		}()
	}
}

func (k *kafkaListener) closeReader(partitionId int) {
	readerMutex.Lock()
	defer readerMutex.Unlock()

	if v := k.readers[partitionId]; v != nil {
		_ = v.Close()
	}

	delete(k.readers, partitionId)
}

func (k *kafkaListener) Close() error {
	k.cancelFn()

	runningReq := false

	if k.hasRunningRequest {
		runningReq = true

		for i := 1; i < 5; i++ {
			if !k.hasRunningRequest {
				runningReq = false
				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	for partitionId := range k.readers {
		k.closeReader(partitionId)
	}

	if runningReq {
		return errors.New("kafka listener still has running requests")
	}

	return nil
}

func (k *kafkaListener) listen(maxBatchSize int, maxDuration time.Duration, reader *kafka.Reader) error {
	messagePool := make([]kafka.Message, maxBatchSize)

	messageIndex := 0
	var successfullyProcessedMessages []kafka.Message

	listenCtx, cancel := context.WithCancel(k.ctx)

	defer func() {
		cancel()
	}()

	for listenCtx.Err() == nil {
		message2, err := reader.FetchMessage(listenCtx)

		apmTransaction := apm_helper.StartNewApmTransaction(k.listenerName, "kafka_listener", nil,
			nil)

		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				apmTransaction.Discard()
				return err
			}

			return err
		}

		k.hasRunningRequest = true

		messagePool[0] = message2
		messageIndex = 1

		if maxBatchSize > 1 {
			innerCtx, innerCancelFn := context.WithTimeout(listenCtx, maxDuration)
			kafkaReadSpan := apmTransaction.StartSpan(fmt.Sprintf("kafka batching [%v]", k.cfg.Topic),
				"kafka", nil)

			kafkaReadSpan.Context.SetDestinationService(apm.DestinationServiceSpanContext{
				Name:     "kafka",
				Resource: k.targetTopic,
			})
			kafkaReadSpan.Context.SetMessage(apm.MessageSpanContext{QueueName: k.cfg.Topic})

			for innerCtx.Err() == nil {
				message1, err1 := reader.FetchMessage(innerCtx)

				if err1 == context.DeadlineExceeded {
					break
				}

				if err1 != nil {
					if errors.Is(err1, io.EOF) {
						innerCancelFn()
						k.hasRunningRequest = false

						kafkaReadSpan.End()
						return err1
					}
					log.Err(err1).Send()
				}

				if err1 == nil {
					messagePool[messageIndex] = message1
					messageIndex += 1
				}

				if messageIndex >= maxBatchSize {
					innerCancelFn()
				}
			}

			innerCancelFn()

			kafkaReadSpan.Context.SetLabel("message_count", messageIndex)
			kafkaReadSpan.End()

			apm_helper.AddApmData(apmTransaction, "messages_count", messageIndex)
		}

		if k.ctx.Err() != nil {
			apmTransaction.Discard()
			k.hasRunningRequest = false
			break // discard messages
		}

		commandExecutionContext := apm.ContextWithTransaction(listenCtx, apmTransaction)

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = time.Duration(k.cfg.MaxBackOffTimeMilliseconds) * time.Millisecond
		b.InitialInterval = time.Duration(k.cfg.BackOffTimeIntervalMilliseconds) * time.Millisecond

		b.Reset()

		retryCount := 0

		messagesToProcess := messagePool[:messageIndex]

		requestProcessingErrors := backoff.Retry(func() error {
			retryCount += 1

			processingSpan := apmTransaction.StartSpan(fmt.Sprintf("%v with retry #%v", k.command.GetFancyName(),
				retryCount), "processing", nil)

			processingSpan.Context.SetLabel("messages_to_process", len(messagesToProcess))

			successfullyProcessedMessages = k.command.Execute(ExecutionData{
				ApmTransaction: apmTransaction,
				Context:        commandExecutionContext,
			}, messagesToProcess...)

			processingSpan.Context.SetLabel("successfully_processed_messages", len(successfullyProcessedMessages))

			processingSpan.End()

			if err = k.commitMessages(successfullyProcessedMessages, apmTransaction,
				reader, commandExecutionContext); err != nil {
				return &backoff.PermanentError{Err: err}
			}

			allProcessedMessages := map[string]struct{}{}

			for _, m := range successfullyProcessedMessages {
				allProcessedMessages[extractKeyFromKafkaMessage(m)] = struct{}{}
			}

			if len(allProcessedMessages) == len(messagesToProcess) { // if unique key count equals to message count, then we think that its ok
				return nil // awesome, most of the scenarios ends here, no errors
			}

			// else messages are processed partially

			nextMessagesToProcess := make([]kafka.Message, 0)

			for _, incoming := range messagesToProcess {
				if _, ok := allProcessedMessages[extractKeyFromKafkaMessage(incoming)]; !ok {
					nextMessagesToProcess = append(nextMessagesToProcess, incoming)
				}
			}

			messagesToProcess = nextMessagesToProcess

			if len(messagesToProcess) > 0 {
				return errors.New("there are messages to process")
			}

			return nil
		}, b)

		if requestProcessingErrors != nil { // it`s a permanent error, we should try to commit all messages which we had
			if err = k.commitMessages(messagePool[:messageIndex], apmTransaction, reader, commandExecutionContext); err != nil { // we have no power here
				apm_helper.CaptureApmError(errors.Wrap(err, "can not commit messages after retry policy"), apmTransaction)
			}
		}

		k.hasRunningRequest = false
		apmTransaction.End()
	}

	k.hasRunningRequest = false

	return nil
}

func (k *kafkaListener) commitMessages(messages []kafka.Message, apmTransaction *apm.Transaction,
	reader *kafka.Reader, ctx context.Context) error {
	if !k.isConsumerGroupMode || len(messages) == 0 {
		return nil
	}

	kafkaCommitSpan := apmTransaction.StartSpan(fmt.Sprintf("kafka commit [%v]",
		k.cfg.Topic), "kafka", nil)

	kafkaCommitSpan.Context.SetMessage(apm.MessageSpanContext{QueueName: k.cfg.Topic})
	kafkaCommitSpan.Context.SetDestinationService(apm.DestinationServiceSpanContext{
		Name:     "kafka",
		Resource: k.cfg.Topic,
	})

	kafkaCommitSpan.Context.SetLabel("count", len(messages))

	if err := reader.CommitMessages(ctx, messages...); err != nil {
		apm_helper.CaptureApmError(err, apmTransaction)

		kafkaCommitSpan.End()

		return errors.WithStack(err)
	}

	kafkaCommitSpan.End()

	return nil
}

func extractKeyFromKafkaMessage(message kafka.Message) string {
	return fmt.Sprintf("%v_%v", message.Partition, message.Offset)
}
