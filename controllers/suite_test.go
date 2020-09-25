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
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/coderanger/controller-utils/tests"
	rabbitmqv1beta1 "github.com/coderanger/rabbitmq-operator/api/v1beta1"
	// +kubebuilder:scaffold:imports
)

var suiteHelper *tests.FunctionalSuiteHelper

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	suiteHelper = tests.Functional().
		CRDPath(filepath.Join("..", "config", "crd", "bases")).
		API(rabbitmqv1beta1.AddToScheme).
		MustBuild()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	suiteHelper.MustStop()
})
