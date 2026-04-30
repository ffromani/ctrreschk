// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv0 "github.com/ffromani/ctrreschk/api/v0"
)

var _ = Describe("ctrreschk alignment", func() {
	It("should report alignment for a Guaranteed QoS pod", func(ctx context.Context) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ctrreschk-guaranteed",
				Namespace: testNamespace,
			},
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				Containers: []corev1.Container{
					{
						Name:            "ctrreschk",
						Image:           testImage,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command:         []string{"/ctrreschk", "align"},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("2"),
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("2"),
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
						},
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

		var result apiv0.Allocation
		err := json.Unmarshal([]byte(logs), &result)
		Expect(err).NotTo(HaveOccurred(), "pod output should be valid JSON, got: %s", logs)

		Expect(result.Aligned).NotTo(BeNil(), "aligned section should be present")
		Expect(result.Aligned.NUMA).NotTo(BeEmpty(), "aligned NUMA info should have at least one entry")

		for numaID, details := range result.Aligned.NUMA {
			Expect(details.CPUs).NotTo(BeEmpty(), "NUMA node %d should have CPUs listed", numaID)
		}
	})
})
