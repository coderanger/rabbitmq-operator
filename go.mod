module github.com/coderanger/rabbitmq-operator

go 1.15

replace sigs.k8s.io/controller-runtime => github.com/coderanger/controller-runtime v0.2.0-beta.1.0.20201115004253-9bec1fefa8ca

require (
	github.com/coderanger/controller-utils v0.0.0-20201221100905-e26c5734ecc9
	github.com/michaelklishin/rabbit-hole/v2 v2.0.0-20201216035320-4572900f3492
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546 // indirect
	github.com/streadway/amqp v1.0.0
	golang.org/x/tools v0.0.0-20200616195046-dc31b401abb5 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3
)
