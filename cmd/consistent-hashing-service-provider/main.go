package main

import (
	"context"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	conf "github.com/amazingchow/consistent-hashing-service-provider/internal/config"
	"github.com/amazingchow/consistent-hashing-service-provider/internal/util"
)

var (
	cfgPathFlag = flag.String("conf", "conf/node.json", "node config file")
	nodeIDFlag  = flag.String("id", "", "node id")
	verboseFlag = flag.Bool("verbose", false, "set verbose output")
)

func main() {
	flag.Parse()

	// 设置全局logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *verboseFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// 加载节点配置
	var cfg conf.Node
	util.LoadConfigFileOrPanic(*cfgPathFlag, &cfg)

	// 设置节点ID, 节点ID用于节点间通信
	nID := *nodeIDFlag
	if nID == "" {
		hn, _ := os.Hostname()
		nID = fmt.Sprintf("%s:%d", hn, util.GetPort(cfg.GRPCEndpoint))
		log.Warn().Msgf("since NodeID not provided, set default value, NodeID=%s", nID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	srv := newConsistentHashingServiceServer(ctx, nID, &cfg)
	defer func() {
		srv.close()
	}()

	// 开启grpc服务
	go serveGPRC(ctx, srv, &cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	for range sigCh {
		break
	}

	log.Info().Msg("stop consistent hashing service")
}
