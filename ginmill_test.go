package ginmill_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ginmills/ginmill"
)

type cheersHandler interface {
	SayCheers(c *gin.Context)
}

type stubCheers struct{}

func (stubCheers) SayCheers(c *gin.Context) {
	c.String(http.StatusOK, "cheers")
}

type customCheers struct {
	reply string
}

func (ch customCheers) SayCheers(c *gin.Context) {
	c.String(http.StatusOK, ch.reply)
}

func newCheersFeatures(h cheersHandler) *ginmill.Features {
	r := gin.New()
	r.GET("/cheers", h.SayCheers)
	return ginmill.NewFeatures(r.Routes())
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func routes() gin.RoutesInfo {
	e := gin.New()
	e.GET("/get", func(_ *gin.Context) {})
	return e.Routes()
}

func assertRoutesMatch(t *testing.T, got, want gin.RoutesInfo) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("routes length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Method != want[i].Method || got[i].Path != want[i].Path {
			t.Errorf("route[%d] = %s %s, want %s %s",
				i, got[i].Method, got[i].Path, want[i].Method, want[i].Path)
		}
	}
}

func performRequest(t *testing.T, engine *gin.Engine, method, target string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, target, http.NoBody)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestNewFeatures(t *testing.T) {
	t.Parallel()

	sample := routes()

	tests := []struct {
		name      string
		routes    gin.RoutesInfo
		wantCount int
		want      gin.RoutesInfo
	}{
		{
			name:      "non-empty routes",
			routes:    sample,
			wantCount: 1,
			want:      sample,
		},
		{
			name:      "empty slice",
			routes:    gin.RoutesInfo{},
			wantCount: 0,
		},
		{
			name:      "nil slice",
			routes:    nil,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := ginmill.NewFeatures(tt.routes)
			if f == nil {
				t.Fatal("NewFeatures() = nil, want non-nil *Features")
			}

			got := f.GetRoutes()
			if len(got) != tt.wantCount {
				t.Fatalf("len(GetRoutes()) = %d, want %d", len(got), tt.wantCount)
			}
			if tt.want != nil {
				assertRoutesMatch(t, got, tt.want)
				if got[0].HandlerFunc == nil {
					t.Error("route[0].HandlerFunc = nil, want non-nil")
				}
			}
		})
	}
}

func TestFeatures_GetRoutes(t *testing.T) {
	t.Parallel()

	t.Run("returns defensive copy without mutating internal routes", func(t *testing.T) {
		t.Parallel()

		want := routes()
		f := ginmill.NewFeatures(want)

		got := f.GetRoutes()
		assertRoutesMatch(t, got, want)

		got[0].Path = "/mutated"
		got[0].Method = http.MethodDelete

		after := f.GetRoutes()
		assertRoutesMatch(t, after, want)
	})

	t.Run("empty features returns empty slice", func(t *testing.T) {
		t.Parallel()

		f := ginmill.NewFeatures(nil)
		got := f.GetRoutes()
		if got == nil || len(got) != 0 {
			t.Fatalf("GetRoutes() = %#v, want empty non-nil gin.RoutesInfo", got)
		}
	})
}

