package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (s *Server) createUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))

		return
	}

	user, err := s.store.CreateUser(c, db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok { // nolint: errorlint
			switch pqErr.Code.Name() {
			case "unique_violation":
				c.JSON(http.StatusForbidden, errorResponse(err))

				return
			}
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))

		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

func (s *Server) loginUser(c *gin.Context) {
	var req loginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	user, err := s.store.GetUser(c, req.Username)
	if err != nil {
		errorHTTPCode := http.StatusInternalServerError
		if errors.Is(errors.Cause(err), sql.ErrNoRows) {
			errorHTTPCode = http.StatusNotFound
		}

		c.JSON(errorHTTPCode, errorResponse(err))

		return
	}

	if err = util.CheckPassword(req.Password, user.HashedPassword); err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(err))

		return
	}

	// Note: only use username
	accessToken, err := s.tokenMaker.CreateToken(0, user.Username, s.config.AccessTokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))

		return
	}

	c.JSON(http.StatusOK, loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	})
}
