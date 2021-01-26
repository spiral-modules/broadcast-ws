package ws

import (
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ws "github.com/spiral/broadcast-ws"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/cmd/util"
)

func init() {
	cobra.OnInitialize(func() {
		if rr.Debug {
			svc, _ := rr.Container.Get(ws.ID)
			if svc, ok := svc.(*ws.Service); ok {
				svc.AddListener((&debugger{logger: rr.Logger}).listener)
			}
		}
	})
}

// listener provide debug callback for system events. With colors!
type debugger struct{ logger *logrus.Logger }

// listener listens to http events and generates nice looking output.
func (s *debugger) listener(event int, ctx interface{}) {
	switch event {
	case ws.EventConnect:
		conn := ctx.(*websocket.Conn)
		s.logger.Debug(util.Sprintf(
			"[ws] <green+hb>%s</reset> connected",
			conn.RemoteAddr(),
		))

	case ws.EventDisconnect:
		conn := ctx.(*websocket.Conn)
		s.logger.Debug(util.Sprintf(
			"[ws] <yellow+hb>%s</reset> disconnected",
			conn.RemoteAddr(),
		))

	case ws.EventJoin:
		e := ctx.(*ws.TopicEvent)
		s.logger.Debug(util.Sprintf(
			"[ws] <white+hb>%s</reset> join <cyan+hb>[%s]</reset>",
			e.Conn.RemoteAddr(),
			strings.Join(e.Topics, ", "),
		))

	case ws.EventLeave:
		e := ctx.(*ws.TopicEvent)
		s.logger.Debug(util.Sprintf(
			"[ws] <white+hb>%s</reset> leave <magenta+hb>[%s]</reset>",
			e.Conn.RemoteAddr(),
			strings.Join(e.Topics, ", "),
		))

	case ws.EventError:
		e := ctx.(*ws.ErrorEvent)
		if e.Conn != nil {
			s.logger.Debug(util.Sprintf(
				"[ws] <grey+hb>%s</reset> <yellow>%s</reset>",
				e.Conn.RemoteAddr(),
				e.Error.Error(),
			))
		} else {
			s.logger.Error(util.Sprintf(
				"[ws]: <red+hb>%s</reset>",
				e.Error.Error(),
			))
		}
	}
}
