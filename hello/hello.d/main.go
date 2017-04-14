package main

import (
	ht "github.com/go-kit/kit/transport/http"
	"net/http"
	"github.com/gorilla/mux"
	"context"
	"log"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/ru-rocker/gokit-consul/hello"
)

// to execute: go run src/github.com/ru-rocker/gokit-consul/hello/hello.d/main.go -consul.addr 172.20.20.30 -consul.port 8500 -advertise.addr 10.71.6.68 -advertise.port 7002
func main() {
	ctx := context.Background()

	var (
		consulAddr = flag.String("consul.addr", "", "consul address")
		consulPort = flag.String("consul.port", "", "consul port")
		advertiseAddr = flag.String("advertise.addr", "", "advertise address")
		advertisePort = flag.String("advertise.port", "", "advertise port")
	)
	flag.Parse()

	errChan := make(chan error)

	var svc hello.Service
	svc = hello.HelloService{}

	r := mux.NewRouter()

	r.Handle("/hello", ht.NewServer(
		hello.MakeHelloEndpoint(svc),
		hello.DecodeHelloRequest,
		hello.EncodeResponse,
	))

	r.Methods("GET").Path("/health").Handler(ht.NewServer(
		hello.MakeHealthEndpoint(svc),
		hello.DecodeHealthRequest,
		hello.EncodeResponse,
	))

	registar := hello.Register(*consulAddr, *consulPort, *advertiseAddr, *advertisePort)

	// HTTP transport
	go func() {
		log.Println("httpAddress", *advertisePort)
		registar.Register()
		handler := r
		errChan <- http.ListenAndServe(":" + *advertisePort, handler)
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