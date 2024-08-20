package xcluster

import (
	log "github.com/beijian01/xgame/framework/logger"
	"github.com/nats-io/nats.go"
)

type (
	natsSubject struct {
		ch           chan *nats.Msg
		subject      string
		subscription *nats.Subscription
	}
)

func newNatsSubject(subject string, size int) *natsSubject {
	return &natsSubject{
		ch:           make(chan *nats.Msg, size),
		subject:      subject,
		subscription: nil,
	}
}

func (p *natsSubject) stop() {
	err := p.subscription.Unsubscribe()
	if err != nil {
		log.Warnf("Unsubscribe error. [subject = %s, err = %v]", p.subject, err)
	}
	close(p.ch)
}
