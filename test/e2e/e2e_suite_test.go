// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset     kubernetes.Interface
	testImage     string
	testNamespace string
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ctrreschk E2E Suite")
}

var _ = BeforeSuite(func() {
	testNamespace = "default"
	By("using the test namespace: " + testNamespace)

	testImage = os.Getenv("CTRRESCHK_E2E_IMAGE")
	Expect(testImage).NotTo(BeEmpty(), "CTRRESCHK_E2E_IMAGE must be set")
	By("using the test image: " + testImage)

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.Getenv("HOME") + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	Expect(err).NotTo(HaveOccurred(), "failed to build kubeconfig")

	clientset, err = kubernetes.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred(), "failed to create kubernetes client")
})

func createPod(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, func(context.Context)) {
	GinkgoHelper()
	created, err := clientset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred(), "failed to create pod %s", pod.Name)
	return created, func(dctx context.Context) {
		derr := clientset.CoreV1().Pods(created.Namespace).Delete(dctx, created.Name, metav1.DeleteOptions{})
		if derr != nil {
			fmt.Fprintf(GinkgoWriter, "warning: failed to delete pod %s/%s: %v\n", created.Namespace, created.Name, derr)
		}
	}
}

func waitForPodDone(ctx context.Context, namespace, name string) *corev1.Pod {
	GinkgoHelper()
	var pod *corev1.Pod
	Eventually(func(g Gomega) {
		var err error
		pod, err = clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(pod.Status.Phase).To(BeElementOf(corev1.PodSucceeded, corev1.PodFailed),
			"pod %s phase is %s", name, pod.Status.Phase)
	}).WithTimeout(2 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())
	return pod
}

func getPodLogs(ctx context.Context, namespace, name string) string {
	GinkgoHelper()
	req := clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{})
	data, err := req.DoRaw(ctx)
	Expect(err).NotTo(HaveOccurred(), "failed to get logs for pod %s", name)
	return string(data)
}
