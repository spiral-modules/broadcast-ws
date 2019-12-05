package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/spiral/broadcast"
)

type ConnContext struct {
	// Conn to the client.
	Conn *websocket.Conn

	// Topics contain list of currently subscribed topics.
	Topics []string

	// upstream to push messages into.
	upstream chan *broadcast.Message
}

func (ctx *ConnContext) SendMessage(topic string, payload interface{}) (err error) {
	msg := &broadcast.Message{Topic: topic}
	msg.Payload, err = json.Marshal(payload)

	if err == nil {
		ctx.upstream <- msg
	}

	return err
}

func (ctx *ConnContext) addTopics(topics ...string) {
	for _, topic := range topics {
		found := false
		for _, e := range ctx.Topics {
			if e == topic {
				found = true
				break
			}
		}

		if !found {
			ctx.Topics = append(ctx.Topics, topic)
		}
	}
}

func (ctx *ConnContext) dropTopic(topics ...string) {
	for _, topic := range topics {
		for i, e := range ctx.Topics {
			if e == topic {
				ctx.Topics[i] = ctx.Topics[len(ctx.Topics)-1]
				ctx.Topics = ctx.Topics[:len(ctx.Topics)-1]
			}
		}
	}
}
