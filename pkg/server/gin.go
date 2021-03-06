package metrics

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type GinHandler struct {
	Method HttpMethod        `json:"http_method"`
	Group  *gin.RouterGroup  `json:"router_group"`
	Funcs  []gin.HandlerFunc `json:"handler_funcs"`
	Path   string            `json:"path"`
}

func NewGinEngine(middleware []GinHandler, handlers []GinHandler) *gin.Engine {
	engine := gin.Default()
	for _, mw := range middleware {
		if mw.Group != nil {
			mw.Group.Use(mw.Funcs...)
		} else {
			engine.Use(mw.Funcs...)
		}
	}

	for _, h := range handlers {
		if h.Group != nil {
			setGroupHandler(h.Group, h.Path, h.Method, h.Funcs...)
		} else {
			setEngineHandler(engine, h.Path, h.Method, h.Funcs...)
		}
	}

	return engine
}

func StartGinServer(e *gin.Engine, port int) error {
	if err := e.Run(fmt.Sprintf(":%d", port)); err != nil {
		return fmt.Errorf("router failed to start: %v", err)
	}
	return nil
}

func setGroupHandler(g *gin.RouterGroup, path string, m HttpMethod, h ...gin.HandlerFunc) {
	switch m {
	case OPTIONS, GET, PUT, POST, PATCH, DELETE, CONNECT, TRACE:
		g.Handle(m.String(), path, h...)
	case ANY:
		fallthrough
	default:
		g.Any(path, h...)
	}
}

func setEngineHandler(g *gin.Engine, path string, m HttpMethod, h ...gin.HandlerFunc) {
	switch m {
	case OPTIONS, GET, PUT, POST, PATCH, DELETE, CONNECT, TRACE:
		g.Handle(m.String(), path, h...)
	case ANY:
		fallthrough
	default:
		g.Any(path, h...)
	}
}
