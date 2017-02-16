package main

import (
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/endpoint"
	"io"
	"strings"
	"net/url"
	"net/http"
	"golang.org/x/net/context"
	ht "github.com/go-kit/kit/transport/http"
	consulsd "github.com/go-kit/kit/sd/consul"
	"os"
	"github.com/go-kit/kit/log"
	"github.com/hashicorp/consul/api"
	"github.com/go-kit/kit/sd/lb"
	"time"
	"github.com/gorilla/mux"
	"os/signal"
	"syscall"
	"fmt"
	"flag"
	"github.com/ru-rocker/gokit-consul/hello"
)

//to execute: go run src/github.com/ru-rocker/gokit-consul/hello/discover.d/main.go -consul.addr 172.20.20.30 -consul.port 8500 -http 9000
func main() {

	var (
		httpAddr = flag.String("http", ":9000", "http listen address")
		consulAddr = flag.String("consul.addr", "", "consul address")
		consulPort = flag.String("consul.port", "", "consul port")
	)
	flag.Parse()

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}


	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()

		consulConfig.Address = "http://" + *consulAddr + ":" + *consulPort
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		client = consulsd.NewClient(consulClient)
	}

	tags := []string{"hello", "playgound"}
	passingOnly := true
	duration := 500 * time.Millisecond
	var helloEndpoint endpoint.Endpoint

	ctx := context.Background()
	r := mux.NewRouter()

	factory := helloFactory(ctx, "GET", "/hello")
	subscriber := consulsd.NewSubscriber(client, factory, logger, "hello", tags, passingOnly)
	balancer := lb.NewRoundRobin(subscriber)
	retry := lb.Retry(3, duration, balancer)
	helloEndpoint = retry

	r.Handle("/hello/rocket", ht.NewServer(ctx, helloEndpoint, hello.DecodeHelloRequest, hello.EncodeResponse))

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func helloFactory(_ context.Context, method, path string) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		if !strings.HasPrefix(instance, "http") {
			instance = "http://" + instance
		}

		tgt, err := url.Parse(instance)
		if err != nil {
			return nil, nil, err
		}
		tgt.Path = path

		var (
			enc ht.EncodeRequestFunc
			dec ht.DecodeResponseFunc
		)
		enc, dec = hello.EncodeJSONRequest, hello.DecodeHelloResponse

		return ht.NewClient(method, tgt, enc, dec).Endpoint(), nil, nil
	}
}