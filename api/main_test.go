package api_test

import (
	"os"
	"testing"
	"time"

	"github.com/ifantsai/simple-bank-api/api"

	"github.com/IfanTsai/go-lib/randutils"
	"github.com/gin-gonic/gin"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/ifantsai/simple-bank-api/util"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func NewTestServer(t *testing.T, store db.Store) *api.Server {
	config := util.Config{
		TokenSymmetricKey:   randutils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := api.NewServer(config, store, "")
	require.NoError(t, err)

	return server
}
