package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(context.Context, *PayloadSendVerifyEmail, ...asynq.Option) error
}

type TaskProcessor interface {
	Start() error
	Stop(ctx context.Context) error
	ProcessTaskSendVerifyEmail(context.Context, *asynq.Task) error
}
