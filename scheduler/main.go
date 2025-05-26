//go:build !test

package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/oiler-backup/postgres-adapter/scheduler/internal/config"
	"github.com/oiler-backup/postgres-adapter/scheduler/internal/server"

	loggerbase "github.com/oiler-backup/base/logger"
)

func main() {
	logger, err := loggerbase.GetLogger(loggerbase.PRODUCTION)
	if err != nil {
		panic(err)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Panicw("Failed to get config", "error", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprint(":", cfg.Port))
	if err != nil {
		logger.Panicw("Failed to listen port", "error", err)
	}

	grpcServer := grpc.NewServer()

	err = server.RegisterBackupServer(grpcServer, cfg.SystemNamespace, cfg.BackuperVersion, cfg.RestorerVersion)
	if err != nil {
		logger.Panicw("Failed to register backup server", "error", err)
	}
	logger.Infof("Running grpc server on port %d...", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatalw("Failed running server", "error", err)
	}
}
