package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rafaeljesus/retry-go"
	"github.com/sirupsen/logrus"

	"github.com/dink10/proxy-client"
)

const (
	logLevel = "debug"
)

func main() {
	logger := logrus.New()
	err := InitLogger(logger, logLevel)
	if err != nil {
		logrus.Fatal(err)
	}

	client := proxy_client.NewClient(proxy_client.Config{}, logger)
	var res []byte
	if err := retry.Do(func() error {
		var err error
		res, err = client.DoRequest(
			"https://google.com",
			"GET",
			proxy_client.Options{},
		)
		return err
	}, 100, 1*time.Millisecond); err != nil {
		logger.Fatal(err)
	}

	logger.Println(string(res))
	client.Stop()
}

// InitLogger inits logger
func InitLogger(logger *logrus.Logger, logLevel string) error {
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("failed to parse log level. %v", err)
	}
	logger.SetLevel(lvl)

	return nil
}
