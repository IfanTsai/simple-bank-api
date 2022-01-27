package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/IfanTsai/go-lib/gin/middlewares"
	"github.com/gin-gonic/gin"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (s *Server) createAccount(c *gin.Context) {
	var req createAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	// A logged-in user can only create an account for him/herself
	// note: set payload in auth middleware
	username, err := middlewares.GetUsername(c)
	if err != nil {
		return
	}

	account, err := s.store.CreateAccount(c, db.CreateAccountParams{
		Owner:    username,
		Currency: req.Currency,
		Balance:  0,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok { // nolint: errorlint
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				c.JSON(http.StatusForbidden, errorResponse(err))

				return
			}
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))

		return
	}

	c.JSON(http.StatusOK, account)
}

func (s *Server) getAccount(c *gin.Context) {
	var req getAccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	account, err := s.store.GetAccount(c, req.ID)
	if err != nil {
		httpCode := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			httpCode = http.StatusNotFound
		}

		c.JSON(httpCode, errorResponse(err))

		return
	}

	// A logged-in user can only get accounts that he/she owns
	username, err := middlewares.GetUsername(c)
	if err != nil {
		return
	}

	if account.Owner != username {
		err := errors.New("account doesn't belong to the authenticated user")
		c.JSON(http.StatusUnauthorized, errorResponse(err))

		return
	}

	c.JSON(http.StatusOK, account)
}

func (s *Server) listAccount(c *gin.Context) {
	var req listAccountRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	// A logged-in user can only list accounts that belong to him/her
	username, err := middlewares.GetUsername(c)
	if err != nil {
		return
	}

	accounts, err := s.store.ListAccounts(c, db.ListAccountsParams{
		Owner:  username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	})
	if err != nil {
		httpCode := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			httpCode = http.StatusNotFound
		}

		c.JSON(httpCode, errorResponse(err))

		return
	}

	c.JSON(http.StatusOK, accounts)
}
