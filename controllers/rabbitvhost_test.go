/*
Copyright 2020 Noah Kantrowitz

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

package controllers

import (
	"context"

	cu "github.com/coderanger/controller-utils"
	"github.com/coderanger/controller-utils/conditions"
	"github.com/coderanger/controller-utils/randstring"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("RabbitVhost controller", func() {
	var helper *cu.FunctionalHelper
	var rmqc *rabbithole.Client

	BeforeEach(func() {
		helper = suiteHelper.MustStart(RabbitVhost, func(mgr manager.Manager) error {
			// We need the webhook running because the functional helpers install all the webhook manifests, not just for this type.
			return ctrl.NewWebhookManagedBy(mgr).For(&rabbitv1beta1.RabbitUser{}).Complete()
		})
		rmqc = connect()
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			helper.DebugList(&rabbitv1beta1.RabbitVhostList{})
			helper.DebugList(&rabbitv1beta1.RabbitUserList{})
		}
		helper.MustStop()
		helper = nil
	})

	It("runs a basic reconcile", func() {
		c := helper.TestClient

		vhost := &rabbitv1beta1.RabbitVhost{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitVhostSpec{
				VhostName: "testing-" + randstring.MustRandomString(5),
			},
		}
		c.Create(vhost)
		// Advance the user to ready.
		user := &rabbitv1beta1.RabbitUser{}
		c.EventuallyGetName("testing", user)
		conditions.SetStatusCondition(&user.Status.Conditions, conditions.Condition{
			Type:   "Ready",
			Status: metav1.ConditionTrue,
			Reason: "Fake",
		})
		c.Status().Update(user)

		// Wait for ready.
		c.EventuallyGetName("testing", vhost, c.EventuallyReady())
		Expect(vhost.Finalizers).To(ContainElement("rabbitvhost.rabbitmq.coderanger.net/vhost"))

		// Check that the vhost exists
		_, err := rmqc.GetVhost(vhost.Spec.VhostName)
		Expect(err).ToNot(HaveOccurred())

		// Check that the automatic user was created.
		c.GetName("testing", user)
		Expect(user.Spec.Permissions).To(ConsistOf(
			rabbitv1beta1.RabbitPermission{
				Vhost:     vhost.Spec.VhostName,
				Read:      ".*",
				Write:     ".*",
				Configure: ".*",
			},
		))

		// Delete the vhost and make sure it is cleaned up.
		c.Delete(vhost)
		Eventually(func() bool {
			err := helper.Client.Get(context.Background(), types.NamespacedName{Name: "testing", Namespace: helper.Namespace}, vhost)
			return err != nil && kerrors.IsNotFound(err)
		}).Should(BeTrue())
		_, err = rmqc.GetVhost(vhost.Spec.VhostName)
		Expect(err).To(HaveOccurred())
		rmqErr := err.(rabbithole.ErrorResponse)
		Expect(rmqErr.StatusCode).To(Equal(404))
	})
})
