package ginmill

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func routes() (r gin.RoutesInfo) {
	e := gin.New()

	e.GET("/get", func(c *gin.Context) {})

	return r
}

func TestNewFeatures(t *testing.T) {
	type args struct {
		routes gin.RoutesInfo
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test NewFeatures",
			args: args{
				routes: routes(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFeatures(tt.args.routes)
			r := reflect.TypeOf(got)
			if r.Kind() != reflect.Ptr || r.Elem().Name() != "Features" {
				t.Errorf("NewFeatures() = %v, want *Features", got)
			}
		})
	}
}


func TestServer_With(t *testing.T) {
	type args struct {
		r gin.RoutesInfo
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test With",
			args: args{
				r: routes(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := gin.New()
			s := &Server{
				Engine: engine,
			}
			f := NewFeatures(tt.args.r)

			if s.With(f); !reflect.DeepEqual(tt.args.r, engine.Routes()) {
				t.Errorf("Server.With() r = %v, routes %v", tt.args.r, engine.Routes())
			}
		})
	}
}
