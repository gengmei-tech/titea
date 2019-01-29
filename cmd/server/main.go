package main

import (
	"flag"
	"github.com/gengmei-tech/titea/pkg/util/logutil"
	"github.com/gengmei-tech/titea/server"
	"github.com/gengmei-tech/titea/server/config"
	"github.com/pingcap/pd/pkg/metricutil"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	host       = flag.String("host", "0.0.0.0", "kv server host")
	port       = flag.Uint("port", 0, "kv server port")
	pdAddress  = flag.String("pd-address", "", "pd address")
	loglevel   = flag.String("loglevel", "info", "loglevel")
	logpath    = flag.String("log-path", "", "log path")
	configPath = flag.String("config", "", "config file path")
	runExpGc   = flag.Bool("run-gc", false, "run expire and gc, only one instance can run")
)

var (
	cfg *config.Config
	s   *server.Server
)

func main() {
	flag.Parse()
	loadConfig()
	setupLog()
	log.Info("server started")
	createServer()
	setupMetrics()
	setupSignalHandler()
	setupProbe()
	s.Run()
	os.Exit(0)
}

func loadConfig() {
	cfg = config.NewConfig()
	if *configPath != "" {
		cfg.LoadConfig(*configPath)
	}

	if *host != "" {
		cfg.Host = *host
	}

	if *port != 0 {
		cfg.Port = *port
	}

	if cfg.Port == 0 {
		cfg.Port = 5379
	}

	if *pdAddress != "" {
		cfg.Backend.PdAddr = *pdAddress
	}

	if *loglevel != "" {
		cfg.Log.Level = *loglevel
	}

	if *logpath != "" {
		cfg.Log.LogPath = *logpath
	}
	// 是否运行异步过期及gc 多台实例时仅有一台可以运行expire and gc
	cfg.RunExpGc = *runExpGc
}

func setupLog() {
	logutil.InitLogger(cfg.ToLogConfig())
}

func setupMetrics() {
	if cfg.Metric.PushAddress != "" && cfg.Metric.PushInterval.Duration != time.Duration(0) {
		metricutil.Push(&cfg.Metric)
	}
}

func createServer() {
	s = server.NewServer(cfg)
}

func setupSignalHandler() {
	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {
		sig := <-quitCh
		log.Infof("Got signal [%s] to exit.", sig)
		s.Close()
	}()
}

func setupProbe() {
	go func() {
		mux := newProbeMux()
		if err := http.ListenAndServe(":5000", mux); err != nil {
			log.Infof("Probe servier died:%s", err)
		}
	}()
}
