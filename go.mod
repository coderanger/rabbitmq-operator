module github.com/coderanger/rabbitmq-operator

go 1.15

replace sigs.k8s.io/controller-runtime => github.com/coderanger/controller-runtime v0.2.0-beta.1.0.20201115004253-9bec1fefa8ca

require (
	github.com/coderanger/controller-utils v0.0.0-20201009072932-baf39c5caeb1
	github.com/michaelklishin/rabbit-hole v1.5.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546 // indirect
	github.com/streadway/amqp v1.0.0
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/controller-tools v0.4.0 // indirect
)
