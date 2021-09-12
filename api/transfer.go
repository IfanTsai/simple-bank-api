package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	xerrors "github.com/pkg/errors"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (s *Server) createTransfer(c *gin.Context) {
	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	if !s.validAccount(c, req.FromAccountID, req.Currency) || !s.validAccount(c, req.ToAccountID, req.Currency) {
		return
	}

	result, err := s.store.TransferTx(c, db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) validAccount(c *gin.Context, accountID int64, currency string) bool {
	account, err := s.store.GetAccount(c, accountID)
	if err != nil {
		httpCode := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			httpCode = http.StatusNotFound
		}

		c.JSON(httpCode, errorResponse(err))

		return false
	}

	if account.Currency != currency {
		err := xerrors.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return false
	}

	return true
}
