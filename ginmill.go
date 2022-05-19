package ginmill

import (
	"github.com/gin-gonic/gin"
)

// Ginmill is an interface intend to describe features of your gin engine
type Ginmill interface {
	With(f *Features) (err error)
}

// Server contains A gin engine and a ginmill
type Server struct {
	Ginmill
	Engine *gin.Engine
}

// Features are pre-defined routes
type Features struct {
	routes gin.RoutesInfo
}

// NewFeatures create Features from RoutesInfo
func NewFeatures(routes gin.RoutesInfo) *Features {
	f := &Features{
		routes: routes,
	}

	return f
}

// With pre-defined Features for ginmill
func (s *Server) With(f *Features) *Server {
	for _, r := range f.routes {
		s.Engine.Handle(r.Method, r.Path, r.HandlerFunc)
	}

	return s
}
