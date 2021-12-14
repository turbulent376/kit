package listener

import (
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"git.jetbrains.space/orbi/fcsd/kit/queue"
	"sync"
)

// QueueMessageHandler is a func which acts as message handler
type QueueMessageHandler func(payload []byte) error

// QueueListener supports multiple subscriptions on Queue topics
type QueueListener interface {
	// Add adds handlers
	Add(qt queue.QueueType, topic string, h ...QueueMessageHandler)
	// AddLb adds handlers with load balancing
	AddLb(qt queue.QueueType, topic, lbGroup string, h ...QueueMessageHandler)
	// ListenAsync starts goroutine which is listening incoming messages and calls proper handlers
	ListenAsync()
	// Stop stops listening
	Stop()
	// Clear clears all handlers
	Clear()
}

// topicKey used as a key for handlers
type topicKey struct {
	Topic   string // Queue topic
	LbGroup string // LB group
}

func NewQueueListener(q queue.Queue, logger log.CLoggerFunc) QueueListener {

	th := map[queue.QueueType]map[topicKey][]QueueMessageHandler{}
	th[queue.QueueTypeAtLeastOnce] = make(map[topicKey][]QueueMessageHandler)
	th[queue.QueueTypeAtMostOnce] = make(map[topicKey][]QueueMessageHandler)

	return &queueListener{
		topicHandlers: th,
		listening:     false,
		queue:         q,
		logger:        logger,
	}
}

type queueListener struct {
	sync.RWMutex
	queue         queue.Queue
	topicHandlers map[queue.QueueType]map[topicKey][]QueueMessageHandler
	quit          chan struct{}
	listening     bool
	logger        log.CLoggerFunc
}

func (q *queueListener) add(qt queue.QueueType, topic, lbGroup string, h ...QueueMessageHandler) {

	q.Stop()

	q.Lock()
	defer q.Unlock()

	key := topicKey{Topic: topic, LbGroup: lbGroup}

	var handlers []QueueMessageHandler
	handlers, ok := q.topicHandlers[qt][key]
	if !ok {
		handlers = []QueueMessageHandler{}
	}

	handlers = append(handlers, h...)
	q.topicHandlers[qt][key] = handlers

}

func (q *queueListener) Add(qt queue.QueueType, topic string, h ...QueueMessageHandler) {
	q.add(qt, topic, "", h...)
}

func (q *queueListener) AddLb(qt queue.QueueType, topic, lbGroup string, h ...QueueMessageHandler) {
	q.add(qt, topic, lbGroup, h...)
}

func (q *queueListener) ListenAsync() {

	// go through all queue types
	for queueType, topicHandlers := range q.topicHandlers {
		// go through handlers of the queue type
		for key, handlers := range topicHandlers {

			// start goroutine for each subscriber
			go func(qt queue.QueueType, tp, lb string, hnds []QueueMessageHandler) {
				c := make(chan []byte)

				// if load balancing group specified, make load balanced subscription
				if lb == "" {
					_ = q.queue.Subscribe(qt, tp, c)
				} else {
					_ = q.queue.SubscribeLB(qt, tp, lb, c)
				}

				// waiting for messages
				for {
					select {
					case msg := <-c:
						for _, h := range hnds {
							m := msg
							h := h
							// execute handlers within a separate goroutines
							go func() {
								l := q.logger().Pr("queue").Cmp("listener").F(log.FF{"topic": tp}).TrcF("%s", string(m))
								if err := h(m); err != nil {
									l.E(err).St().Err()
								}
							}()
						}
					case <-q.quit:
						return
					}
				}
			}(queueType, key.Topic, key.LbGroup, handlers)
		}
	}

}

func (q *queueListener) Stop() {
	q.RLock()
	l := q.listening
	q.RUnlock()

	if l {
		q.quit <- struct{}{}
		q.Lock()
		defer q.Unlock()
		q.listening = false
	}
}

func (q *queueListener) Clear() {
	q.Stop()
	q.quit <- struct{}{}
	q.Lock()
	defer q.Unlock()
	q.topicHandlers[queue.QueueTypeAtLeastOnce] = make(map[topicKey][]QueueMessageHandler)
	q.topicHandlers[queue.QueueTypeAtMostOnce] = make(map[topicKey][]QueueMessageHandler)
}
