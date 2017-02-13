package hello

type Service interface {
	HealthCheck() bool
	SayHello(name string) string
}

type HelloService struct {

}

func (HelloService) SayHello(name string) string  {
	return "Hello " + name
}

func (HelloService) HealthCheck() bool {
	return true
}