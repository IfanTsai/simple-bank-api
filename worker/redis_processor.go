package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	db "github.com/ifantsai/simple-bank-api/db/sqlc"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	logger *zap.Logger
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, opts ...Option) TaskProcessor {
	opt := &option{}
	for _, f := range opts {
		f(opt)
	}

	return &RedisTaskProcessor{
		server: asynq.NewServer(redisOpt, asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  1,
			},
		}),
		store:  store,
		logger: opt.logger,
	}
}

func (rtp *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, rtp.ProcessTaskSendVerifyEmail)

	log.Println("starting task processor")

	return errors.Wrap(rtp.server.Run(mux), "failed to run task processor server")
}

func (rtp *RedisTaskProcessor) Stop(ctx context.Context) error {
	rtp.server.Shutdown()

	return nil
}

func (rtp *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return errors.Wrapf(err, "failed to unmarshal payload for task %s", task.Type())
	}

	user, err := rtp.store.GetUser(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}

		return errors.Wrap(err, "failed to get user")
	}

	// TODO: send email to user
	_ = user

	if rtp.logger != nil {
		rtp.logger.Info("task processed",
			zap.String("type", task.Type()),
			zap.ByteString("payload", task.Payload()),
			zap.String("email", user.Email),
		)
	}

	return nil
}
