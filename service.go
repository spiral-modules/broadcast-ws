package ws

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/spiral/broadcast"
	"github.com/spiral/roadrunner/service/env"
	rhttp "github.com/spiral/roadrunner/service/http"
	"github.com/spiral/roadrunner/service/http/attributes"
	"github.com/spiral/roadrunner/service/rpc"
	"net/http"
	"strings"
)

// ID defines service id.
const ID = "ws"

// Service to manage websocket clients.
type Service struct {
	cfg       *Config
	upgrade   websocket.Upgrader
	client    *broadcast.Client
	connPool  *connPool
	listeners []func(event int, ctx interface{})
	stop      chan error
}

// AddListener attaches server event controller.
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.listeners = append(s.listeners, l)
}

// Init the service.
func (s *Service) Init(
	cfg *Config,
	env env.Environment,
	http *rhttp.Service,
	rpc *rpc.Service,
	broadcast *broadcast.Service,
) (bool, error) {
	if broadcast == nil || rpc == nil {
		// unable to activate
		return false, nil
	}

	s.cfg = cfg
	s.client = broadcast.NewClient()
	s.connPool = newPool(s.client, s.reportError)

	if err := rpc.Register(ID, &rpcService{svc: s}); err != nil {
		return false, err
	}

	if env != nil {
		// ensure that underlying kernel knows what route to handle
		env.SetEnv("RR_BROADCAST_URL", cfg.Path)
	}

	// init all this stuff
	s.upgrade = websocket.Upgrader{}
	http.AddMiddleware(s.middleware)

	return true, nil
}

// Serve the websocket connections.
func (s *Service) Serve() error {
	defer s.client.Close()
	defer s.connPool.close()

	s.stop = make(chan error)
	return <-s.stop
}

// Stop the service and disconnect all connections.
func (s *Service) Stop() {
	close(s.stop)
}

// middleware intercepts websocket connections.
func (s *Service) middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != s.cfg.Path {
			f(w, r)
			return
		}

		// checking server access
		if err := s.assertServerAccess(f, r); err != nil {
			// show the error to the user
			err.copy(w)
			return
		}

		conn, err := s.upgrade.Upgrade(w, r, nil)
		if err != nil {
			s.reportError(err, nil)
			return
		}

		s.throw(EventConnect, conn)

		// manage connection
		ctx, err := s.connPool.connect(conn)
		if err != nil {
			s.reportError(err, conn)
			return
		}

		s.serveConn(ctx, f, r)
	}
}

// send and receive messages over websocket
func (s *Service) serveConn(ctx *ConnContext, f http.HandlerFunc, r *http.Request) {
	defer func() {
		if err := s.connPool.disconnect(ctx.Conn); err != nil {
			s.reportError(err, ctx.Conn)
		}
		s.throw(EventDisconnect, ctx.Conn)
	}()

	cmd := &Command{}
	for {
		if err := ctx.Conn.ReadJSON(cmd); err != nil {
			s.reportError(err, ctx.Conn)
			return
		}

		switch cmd.Cmd {
		case "join":
			topics := make([]string, 0)
			if err := cmd.Unmarshal(&topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			if len(topics) == 0 {
				continue
			}

			if err := s.assertAccess(f, r, topics...); err != nil {
				s.reportError(err, ctx.Conn)

				if err := ctx.SendMessage("@deny", topics); err != nil {
					s.reportError(err, ctx.Conn)
					return
				}

				return
			}

			if err := s.connPool.subscribe(ctx, topics...); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}
			if err := ctx.SendMessage("@join", topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			s.throw(EventJoin, &TopicEvent{Conn: ctx.Conn, Topics: topics})
		case "leave":
			topics := make([]string, 0)
			if err := cmd.Unmarshal(&topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			if len(topics) == 0 {
				continue
			}

			if err := s.connPool.unsubscribe(ctx, topics...); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			if err := ctx.SendMessage("@leave", topics); err != nil {
				s.reportError(err, ctx.Conn)
				return
			}

			s.throw(EventLeave, &TopicEvent{Conn: ctx.Conn, Topics: topics})
		}
	}
}

// handle connection error
func (s *Service) reportError(err error, conn *websocket.Conn) {
	s.throw(EventError, &ErrorEvent{Conn: conn, Error: err})
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.listeners {
		l(event, ctx)
	}
}

// todo: move it into validator (!)

// assertServerAccess checks if user can join server and returns error and body if user can not. Must return nil in
// case of error
func (s *Service) assertServerAccess(f http.HandlerFunc, r *http.Request) *accessValidator {
	if err := attributes.Set(r, "ws:joinServer", true); err != nil {
		//	return err
		// todo: need to update it
	}

	defer delete(attributes.All(r), "ws:joinServer")

	w := newValidator()
	f(w, r)

	if !w.IsOK() {
		return w
	}

	return nil
}

// assertAccess checks if user can access given upstream, the application will receive all user headers and cookies.
// the decision to authorize user will be based on response code (200).
func (s *Service) assertAccess(f http.HandlerFunc, r *http.Request, channels ...string) error {
	if err := attributes.Set(r, "ws:joinTopics", strings.Join(channels, ",")); err != nil {
		return err
	}

	defer delete(attributes.All(r), "ws:joinTopics")

	w := newValidator()
	f(w, r)

	if !w.IsOK() {
		return errors.New(string(w.Body()))
	}

	return nil
}
