package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/token"
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

	fromAccount, valid := s.validAccount(c, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	// A logged-in user can only send money from his/her own account
	authPayload, ok := c.MustGet(authorizationPayloadKey).(*token.Payload)
	if !ok {
		return
	}

	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		c.JSON(http.StatusUnauthorized, errorResponse(err))

		return
	}

	_, valid = s.validAccount(c, req.ToAccountID, req.Currency)
	if !valid {
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

func (s *Server) validAccount(c *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := s.store.GetAccount(c, accountID)
	if err != nil {
		httpCode := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			httpCode = http.StatusNotFound
		}

		c.JSON(httpCode, errorResponse(err))

		return account, false
	}

	if account.Currency != currency {
		err := xerrors.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return account, false
	}

	return account, true
}
