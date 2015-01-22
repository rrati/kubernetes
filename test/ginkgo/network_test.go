package kubernetes_test

import (
	"fmt"
//	"io/ioutil"
//	"path/filepath"
//	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
//	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/latest"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
//	"github.com/GoogleCloudPlatform/kubernetes/pkg/labels"
//	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Networking", func() {
  var (
    kubeClient *client.Client
    err error
    body []byte
  )

  BeforeEach(func() {
    kubeClient, err = client.New(&client.Config{Host: "http://localhost:8080"})
    if err != nil {
      Fail("Failed to create client")
    }
  })
 

  It("should verify network functions", func() {
    ns := api.NamespaceDefault
    svc, err := kubeClient.Services(ns).Create(loadObjectOrDie(assetPath(
                "contrib", "for-tests", "network-tester", "service.json",
                )).(*api.Service))
    if err != nil {
      Fail(fmt.Sprintf("unable to create test service: %v", err))
    }
    // Clean up service
    defer func() {
      if err = kubeClient.Services(ns).Delete(svc.Name); err != nil {
        defer GinkgoRecover()
        Fail(fmt.Sprintf("unable to delete svc %v: %v", svc.Name, err))
      }
    }()

    rc, err := kubeClient.ReplicationControllers(ns).Create(loadObjectOrDie(assetPath(
               "contrib", "for-tests", "network-tester", "rc.json",
               )).(*api.ReplicationController))
    if err != nil {
      Fail(fmt.Sprintf("unable to create test rc: %v", err))
    }
    // Clean up rc
    defer func() {
      defer GinkgoRecover()
      rc.Spec.Replicas = 0
      rc, err = kubeClient.ReplicationControllers(ns).Update(rc)
      if err != nil {
        Fail(fmt.Sprintf("unable to modify replica count for rc %v: %v", rc.Name, err))
      }
      if err = kubeClient.ReplicationControllers(ns).Delete(rc.Name); err != nil {
        Fail(fmt.Sprintf("unable to delete rc %v: %v", rc.Name, err))
      }
    }()

    const maxAttempts = 60
    for i := 0; i < maxAttempts; i++ {
      time.Sleep(time.Second)
      body, err = kubeClient.Get().Prefix("proxy").Resource("services").Name(svc.Name).Suffix("status").Do().Raw()
      if err != nil {
        continue
      }
      switch string(body) {
        case "pass":
Fail("Test passed")
          break
        case "running":
          continue
        case "fail":
          if body, err = kubeClient.Get().Prefix("proxy").Resource("services").Name(svc.Name).Suffix("read").Do().Raw(); err != nil {
            Fail(fmt.Sprintf("Failed on attempt %v. Cleaning up. Error reading details: %v", i, err))
          } else {
            Fail(fmt.Sprintf("Failed on attempt %v. Cleaning up. Details:\n%v", i, string(body)))
          }
      }
    }
/*
    if body, err = kubeClient.Get().Prefix("proxy").Resource("services").Name(svc.Name).Suffix("read").Do().Raw(); err != nil {
      Fail(fmt.Sprintf("Timed out. Cleaning up. Error reading details: %v", err))
    } else {
      Fail(fmt.Sprintf("Timed out. Cleaning up. Details:\n%v", string(body)))
    }
*/
    Expect(string(body)).To(Equal("pass"))
  })
})
