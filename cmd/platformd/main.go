package main

import (
	"flag"
	"github.com/peterbourgon/ff/v3"
	"github.com/spacechunks/platform/internal/platformd"
	"log/slog"
	"os"
)

func main() {
	var (
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
		fs     = flag.NewFlagSet("platformd", flag.ContinueOnError)

		proxyServiceListenSock = fs.String("proxy-service-listen-sock", "/var/run/platformd/platformd.sock", "path to the unix domain socket to listen on")
		_                      = fs.String("config", "/etc/platformd/config.json", "path to the config file")
	)
	if err := ff.Parse(fs, os.Args[1:],
		ff.WithEnvVarPrefix("PLATFORMD"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.JSONParser),
	); err != nil {
		die("failed to parse config", err, logger)
	}

	var (
		cfg = platformd.Config{
			ProxyServiceListenSock: *proxyServiceListenSock,
		}
		server = platformd.NewServer(logger)
	)

	if err := server.Run(cfg); err != nil {
		die("failed to start server", err, logger)
	}
}

func die(msg string, err error, logger *slog.Logger) {
	logger.Error(msg, "err", err)
	os.Exit(1)
}
