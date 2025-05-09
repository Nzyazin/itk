package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"crypto/tls"

	"github.com/gorilla/mux"
	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/Nzyazin/itk/internal/core/handler"
	"github.com/Nzyazin/itk/internal/core/repository/postgres"
	"github.com/Nzyazin/itk/internal/core/usecase"
	"github.com/Nzyazin/itk/pkg/config"
	"github.com/Nzyazin/itk/pkg/postgresdb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slok/go-http-metrics/middleware/std"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/metrics/prometheus"
	middlWre "github.com/Nzyazin/itk/internal/core/middleware"
)

type Server struct {
	router *mux.Router
	log    logger.Logger
	httpServer *http.Server
	walletHandler *handler.WalletHandler
	db *postgresdb.Database
}

func NewServer(log logger.Logger) (*Server, error) {

	cfgDB, err := config.LoadConfigDB()
	if err != nil {
		return nil, err
	}

	db, err := postgresdb.NewPostgresDB(*cfgDB, log)
	if err != nil {
		return nil, err
	}

	walletRepository := postgres.NewPostgresWalletRepo(db.DB, log)
	walletUsecase := usecase.NewWalletUsecase(walletRepository, log)
	walletHandler := handler.NewWalletHandler(walletUsecase, log)
	server := &Server{
		log:    log,
		router: mux.NewRouter(),
		walletHandler: walletHandler,
		db: db,
	}

	server.router.Use(loggingMiddleware(server.log))

	mw := middleware.New(middleware.Config{
		Recorder: prometheus.NewRecorder(prometheus.Config{}),
	})
	
	server.router.Use(func(next http.Handler) http.Handler {
		return std.Handler("", mw, next)
	})

	server.RegisterRoutes()

	return server, nil
}

func (s *Server) RegisterRoutes() {
	s.router.Use(
		middlWre.WithErrorHandler(s.log),
		middlWre.Recovery(s.log),
	)
	s.router.HandleFunc("/api/v1/wallet", s.walletHandler.ProcessWalletOperation).Methods("POST")
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	s.router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
}

func (s *Server) Run(addr string) error {
	srv := &http.Server{
		Addr: addr,
		Handler: s.router,
		ReadTimeout:       9 * time.Second,
		WriteTimeout:      12 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
	}

	s.httpServer = srv

	return srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	var shutdownErr error

	go func() {
		if s.httpServer != nil {
			err := s.httpServer.Shutdown(ctx)
			if err != nil {
				s.log.Error("failed to shutdown HTTP server", logger.ErrorField("error", err))
				shutdownErr = fmt.Errorf("HTTP server shutdown error: %w", err)
			}
		}

		if s.db != nil {
			err := s.db.Close()
			if err != nil {
				s.log.Error("failed to close database connection", logger.ErrorField("error", err))
				shutdownErr = fmt.Errorf("database shutdown error: %w", err)
			}
		}

		close(done)
	}()

	select {
	case <-done:
		return shutdownErr
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}
}

func (s *Server) RunTLS(addr, certFile, keyFile string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadTimeout:       9 * time.Second,
		WriteTimeout:      9 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 6 * time.Second,
		TLSConfig:         &tls.Config{MinVersion: tls.VersionTLS12},
	}

	s.httpServer = srv
	return srv.ListenAndServeTLS(certFile, keyFile)
}

func loggingMiddleware(log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("HTTP request",
                logger.StringField("method", r.Method),
                logger.StringField("path", r.URL.Path),
                logger.StringField("remote_addr", r.RemoteAddr),
                logger.StringField("user_agent", r.UserAgent()),
            )
			next.ServeHTTP(w, r)
		})
	}
}
