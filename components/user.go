/*
Copyright 2020 Noah Kantrowitz
Copyright 2019-2020 Ridecell, Inc.

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
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"io/ioutil"
	"time"

	cu "github.com/coderanger/controller-utils"
	rabbithole "github.com/michaelklishin/rabbit-hole"
	"github.com/pkg/errors"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

const DEFAULT_HASH_ALGORITHM = rabbithole.HashingAlgorithmSHA256

type userComponent struct {
	clientFactory rabbitClientFactory
}

func User() *userComponent {
	return &userComponent{clientFactory: rabbitholeClientFactory}
}

func (comp *userComponent) Reconcile(ctx *cu.Context) (cu.Result, error) {
	obj := ctx.Object.(*rabbitv1beta1.RabbitUser)
	ctx.Conditions.SetUnknown("UserReady", "Unknown")

	// Connect to the RabbitMQ server.
	rmqc, err := connect(ctx, &obj.Spec.Connection, obj.Namespace, ctx.Client, comp.clientFactory)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error connecting to rabbitmq")
	}

	// Get the core data for the user from the object/context.
	username := obj.Spec.Username
	if username == "" { // TODO Switch this to a defaulting webhook.
		username = obj.Name
	}
	password, ok := ctx.Data.GetString("password")
	if !ok {
		return cu.Result{}, errors.Wrap(err, "user password not set in context")
	}

	// Get the existing user data, if any.
	var createUser, updateUser bool
	existingUser, err := rmqc.GetUser(username)
	if err != nil {
		rabbitErr, ok := err.(rabbithole.ErrorResponse)
		if ok && rabbitErr.StatusCode == 404 {
			createUser = true
		} else {
			return cu.Result{}, errors.Wrapf(err, "error getting user %s", username)
		}
	} else {
		// Diff the existing user.
		if obj.Spec.Tags != existingUser.Tags {
			updateUser = true
		}
		hashedPassword, err := hashRabbitPassword(password, existingUser.HashingAlgorithm, existingUser.PasswordHash)
		if err != nil {
			// ??? Should this actually error? It could just mark for update and let it get overwritten.
			return cu.Result{}, errors.Wrap(err, "error hashing password for comparison")
		}
		if hashedPassword != existingUser.PasswordHash {
			updateUser = true
		}
	}

	if createUser || updateUser {
		// Always rehash even for an update so we get a new salt.
		hashedPassword, err := hashRabbitPassword(password, DEFAULT_HASH_ALGORITHM, "")
		if err != nil {
			return cu.Result{}, errors.Wrap(err, "error hashing password for put")
		}

		// Put the user, this will create or update depending on if the user already exists.
		resp, err := rmqc.PutUser(username, rabbithole.UserSettings{PasswordHash: hashedPassword, HashingAlgorithm: DEFAULT_HASH_ALGORITHM, Tags: obj.Spec.Tags})
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error putting user %s", username)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return cu.Result{}, errors.Wrapf(err, "error reading PutUser response: %s", resp.Status)
			}
			return cu.Result{}, errors.Errorf("error putting user %s: %s %s", username, resp.Status, body)
		}

		// Create an event.
		var event, eventMessage string
		if createUser {
			event = "UserCreated"
			eventMessage = "created"
		} else {
			event = "UserUpdated"
			eventMessage = "updated"
		}
		ctx.Events.Eventf(obj, "Normal", event, "RabbitMQ user %s %s", username, eventMessage)

		if resp.StatusCode == 201 {
			// If this is the initial creation of the user reconcile again after 10 seconds
			// This is a hack to remedy amqp permissions being applied incorrectly immediately after creation.
			ctx.Conditions.SetfFalse("UserReady", "UserPending", "RabbitMQ user %s has been created", username)
			return cu.Result{RequeueAfter: time.Second * 10, SkipRemaining: true}, nil
		}
	}

	// Good to go.
	ctx.Conditions.SetfTrue("UserReady", "UserExists", "RabbitMQ user %s exists", username)
	return cu.Result{}, nil
}

var hashAlgorithms = map[rabbithole.HashingAlgorithm]func() hash.Hash{
	rabbithole.HashingAlgorithmSHA256: sha256.New,
	rabbithole.HashingAlgorithmSHA512: sha512.New,
}

func hashRabbitPassword(password string, algorithm rabbithole.HashingAlgorithm, existingHash string) (string, error) {
	var salt []byte
	if existingHash == "" {
		salt = make([]byte, 4)
		_, err := rand.Read(salt)
		if err != nil {
			return "", errors.Wrap(err, "error generating salt")
		}
	} else {
		decodedExistingHash := make([]byte, base64.StdEncoding.DecodedLen(len(existingHash)))
		_, err := base64.StdEncoding.Decode(decodedExistingHash, []byte(existingHash))
		if err != nil {
			return "", errors.Wrap(err, "error decoding existing hash")
		}
		salt = decodedExistingHash[:4]
	}

	hashFactory, ok := hashAlgorithms[algorithm]
	if !ok {
		return "", errors.Errorf("unknown algorithm %s", algorithm)
	}
	hash := hashFactory()
	_, _ = hash.Write(salt)
	_, _ = hash.Write([]byte(password))
	hashed := hash.Sum(nil)

	return base64.StdEncoding.EncodeToString(append(salt, hashed...)), nil
}
