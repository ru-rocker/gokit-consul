# gokit-consul
Gokit example integrated with consul

# Run Consul
For development purpose, consul is executed via docker

    docker run --rm -p 8400:8400 -p 8500:8500 -p 8600:53/udp -h node1 progrium/consul -server -bootstrap -ui-dir /ui -advertise 10.71.8.125

# Register Service to consul

    go run src/github.com/ru-rocker/hello/hello.d/main.go

# Discover or subscribe from consul

    go run src/github.com/ru-rocker/hello/discover.d/main.go

# File structure
* `service.go` : service / business logic
* `transport.go` : make endpoints and json encode/decode
* `discovery.go` : register service to consul.

### Notes:
For this example, there are two endpoints. First one is hello endpoint, for saying hello.
The other one is health endpoint. This endpoint is intended for consul health checks.

For another microservice utilities, such us prometheus or log tracing will be implemented soon.
Or using another project.