package eventbus

import (
    "context"
    "sync"

    "github.com/ThreeDotsLabs/watermill"
    "github.com/ThreeDotsLabs/watermill/message"
    gochannel "github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// Topics used by the REPL timeline bus.
const (
    TopicReplEvents = "repl.events"
    TopicUIEntities = "ui.entities"
)

// Bus wraps a Watermill in-memory router + pubsub.
type Bus struct {
    Router     *message.Router
    Publisher  message.Publisher
    Subscriber message.Subscriber

    runOnce sync.Once
}

// NewInMemoryBus creates an in-memory Watermill router with gochannel pubsub.
func NewInMemoryBus() (*Bus, error) {
    logger := watermill.NewStdLogger(false, false)
    pubsub := gochannel.NewGoChannel(gochannel.Config{OutputChannelBuffer: 1024}, logger)
    r, err := message.NewRouter(message.RouterConfig{}, logger)
    if err != nil {
        return nil, err
    }
    return &Bus{Router: r, Publisher: pubsub, Subscriber: pubsub}, nil
}

// AddHandler subscribes to a topic and handles messages (no publishing from handler by default).
func (b *Bus) AddHandler(name, topic string, handler func(*message.Message) error) {
    b.Router.AddNoPublisherHandler(name, topic, b.Subscriber, handler)
}

// Run starts the router until ctx is done.
func (b *Bus) Run(ctx context.Context) error {
    var err error
    b.runOnce.Do(func() {
        go func() {
            // Router blocks; stop when context cancelled
            <-ctx.Done()
            _ = b.Router.Close()
        }()
        err = b.Router.Run(ctx)
    })
    return err
}

