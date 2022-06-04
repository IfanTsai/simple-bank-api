package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type refreshAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) refreshAccessToken(c *gin.Context) {
	var req refreshAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	refreshTokenPayload, err := s.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(err))

		return
	}

	session, err := s.store.GetSession(c, refreshTokenPayload.ID)
	if err != nil {
		errorHTTPCode := http.StatusInternalServerError
		if errors.Is(errors.Cause(err), sql.ErrNoRows) {
			errorHTTPCode = http.StatusNotFound
		}

		c.JSON(errorHTTPCode, errorResponse(err))

		return
	}

	if session.IsBlocked {
		c.JSON(http.StatusUnauthorized, errorResponse(errors.New("blocked session")))

		return
	}

	if session.Username != refreshTokenPayload.Username {
		c.JSON(http.StatusUnauthorized, errorResponse(errors.New("incorrect session user")))

		return
	}

	if session.RefreshToken != req.RefreshToken {
		c.JSON(http.StatusUnauthorized, errorResponse(errors.New("incorrect session token")))

		return
	}

	if time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, errorResponse(errors.New("expired session")))

		return
	}

	// Note: only use username
	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(
		0, refreshTokenPayload.Username, s.config.AccessTokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))

		return
	}

	c.JSON(http.StatusOK, refreshAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessTokenPayload.ExpiredAt,
	})
}
