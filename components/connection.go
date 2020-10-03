/*
Copyright 2020 Noah Kantrowitz
Copyright 2018-2019 Ridecell, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package components

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	rabbithole "github.com/michaelklishin/rabbit-hole"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

type rabbitManager interface {
	ListVhosts() ([]rabbithole.VhostInfo, error)
	GetVhost(string) (*rabbithole.VhostInfo, error)
	PutVhost(string, rabbithole.VhostSettings) (*http.Response, error)
	ListUsers() ([]rabbithole.UserInfo, error)
	GetUser(string) (*rabbithole.UserInfo, error)
	PutUser(string, rabbithole.UserSettings) (*http.Response, error)
	ListPoliciesIn(vhost string) (rec []rabbithole.Policy, err error)
	PutPolicy(vhost string, name string, policy rabbithole.Policy) (res *http.Response, err error)
	DeletePolicy(vhost string, name string) (res *http.Response, err error)
	ListPermissionsOf(username string) (rec []rabbithole.PermissionInfo, err error)
	UpdatePermissionsIn(vhost, username string, permissions rabbithole.Permissions) (res *http.Response, err error)
	ClearPermissionsIn(vhost, username string) (res *http.Response, err error)
}

type rabbitClientFactory func(uri string, user string, pass string, t *http.Transport) (rabbitManager, error)

// Implementation of rabbitMQClientFactory using rabbithole (i.e. a real client).
func rabbitholeClientFactory(uri string, user string, pass string, t *http.Transport) (rabbitManager, error) {
	return rabbithole.NewTLSClient(uri, user, pass, t)
}

// Open a connection to the RabbitMQ server as defined by a RabbitmqConnection object.
func connect(ctx context.Context, connection *rabbitv1beta1.RabbitConnection, namespace string, client client.Client, clientFactory rabbitClientFactory) (rabbitManager, *url.URL, error) {
	defaults, err := url.Parse(os.Getenv("DEFAULT_CONNECTION"))
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to parse $DEFAULT_CONNECTION")
	}

	protocol := connection.Protocol
	if protocol == "" {
		protocol = defaults.Scheme
	}
	if protocol == "" {
		protocol = "amqp"
	}

	host := connection.Host
	if host == "" {
		host = defaults.Hostname()
	}
	if host == "" {
		return nil, nil, errors.New("host is required")
	}

	port := connection.Port
	if port == 0 {
		port, _ = strconv.Atoi(defaults.Port())
	}
	// No overall default, leave it as 0 to be handled later.

	user := connection.Username
	if user == "" {
		user = defaults.User.Username()
	}
	if user == "" {
		return nil, nil, errors.New("username is required")
	}

	var password string
	if connection.PasswordSecretRef != nil {
		secret := &corev1.Secret{}
		err := client.Get(ctx, types.NamespacedName{Name: connection.PasswordSecretRef.Name, Namespace: namespace}, secret)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error getting password secret %s/%s", namespace, connection.PasswordSecretRef.Name)
		}
		key := connection.PasswordSecretRef.Key
		if key == "" {
			key = "password"
		}
		passwordBytes, ok := secret.Data[key]
		if !ok {
			return nil, nil, errors.Errorf("key %s not found in password secret %s/%s", key, namespace, connection.PasswordSecretRef.Name)
		}
		password = string(passwordBytes)
	} else if defaultPassword, ok := defaults.User.Password(); ok {
		password = defaultPassword
	} else if defaultPassword, ok := os.LookupEnv("DEFAULT_CONNECTION_PASSWORD"); ok {
		password = defaultPassword
	}
	// No error for blank password since that is kind of allowed, though a bad idea.

	var insecure bool
	if connection.InsecureSkipVerify != nil {
		insecure = *connection.InsecureSkipVerify
	} else if defaultInsecure := defaults.Query().Get("insecureSkipVerify"); defaultInsecure != "" {
		insecure = (defaultInsecure == "true")
	} else {
		// Already the default zero but just to make it crystal clear.
		insecure = false
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	var hostAndPort string
	if port == 0 {
		hostAndPort = host
	} else {
		hostAndPort = fmt.Sprintf("%s:%d", host, port)
	}

	compiledUri := &url.URL{Scheme: protocol, Host: hostAndPort, User: url.UserPassword(user, password)}

	// TODO connection pooling? Would need to persist the Transport object.

	// Connect to the rabbitmq cluster
	rmqc, err := clientFactory(compiledUri.String(), user, password, transport)
	return rmqc, compiledUri, err
}
