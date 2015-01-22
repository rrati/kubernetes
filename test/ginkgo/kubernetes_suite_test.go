package kubernetes_test

import (
        "fmt"
        "io/ioutil"
        "path/filepath"
        "time"

        "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
        "github.com/GoogleCloudPlatform/kubernetes/pkg/api/latest"
        "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
        "github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func assetPath(pathElements ...string) string {
  return filepath.Join("/home/rrati/repos/git/kubernetes", filepath.Join(pathElements...))
}

func loadObjectOrDie(filePath string) runtime.Object {
  defer GinkgoRecover()
  data, err := ioutil.ReadFile(filePath)
  if err != nil {
    Fail(fmt.Sprintf("Failed to read pod: %v", err))
  }
  obj, err := latest.Codec.Decode(data)
  if err != nil {
    Fail(fmt.Sprintf("Failed to decode pod: %v", err))
  }
  return obj
}

func loadPodOrDie(filePath string) *api.Pod {
  defer GinkgoRecover()
  obj := loadObjectOrDie(filePath)
  pod, ok := obj.(*api.Pod)
  if !ok {
    Fail(fmt.Sprintf("Failed to load pod: %v", obj))
  }
  return pod
}

func waitForPodRunning(c *client.Client, id string, timeout int) {
  timer := time.AfterFunc(time.Duration(timeout) * time.Second, func() {
    defer GinkgoRecover()
    Fail(fmt.Sprintf("Pod was not found running after %v seconds", timeout))
  })
  for {
    time.Sleep(5 * time.Second)
    pod, _ := c.Pods(api.NamespaceDefault).Get(id)
    if pod.Status.Phase == api.PodRunning {
      timer.Stop()
      break
    }
  }
}

func TestKubernetes(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubernetes Suite")
}
