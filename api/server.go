package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/token"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/pkg/errors"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create token")
	}
	server := &Server{
		config:     config,
		store:      store,
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
	router := gin.Default()

	router.POST("/users", s.createUser)
	router.POST("/users/login", s.loginUser)

	router.POST("/accounts", s.createAccount)
	router.GET("/accounts/:id", s.getAccount)
	router.GET("/accounts", s.listAccount)
	router.POST("/transfers", s.createTransfer)

	s.router = router
}

// Start runs the HTTP server on a specific address.
func (s *Server) Start(address string) error {
	return errors.Wrap(s.router.Run(address), "failed to run server")
}

func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
