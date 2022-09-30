package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/IfanTsai/go-lib/gin/middlewares"
	"github.com/IfanTsai/go-lib/logger"
	"github.com/IfanTsai/go-lib/user/token"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/pkg/errors"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
	server     *http.Server
	address    string
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store, address string) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create token")
	}
	server := &Server{
		config:     config,
		store:      store,
		address:    address,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("currency", validCurrency); err != nil {
			log.Fatalln("cannot register currency validation, err:", err)
		}
	}

	server.setupRouter()

	return server, nil
}

func (s *Server) setupRouter() {
	version := "1.0.0"

	jsonLogger := logger.NewJSONLogger(
		logger.WithFileRotationP("./logs/simple-bank.log"),
	)

	router := gin.New()
	router.Use(
		middlewares.Logger(jsonLogger),
		middlewares.Recovery(version, jsonLogger, true),
		middlewares.Jsonifier(version),
	)

	if gin.Mode() != gin.TestMode {
		middlewares.NewPrometheus("simple_bank", "api").Use(router)
	}

	v1API := router.Group("/v1")

	v1API.POST("users", s.createUser)
	v1API.POST("users/login", s.loginUser)
	v1API.POST("token/refresh_access", s.refreshAccessToken)

	authRoutes := v1API.Use(middlewares.Authorization(version, s.tokenMaker))
	authRoutes.POST("accounts", s.createAccount)
	authRoutes.GET("accounts/:id", s.getAccount)
	authRoutes.GET("accounts", s.listAccount)
	authRoutes.POST("transfers", s.createTransfer)

	s.router = router
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start() error {
	server := &http.Server{
		Addr:              s.address,
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	s.server = server

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "failed to start Gin server")
	}

	return nil
}

// Stop stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	return errors.Wrap(s.server.Shutdown(ctx), "failed to shutdown http server")
}

func (s *Server) GetTokenMaker() token.Maker {
	return s.tokenMaker
}

func (s *Server) Getrouter() *gin.Engine {
	return s.router
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
