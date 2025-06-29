package server

import "go.uber.org/zap"

type BaseMiddleware struct {
	logger *zap.Logger
}
