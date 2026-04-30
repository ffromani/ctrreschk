// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv0 "github.com/ffromani/ctrreschk/api/v0"
)

var _ = Describe("ctrreschk alignmem smoke test", func() {
	It("should run and produce valid NUMAMapsInfo JSON", func(ctx context.Context) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ctrreschk-alignmem-smoke",
				Namespace: testNamespace,
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:            "ctrreschk",
						Image:           testImage,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/ctrreschk", "alignmem"},
					},
				},
			},
		}

		created, deletePod := createPod(ctx, pod)
		DeferCleanup(deletePod, ctx)

		finished := waitForPodDone(ctx, created.Namespace, created.Name)
		Expect(finished.Status.Phase).To(Equal(corev1.PodSucceeded), "pod should complete successfully")

		logs := getPodLogs(ctx, created.Namespace, created.Name)
		Expect(logs).NotTo(BeEmpty(), "pod logs should not be empty")

		var result apiv0.NUMAMapsInfo
		err := json.Unmarshal([]byte(logs), &result)
		Expect(err).NotTo(HaveOccurred(), "pod output should be valid JSON, got: %s", logs)

		Expect(result.Nodes).NotTo(BeEmpty(), "should report at least one NUMA node with pages")
		Expect(result.LocalPages).To(BeNumerically(">", 0), "should have local pages")
	})
})