func TestServer_With(t *testing.T) {
	t.Parallel()

	t.Run("mounts single feature and dispatches HTTP request", func(t *testing.T) {
		t.Parallel()

		engine := gin.New()
		s := &ginmill.Server{Engine: engine}
		f := newCheersFeatures(stubCheers{})

		gotServer := s.With(f)
		if gotServer != s {
			t.Errorf("Server.With() returned %p, want %p", gotServer, s)
		}

		rec := performRequest(t, engine, http.MethodGet, "/cheers")
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /cheers status = %d, want %d", rec.Code, http.StatusOK)
		}
		if gotBody := rec.Body.String(); gotBody != "cheers" {
			t.Errorf("GET /cheers body = %q, want %q", gotBody, "cheers")
		}
	})

	t.Run("reuses the same shared route contract across distinct implementations and servers", func(t *testing.T) {
		t.Parallel()

		engineA := gin.New()
		(&ginmill.Server{Engine: engineA}).With(newCheersFeatures(stubCheers{}))

		engineB := gin.New()
		(&ginmill.Server{Engine: engineB}).With(newCheersFeatures(customCheers{reply: "from-service-b"}))

		recA := performRequest(t, engineA, http.MethodGet, "/cheers")
		if recA.Code != http.StatusOK || recA.Body.String() != "cheers" {
			t.Errorf("serverA GET /cheers = (%d, %q), want (%d, %q)",
				recA.Code, recA.Body.String(), http.StatusOK, "cheers")
		}

		recB := performRequest(t, engineB, http.MethodGet, "/cheers")
		if recB.Code != http.StatusOK || recB.Body.String() != "from-service-b" {
			t.Errorf("serverB GET /cheers = (%d, %q), want (%d, %q)",
				recB.Code, recB.Body.String(), http.StatusOK, "from-service-b")
		}
	})

	t.Run("mounts grouped route with path parameter", func(t *testing.T) {
		t.Parallel()

		def := gin.New()
		v1 := def.Group("/api/v1")
		v1.GET("/statuses/:id", func(c *gin.Context) {
			c.String(http.StatusOK, "status:"+c.Param("id"))
		})

		engine := gin.New()
		s := &ginmill.Server{Engine: engine}
		s.With(ginmill.NewFeatures(def.Routes()))

		rec := performRequest(t, engine, http.MethodGet, "/api/v1/statuses/42")
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/v1/statuses/42 status = %d, want %d", rec.Code, http.StatusOK)
		}
		if gotBody := rec.Body.String(); gotBody != "status:42" {
			t.Errorf("GET /api/v1/statuses/42 body = %q, want %q", gotBody, "status:42")
		}
	})

	t.Run("mounts empty feature without affecting existing routes", func(t *testing.T) {
		t.Parallel()

		engine := gin.New()
		engine.GET("/existing", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
		s := &ginmill.Server{Engine: engine}
		s.With(ginmill.NewFeatures(nil))

		rec := performRequest(t, engine, http.MethodGet, "/existing")
		if rec.Code != http.StatusOK || rec.Body.String() != "ok" {
			t.Fatalf("GET /existing = (%d, %q), want (%d, %q)",
				rec.Code, rec.Body.String(), http.StatusOK, "ok")
		}
	})

	t.Run("chains multiple distinct features on the same server", func(t *testing.T) {
		t.Parallel()

		r1 := gin.New()
		r1.GET("/alpha", func(c *gin.Context) {
			c.String(http.StatusOK, "alpha")
		})
		r2 := gin.New()
		r2.POST("/beta", func(c *gin.Context) {
			c.String(http.StatusCreated, "beta")
		})

		engine := gin.New()
		s := &ginmill.Server{Engine: engine}
		s.With(ginmill.NewFeatures(r1.Routes())).With(ginmill.NewFeatures(r2.Routes()))

		alphaRec := performRequest(t, engine, http.MethodGet, "/alpha")
		if alphaRec.Code != http.StatusOK || alphaRec.Body.String() != "alpha" {
			t.Errorf("GET /alpha = (%d, %q), want (%d, %q)",
				alphaRec.Code, alphaRec.Body.String(), http.StatusOK, "alpha")
		}

		betaRec := performRequest(t, engine, http.MethodPost, "/beta")
		if betaRec.Code != http.StatusCreated || betaRec.Body.String() != "beta" {
			t.Errorf("POST /beta = (%d, %q), want (%d, %q)",
				betaRec.Code, betaRec.Body.String(), http.StatusCreated, "beta")
		}
	})

	t.Run("executes only route HandlerFunc without carrying temporary group middleware", func(t *testing.T) {
		t.Parallel()

		mwCalled := false
		def := gin.New()
		grp := def.Group("/mw", func(c *gin.Context) {
			mwCalled = true
			c.Next()
		})
		grp.GET("/ping", func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		})

		engine := gin.New()
		s := &ginmill.Server{Engine: engine}
		s.With(ginmill.NewFeatures(def.Routes()))

		rec := performRequest(t, engine, http.MethodGet, "/mw/ping")
		if rec.Code != http.StatusOK || rec.Body.String() != "pong" {
			t.Fatalf("GET /mw/ping = (%d, %q), want (%d, %q)",
				rec.Code, rec.Body.String(), http.StatusOK, "pong")
		}
		if mwCalled {
			t.Error("temporary engine group middleware was invoked; expected only route HandlerFunc")
		}
	})

	t.Run("duplicate method and path panics following Gin semantics", func(t *testing.T) {
		t.Parallel()

		def := gin.New()
		def.GET("/dup", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		f := ginmill.NewFeatures(def.Routes())

		engine := gin.New()
		s := &ginmill.Server{Engine: engine}
		s.With(f)

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("second Server.With() with duplicate route did not panic")
			}
		}()
		s.With(f)
	})
}
