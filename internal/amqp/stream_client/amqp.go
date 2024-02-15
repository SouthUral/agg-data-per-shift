package streamclient

import (
	"fmt"
	"time"

	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"

	log "github.com/sirupsen/logrus"
)

// "rabbitmq-stream://guest:guest@host3:5552/%2f"

func InitRabbit(url, nameConsumer, nameStream string) *Rabbit {
	res := &Rabbit{
		url:          url,
		nameConsumer: nameConsumer,
		nameStream:   nameStream,
	}

	err := res.connRabbit()
	if err != nil {
		log.Error(err)
		return res
	}

	err = res.createConsumer()
	if err != nil {
		err = fmt.Errorf("%w: %w", connectStreamError{}, err)
		log.Error(err)
		return res
	}
	return res
}

type Rabbit struct {
	url          string
	nameConsumer string
	streamEnv    *stream.Environment
	consumer     *stream.Consumer
	nameStream   string
}

func (r *Rabbit) connRabbit() error {
	var err error
	var env *stream.Environment
	env, err = stream.NewEnvironment(stream.NewEnvironmentOptions().SetUri(r.url))
	if err != nil {
		log.Error(env)
		return err
	}

	r.streamEnv = env
	return err
}

// метод получения последнего  offset из stream
func (r *Rabbit) getStreamLastOffset() int64 {
	// stats, err := environment.StreamStats(testStreamName)
	var offset int64

	stats, err := r.streamEnv.StreamStats(r.nameStream)
	if err != nil {
		log.Error(err)
		return offset
	}

	offset, err = stats.LastOffset()
	if err != nil {
		log.Error(err)
	}

	return offset
}

func (r *Rabbit) createConsumer() error {
	var err error

	handleMessages := func(consumerContext stream.ConsumerContext, message *amqp.Message) {
		log.Infof("consumer name: %s, text: %s \n", consumerContext.Consumer.GetName(), message.Data)
		offset := consumerContext.Consumer.GetOffset()
		log.Infof("offset %d", offset)
		consumerContext.Consumer.StoreOffset()
		time.Sleep(5 * time.Second)
	}

	consumer, err := r.streamEnv.NewConsumer(
		r.nameStream,
		handleMessages,
		stream.NewConsumerOptions().
			SetConsumerName(r.nameConsumer).
			SetAutoCommit(stream.NewAutoCommitStrategy().SetCountBeforeStorage(50).SetFlushInterval(10*time.Second)).
			SetOffset(stream.OffsetSpecification{}.First()).
			SetCRCCheck(false))

	if err != nil {
		log.Error(err)
		return err
	}

	r.consumer = consumer
	return err
}

func (r *Rabbit) CheckConn() bool {
	return r.streamEnv.IsClosed()
}

func (r *Rabbit) Shutdown() {
	err := r.consumer.Close()
	if err != nil {
		log.Error(err)
	}

	r.streamEnv.Close()
	if err != nil {
		log.Error(err)
	}
}
