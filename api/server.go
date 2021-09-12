package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/pkg/errors"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("currency", validCurrency); err != nil {
			log.Fatalln("cannot register currency validation, err:", err)
		}
	}

	router.POST("/users", server.createUser)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccount)
	router.POST("/transfers", server.createTransfer)

	server.router = router

	return server
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
