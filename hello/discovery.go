package hello

import (
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/log"
	"os"
	"github.com/hashicorp/consul/api"
	"github.com/go-kit/kit/sd"
	"math/rand"
)

func Register(consulAddress string, consulPort string) (registar sd.Registrar) {
	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	r := rand.New(rand.NewSource(99))

	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulAddress + ":" + consulPort
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	check := api.AgentServiceCheck{
		HTTP: "http://" + consulAddress + ":7777" + "/health",
		Interval: "10s",
		Timeout: "1s",
		Notes: "Basic health checks",
	}

	asr := api.AgentServiceRegistration{
		ID: "hello" + string(r.Int()),
		Name: "hello",
		Address: consulAddress,
		Port: 7777,
		Tags: []string{"hello", "playgound"},
		Check: &check,
	}

	registar = consulsd.NewRegistrar(client, &asr, logger)
	return
}