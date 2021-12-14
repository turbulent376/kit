package http

import (
	"fmt"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"net/http"
	"time"
)

const (
	ReadTimeout     = time.Second * 10
	WriteTimeout    = time.Second * 10
	ReadBufferSize  = 1024
	WriteBufferSize = 1024
)

type Config struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	Debug          bool
	Port           string
}

// Server represents HTTP server
type Server struct {
	Srv          *http.Server        // Srv - internal server
	RootRouter   *mux.Router         // RootRouter - root router
	AuthRouter   *mux.Router         // AuthRouter - router requiring authentication
	NoAuthRouter *mux.Router         // NoAuthRouter - router not requiring authentication
	WsUpgrader   *websocket.Upgrader // WsUpgrader - websocket upgrader
	logger       log.CLoggerFunc     // logger
}

type RouteSetter interface {
	Set(authRouter, noAuthRouter *mux.Router)
}

type WsUpgrades interface {
	Set(noAuthRouter *mux.Router, upgrader *websocket.Upgrader)
}

type CustomRouteSetter interface {
	SetCustom(root *mux.Router) *mux.Router
}

type WsUpgrader interface {
	Set(noAuthRouter *mux.Router, upgrader *websocket.Upgrader)
}

func NewHttpServer(corsOptions *Config, logger log.CLoggerFunc) *Server {

	r := mux.NewRouter()
	noAuthRouter := r.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return true
	}).Subrouter()
	authRouter := r.MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return r.Header.Get("Authorization") != ""
	}).Subrouter()

	s := &Server{
		Srv: &http.Server{
			Addr: fmt.Sprintf(":%s", corsOptions.Port),
			Handler: cors.New(cors.Options{
				AllowedOrigins:   corsOptions.AllowedOrigins,
				AllowedMethods:   corsOptions.AllowedMethods,
				AllowedHeaders:   corsOptions.AllowedHeaders,
				AllowCredentials: true,
				Debug:            corsOptions.Debug,
			}).Handler(r),
			WriteTimeout: WriteTimeout,
			ReadTimeout:  ReadTimeout,
		},
		WsUpgrader: &websocket.Upgrader{
			ReadBufferSize:  ReadBufferSize,
			WriteBufferSize: WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		RootRouter:   r,
		AuthRouter:   authRouter,
		NoAuthRouter: noAuthRouter,
		logger:       logger,
	}

	return s
}

func (s *Server) SetRouters(routeSetters ...RouteSetter) {
	for _, rs := range routeSetters {
		rs.Set(s.AuthRouter, s.NoAuthRouter)
	}
}

func (s *Server) SetWsUpgrader(upgradeSetter WsUpgrader) {
	upgradeSetter.Set(s.NoAuthRouter, s.WsUpgrader)
}

func (s *Server) SetAuthMiddleware(mdws ...mux.MiddlewareFunc) {
	for _, m := range mdws {
		s.AuthRouter.Use(m)
	}
}

func (s *Server) SetNoAuthMiddleware(mdws ...mux.MiddlewareFunc) {
	for _, m := range mdws {
		s.NoAuthRouter.Use(m)
	}
}

func (s *Server) SetMiddleware(mdws ...mux.MiddlewareFunc) {
	for _, m := range mdws {
		s.NoAuthRouter.Use(m)
		s.AuthRouter.Use(m)
	}
}

func (s *Server) Listen() {
	go func() {
		l := s.logger().Pr("http").Cmp("server").Mth("listen").F(log.FF{"url": s.Srv.Addr})
		l.Inf("start listening")

		// if tls parameters are specified, list tls connection
		if err := s.Srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				l.E(ErrHttpSrvListen(err)).St().Err()
			} else {
				l.Dbg("server closed")
			}
			return
		}
	}()
}

func (s *Server) Close() {
	_ = s.Srv.Close()
}
