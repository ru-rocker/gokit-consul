package main

import (
	ht "github.com/go-kit/kit/transport/http"
	"net/http"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"log"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/ru-rocker/gokit-consul/hello"
)

func main() {
	ctx := context.Background()

	var (
		httpAddr = flag.String("http", ":7777",
			"http listen address")
	)
	flag.Parse()

	errChan := make(chan error)

	var svc hello.Service
	svc = hello.HelloService{}

	r := mux.NewRouter()

	r.Handle("/hello", ht.NewServer(
		ctx,
		hello.MakeHelloEndpoint(svc),
		hello.DecodeHelloRequest,
		hello.EncodeResponse,
	))

	r.Methods("GET").Path("/health").Handler(ht.NewServer(
		ctx,
		hello.MakeHealthEndpoint(svc),
		hello.DecodeHealthRequest,
		hello.EncodeResponse,
	))

	registar := hello.Register("10.71.8.125", "8500")

	// HTTP transport
	go func() {
		log.Println("httpAddress", *httpAddr)
		registar.Register()
		handler := r
		errChan <- http.ListenAndServe(*httpAddr, handler)
	}()


	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <- errChan
	registar.Deregister()
	log.Fatalln(error)
}