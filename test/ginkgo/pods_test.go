package kubernetes_test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pods", func() {
  var (
    kubeClient *client.Client
    err error
  )

  BeforeEach(func() {
    kubeClient, err = client.New(&client.Config{Host: "http://localhost:8080"})
    if err != nil {
      Fail("Failed to create client")
    }
  })
 

  It("should submit and remove a pod", func() {
    podClient := kubeClient.Pods(api.NamespaceDefault)
    pod := loadPodOrDie(assetPath("api", "examples", "pod.json"))
    value := strconv.Itoa(time.Now().Nanosecond())
    pod.Labels["e2esuite"] = value
    _, err = podClient.Create(pod)
    if err != nil {
      Fail(fmt.Sprintf("Failed to create pod: %v", err))
    }
    defer podClient.Delete(pod.Name)
    pods, err := podClient.List(labels.SelectorFromSet(labels.Set(map[string]string{"e2esuite": value})))
    if err != nil {
      Fail(fmt.Sprintf("Failed to query for pods: %v", err))
    }
    Expect(len(pods.Items)).To(Equal(1))
    podClient.Delete(pod.Name)
    pods, err = podClient.List(labels.SelectorFromSet(labels.Set(map[string]string{"e2esuite": value})))
    Expect(len(pods.Items)).To(Equal(0))
  })

  It("should update a pod", func() {
    podClient := kubeClient.Pods(api.NamespaceDefault)
    pod := loadPodOrDie(assetPath("api", "examples", "pod.json"))
    value := strconv.Itoa(time.Now().Nanosecond())
    pod.Labels["e2esuite"] = value
    _, err = podClient.Create(pod)
    if err != nil {
      Fail(fmt.Sprintf("Failed to create pod: %v", err))
    }
    defer podClient.Delete(pod.Name)
    waitForPodRunning(kubeClient, pod.Name, 60)
    pods, err := podClient.List(labels.SelectorFromSet(labels.Set(map[string]string{"e2esuite": value})))
    if err != nil {
      Fail(fmt.Sprintf("Failed to query for pods: %v", err))
    }
    Expect(len(pods.Items)).To(Equal(1))
    podOut, err := podClient.Get(pod.Name)
    if err != nil {
      Fail(fmt.Sprintf("Failed to get pod: %v", err))
    }
    value = "time" + value
    pod.Labels["time"] = value
    pod.ResourceVersion = podOut.ResourceVersion
    pod.UID = podOut.UID
    pod, err = podClient.Update(pod)
    if err != nil {
      Fail(fmt.Sprintf("Failed to update pod: %v", err))
    }
    waitForPodRunning(kubeClient, pod.Name, 60)
    pods, err = podClient.List(labels.SelectorFromSet(labels.Set(map[string]string{"time": value})))
    Expect(len(pods.Items)).To(Equal(1))
  })
})
