package worker

import "go.uber.org/zap"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

type Option func(*option)

type option struct {
	logger *zap.Logger
}

func WithLogger(logger *zap.Logger) Option {
	return func(opt *option) {
		opt.logger = logger
	}
}
