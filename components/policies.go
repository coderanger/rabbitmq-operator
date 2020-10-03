/*
Copyright 2020 Noah Kantrowitz
Copyright 2018 Ridecell, Inc.

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
	"encoding/json"
	"fmt"
	"reflect"

	cu "github.com/coderanger/controller-utils"
	rabbithole "github.com/michaelklishin/rabbit-hole"
	"github.com/pkg/errors"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

type policiesComponent struct {
	clientFactory rabbitClientFactory
}

func Policies() *policiesComponent {
	return &policiesComponent{clientFactory: rabbitholeClientFactory}
}

func (comp *policiesComponent) Reconcile(ctx *cu.Context) (cu.Result, error) {
	obj := ctx.Object.(*rabbitv1beta1.RabbitVhost)
	ctx.Conditions.SetUnknown("PoliciesReady", "Unknown")

	// Connect to the RabbitMQ server.
	rmqc, _, err := connect(ctx, &obj.Spec.Connection, obj.Namespace, ctx.Client, comp.clientFactory)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error connecting to rabbitmq")
	}

	// Get the core data for the vhost from the object/context.
	vhost := obj.Spec.VhostName
	if vhost == "" { // TODO Switch this to a defaulting webhook.
		vhost = obj.Name
	}

	// Process the spec policies into a more usable state.
	desiredPolicies := map[string]*rabbithole.Policy{}
	for name, specPolicy := range obj.Spec.Policies {
		policy := &rabbithole.Policy{
			Vhost:      vhost,
			Pattern:    specPolicy.Pattern,
			ApplyTo:    specPolicy.ApplyTo,
			Name:       fmt.Sprintf("%s-%s", vhost, name),
			Priority:   specPolicy.Priority,
			Definition: map[string]interface{}{},
		}
		for defKey, rawDefValue := range specPolicy.Definition {
			var val interface{}
			// TODO in a validate webhook, check that the value is of a type that we know is okay.
			err := json.Unmarshal(rawDefValue.Raw, &val)
			if err != nil {
				return cu.Result{}, errors.Wrapf(err, "error decoding defintiion %s value %s for vhost %s/%s", defKey, string(rawDefValue.Raw), obj.Namespace, obj.Name)
			}
			policy.Definition[defKey] = val
		}
		desiredPolicies[policy.Name] = policy
	}

	// Grab and process the existing policies.
	existingPolicyList, err := rmqc.ListPoliciesIn(vhost)
	if err != nil {
		return cu.Result{}, errors.Wrapf(err, "error fetching policies for vhost %s", vhost)
	}
	existingPolicies := map[string]*rabbithole.Policy{}
	for _, existingPolicy := range existingPolicyList {
		existingPolicies[existingPolicy.Name] = &existingPolicy
	}

	// Double-ended diff the two sets of policies.
	var createPolicies, updatePolicies []*rabbithole.Policy
	var deletePolicies []string
	for name, policy := range desiredPolicies {
		existingPolicy, ok := existingPolicies[name]
		if !ok {
			createPolicies = append(createPolicies, policy)
		} else if !reflect.DeepEqual(*policy, *existingPolicy) {
			updatePolicies = append(updatePolicies, policy)
		}
	}
	for name := range existingPolicies {
		_, ok := desiredPolicies[name]
		if !ok {
			deletePolicies = append(deletePolicies, name)
		}
	}

	// Create needed policies.
	for _, policy := range createPolicies {
		_, err = rmqc.PutPolicy(vhost, policy.Name, *policy)
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error creating policy %s for vhost %s", policy.Name, vhost)
		}
		ctx.Events.Eventf(obj, "Normal", "PolicyCreated", "RabbitMQ policy %s for vhost %s created", policy.Name, vhost)
	}

	// Update needed polices.
	for _, policy := range updatePolicies {
		_, err = rmqc.PutPolicy(vhost, policy.Name, *policy)
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error updating policy %s for vhost %s", policy.Name, vhost)
		}
		ctx.Events.Eventf(obj, "Normal", "PolicyUpdated", "RabbitMQ policy %s for vhost %s updated", policy.Name, vhost)
	}

	// Delete unneeeded policies.
	for _, policy := range deletePolicies {
		_, err = rmqc.DeletePolicy(vhost, policy)
		if err != nil {
			return cu.Result{}, errors.Wrapf(err, "error deleting policy %s for vhost %s", policy, vhost)
		}
		ctx.Events.Eventf(obj, "Normal", "PolicyDeleted", "RabbitMQ policy %s for vhost %s deleted", policy, vhost)
	}

	ctx.Conditions.SetTrue("PoliciesReady", "PoliciesSynced")
	return cu.Result{}, nil
}
