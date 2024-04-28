package pub

import (
	"log"
	"sync"
	"time"
)

type Msg[T any] struct {
	MsgID      int
	MsgContext T
}

func NewPubMsg[T any](id int, ctx T) Msg[T] {
	return Msg[T]{
		MsgID:      id,
		MsgContext: ctx,
	}
}

type (
	Subscriber[T any] chan Msg[T]
	topicFunc[T any]  func(v Msg[T]) bool
)

type ControllerPublisher[T any] struct {
	subscribers map[Subscriber[T]]topicFunc[T]
	m           sync.RWMutex
	bufferSize  int
	timeout     time.Duration
}

func NewPublisher[T any](timeout time.Duration, size int) *ControllerPublisher[T] {
	return &ControllerPublisher[T]{
		subscribers: make(map[Subscriber[T]]topicFunc[T]),
		bufferSize:  size,
		timeout:     timeout,
	}
}

func (p *ControllerPublisher[T]) SubscribeTopic(topic topicFunc[T]) Subscriber[T] {
	p.m.Lock()
	defer p.m.Unlock()
	ch := make(Subscriber[T], p.bufferSize)
	p.subscribers[ch] = topic
	return ch
}

func (p *ControllerPublisher[T]) Subscribe() Subscriber[T] {
	return p.SubscribeTopic(nil)
}

func (p *ControllerPublisher[T]) Evict(sub Subscriber[T]) {
	p.m.Lock()
	defer func() {
		err := recover()
		if err != nil {
			log.Println("already closed", err)
		}
		p.m.Unlock()
	}()
	delete(p.subscribers, sub)
	close(sub)
}

func (p *ControllerPublisher[T]) sendTopic(sub Subscriber[T], topic topicFunc[T], v Msg[T], wg *sync.WaitGroup) {
	defer func() {
		err := recover()
		if err != nil {
			log.Println("send topic error", err)
		}
		wg.Done()
	}()
	if nil != topic && !topic(v) {
		return
	}

	select {
	case sub <- v:
	case <-time.After(p.timeout):
	}
}

func (p *ControllerPublisher[T]) Publish(v Msg[T]) {
	p.m.RLock()
	defer p.m.RUnlock()

	var wg sync.WaitGroup
	for sub, topic := range p.subscribers {
		wg.Add(1)
		go p.sendTopic(sub, topic, v, &wg)
	}
}

func (p *ControllerPublisher[T]) Close() {
	p.m.Lock()
	defer p.m.Unlock()

	for sub := range p.subscribers {
		delete(p.subscribers, sub)
		close(sub)
	}
}
