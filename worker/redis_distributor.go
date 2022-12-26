package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type RedisTaskDistributor struct {
	client *asynq.Client
	logger *zap.Logger
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt, opts ...Option) TaskDistributor {
	opt := &option{}
	for _, f := range opts {
		f(opt)
	}

	return &RedisTaskDistributor{
		client: asynq.NewClient(redisOpt),
		logger: opt.logger,
	}
}

func (rtd *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "failed to marshal payload")
	}

	task := asynq.NewTask(TaskSendVerifyEmail, payloadBytes, opts...)
	taskInfo, err := rtd.client.EnqueueContext(ctx, task)
	if err != nil {
		return errors.Wrap(err, "failed to enqueue task")
	}

	if rtd.logger != nil {
		rtd.logger.Info("task enqueued",
			zap.String("type", task.Type()),
			zap.ByteString("payload", task.Payload()),
			zap.String("queue", taskInfo.Queue),
			zap.Int("max_retry", taskInfo.MaxRetry),
		)
	}

	return nil
}
