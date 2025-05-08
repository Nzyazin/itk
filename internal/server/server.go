package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/Nzyazin/itk/internal/core/logger"
	"github.com/Nzyazin/itk/internal/core/handler"
)

type ServerConfig struct {
	Port int
}

type Server struct {
	router *mux.Router
	cfg    *ServerConfig
	log    logger.Logger
	httpServer *http.Server
	walletHandler *handler.WalletHandler
}

func NewServer(cfg *ServerConfig, log logger.Logger) (*Server, error) {

	walletRepository := repository.NewWalletRepository()
	walletHandler := handler.NewWalletHandler(walletUsecase)
	walletUsecase := usecase.NewWalletUsecase()
	server := &Server{
		cfg:    cfg,
		log:    log,
		router: mux.NewRouter(),
		walletHandler: 
	}

	server.RegisterRoutes()

	return server, nil
}

func (s *Server) RegisterRoutes() {
	s.router.HandleFunc("/api/v1/wallet", s.ProcessWalletOperation).Methods("POST")
}


