package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/segmentio/stats/v4"
	"github.com/segmentio/stats/v4/prometheus"
	"net/http"
	"time"
)

const DefaultFlushDuration = 500 * time.Millisecond
const DefaultRouterDebug = false

type Server struct {
	Port        int `json:"port"`
	StatsEngine *stats.Engine
	Router      *gin.Engine
	Handler     stats.Handler
	FlushEvery  time.Duration
	FlushC      chan error
	FlushDone   chan struct{}
}

func StatsHandlerFunc(h http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func StatsHandler(h http.Handler) GinHandler {
	return GinHandler{
		Method: ANY,
		Funcs: []gin.HandlerFunc{
			StatsHandlerFunc(h),
		},
		Path: "/metrics",
	}
}

func NewDefaultPrometheusMetricServer(port int) *Server {
	h := prometheus.DefaultHandler
	e := stats.DefaultEngine
	return NewPrometheusMetricServer(port, h, e)
}

func NewPrometheusMetricServer(port int, h stats.Handler, eng *stats.Engine) *Server {
	eng.Handler = h
	stats.DefaultEngine = eng
	s := &Server{
		Port:        port,
		StatsEngine: eng,
		Router: NewGinEngine(nil, []GinHandler{
			StatsHandler(h.(http.Handler)),
		}),
		Handler:    h,
		FlushEvery: DefaultFlushDuration,
		FlushC:     nil,
		FlushDone:  make(chan struct{}),
	}
	s.setup()
	return s
}

func (s *Server) setup() {
	s.FlushC = s.StartFlush()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func (s *Server) Run() error {
	return StartGinServer(s.Router, s.Port)
}

func (s *Server) Close() (err error) {
	if s.FlushDone != nil {
		close(s.FlushDone)

		if s.FlushC != nil {
			err = <-s.FlushC
		}
	}

	return nil
}

func (s *Server) StartFlush() chan error {
	errC := make(chan error)
	go func() {
		defer close(errC)
		for range time.Tick(s.FlushEvery) {
			s.StatsEngine.Flush()

			select {
			case <-s.FlushDone:
				return
			default:
			}
		}
	}()
	return errC
}
