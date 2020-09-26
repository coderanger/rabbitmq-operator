module github.com/coderanger/rabbitmq-operator

go 1.15

require (
	github.com/coderanger/controller-utils v0.0.0-20200925022637-d66f965e832d
	github.com/go-logr/logr v0.1.0
	github.com/michaelklishin/rabbit-hole v1.5.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/streadway/amqp v1.0.0 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
