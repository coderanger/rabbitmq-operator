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
	"github.com/coderanger/controller-utils/randstring"
	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	rabbitv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
)

var _ = Describe("RabbitUser controller", func() {
	var helper *cu.FunctionalHelper
	var rmqc *rabbithole.Client

	BeforeEach(func() {
		helper = suiteHelper.MustStart(RabbitUser, func(mgr ctrl.Manager) error {
			// Launch the vhost webhooks because I need them for one test. Or rather I need creating a vhost object to not fail so I can test a watch map.
			return ctrl.NewWebhookManagedBy(mgr).For(&rabbitv1beta1.RabbitVhost{}).Complete()
		})
		rmqc = connect()
	})

	AfterEach(func() {
		helper.MustStop()
		helper = nil
	})

	It("runs a basic reconcile", func() {
		c := helper.TestClient

		user := &rabbitv1beta1.RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitUserSpec{
				Username: "testing-" + randstring.MustRandomString(5),
				Tags:     "management",
			},
		}
		c.Create(user)
		c.EventuallyGetName("testing", user, c.EventuallyReady())
		Expect(user.Finalizers).To(ContainElement("rabbituser.rabbitmq.coderanger.net/user"))

		secret := &corev1.Secret{}
		c.GetName("testing-rabbituser", secret)
		Expect(secret.Data).To(HaveKeyWithValue("RABBIT_PASSWORD", Not(BeEmpty())))
		Expect(secret.Data).To(HaveKeyWithValue("RABBIT_HOST", Not(BeEmpty())))
		rmqcUser := connectUser(user.Spec.Username, string(secret.Data["RABBIT_PASSWORD"]))
		Expect(rmqcUser.Whoami()).ToNot(BeNil())

		// No permissions, so shouldn't be able to see anything.
		vhosts, err := rmqcUser.ListVhosts()
		Expect(err).ToNot(HaveOccurred())
		Expect(vhosts).To(BeEmpty())

		// Delete the user and make sure it is cleaned up.
		c.Delete(user)
		Eventually(func() bool {
			err := helper.Client.Get(context.Background(), types.NamespacedName{Name: "testing", Namespace: helper.Namespace}, user)
			return err != nil && kerrors.IsNotFound(err)
		}).Should(BeTrue())
		_, err = rmqcUser.Whoami()
		Expect(err).To(HaveOccurred())
	})

	It("sets vhost permissions", func() {
		c := helper.TestClient

		user := &rabbitv1beta1.RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitUserSpec{
				Username: "testing-" + randstring.MustRandomString(5),
				Tags:     "administrator",
				Permissions: []rabbitv1beta1.RabbitPermission{
					{
						Vhost:     "/",
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				},
			},
		}
		c.Create(user)
		c.EventuallyGetName("testing", user, c.EventuallyReady())

		secret := &corev1.Secret{}
		c.GetName("testing-rabbituser", secret)
		rmqcUser := connectUser(user.Spec.Username, string(secret.Data["RABBIT_PASSWORD"]))
		Expect(rmqcUser.Whoami()).ToNot(BeNil())

		vhosts, err := rmqcUser.ListVhosts()
		Expect(err).ToNot(HaveOccurred())
		Expect(vhosts).ToNot(BeEmpty())
	})

	It("writes the vhost into the secret when using outputVhost", func() {
		c := helper.TestClient

		vhost := "testing-" + randstring.MustRandomString(5)
		_, err := rmqc.PutVhost(vhost, rabbithole.VhostSettings{})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			_, err := rmqc.DeleteVhost(vhost)
			Expect(err).ToNot(HaveOccurred())
		}()

		user := &rabbitv1beta1.RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitUserSpec{
				Username:    "testing-" + randstring.MustRandomString(5),
				Tags:        "management",
				OutputVhost: true,
				Permissions: []rabbitv1beta1.RabbitPermission{
					{
						Vhost:     vhost,
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				},
			},
		}
		c.Create(user)
		c.EventuallyGetName("testing", user, c.EventuallyReady())

		secret := &corev1.Secret{}
		c.GetName("testing-rabbituser", secret)
		Expect(secret.Data).To(HaveKeyWithValue("RABBIT_HOST", HaveSuffix("/"+vhost)))
	})

	It("sets vhost permissions on a newly created vhost when using * permissions", func() {
		c := helper.TestClient

		vhost := &rabbitv1beta1.RabbitVhost{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitVhostSpec{
				VhostName: "testing-" + randstring.MustRandomString(5),
				SkipUser:  true,
			},
		}

		user := &rabbitv1beta1.RabbitUser{
			ObjectMeta: metav1.ObjectMeta{Name: "testing"},
			Spec: rabbitv1beta1.RabbitUserSpec{
				Username: "testing-" + randstring.MustRandomString(5),
				Tags:     "administrator",
				Permissions: []rabbitv1beta1.RabbitPermission{
					{
						Vhost:     "*",
						Configure: ".*",
						Write:     ".*",
						Read:      ".*",
					},
				},
			},
		}
		c.Create(user)
		c.EventuallyGetName("testing", user, c.EventuallyReady())

		secret := &corev1.Secret{}
		c.GetName("testing-rabbituser", secret)
		rmqcUser := connectUser(user.Spec.Username, string(secret.Data["RABBIT_PASSWORD"]))

		// Confirm that we can't see the vhost.
		_, err := rmqcUser.GetVhost(vhost.Spec.VhostName)
		Expect(err).To(HaveOccurred())

		// Create the vhost object.
		c.Create(vhost)

		// Backdoor the actual creation since the vhost controller isn't running.
		_, err = rmqc.PutVhost(vhost.Spec.VhostName, rabbithole.VhostSettings{})
		Expect(err).ToNot(HaveOccurred())
		defer func() {
			_, err := rmqc.DeleteVhost(vhost.Spec.VhostName)
			Expect(err).ToNot(HaveOccurred())
		}()

		// Should immediately (or almost immediately) be able to see it because of the watch map.
		_, err = rmqcUser.GetVhost(vhost.Spec.VhostName)
		Expect(err).ToNot(HaveOccurred())
	})
})
