package kubernetes_test

import (
	"fmt"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"

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
	if testContext.provider == "vagrant" {
		By("Skipping test which is broken for vagrant (See https://github.com/GoogleCloudPlatform/kubernetes/issues/3580)")
	} else {
		ns := api.NamespaceDefault
		// TODO(satnam6502): Replace call of randomSuffix with call to NewUUID when service
		//                   names have the same form as pod and replication controller names.
		name := "nettest-" + randomSuffix()
		svc, err := kubeClient.Services(ns).Create(&api.Service{
			ObjectMeta: api.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					"name": name,
				},
			},
			Spec: api.ServiceSpec{
				Port:          8080,
				ContainerPort: util.NewIntOrStringFromInt(8080),
				Selector: map[string]string{
					"name": name,
				},
			},
		})
		By(fmt.Sprintf("Creating service with name %s", svc.Name))
		if err != nil {
			Fail(fmt.Sprintf("unable to create test service %s: %v", svc.Name, err))
		}
		// Clean up service
		defer func() {
			defer GinkgoRecover()
			if err = kubeClient.Services(ns).Delete(svc.Name); err != nil {
				Fail(fmt.Sprintf("unable to delete svc %v: %v", svc.Name, err))
			}
		}()
		rc, err := kubeClient.ReplicationControllers(ns).Create(&api.ReplicationController{
			ObjectMeta: api.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					"name": name,
				},
			},
			Spec: api.ReplicationControllerSpec{
				Replicas: 8,
				Selector: map[string]string{
					"name": name,
				},
				Template: &api.PodTemplateSpec{
					ObjectMeta: api.ObjectMeta{
						Labels: map[string]string{"name": name},
					},
					Spec: api.PodSpec{
						Containers: []api.Container{
							{
								Name:    "webserver",
								Image:   "kubernetes/nettest:latest",
								Command: []string{"-service=" + name},
								Ports:   []api.Port{{ContainerPort: 8080}},
							},
						},
					},
				},
			},
		})
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
		passed := false
		Loop:
		for i := 0; i < maxAttempts; i++ {
			time.Sleep(2 * time.Second)
			body, err = kubeClient.Get().Prefix("proxy").Resource("services").Name(svc.Name).Suffix("status").Do().Raw()
			if err != nil {
				fmt.Println(fmt.Sprintf("Attempt %v/%v: service/pod still starting. (error: '%v')", i, maxAttempts, err))
			}
			switch string(body) {
			case "pass":
				fmt.Println(fmt.Sprintf("Passed on attempt %v. Cleaning up.", i))
				passed = true
				break Loop
			case "running":
				fmt.Println(fmt.Sprintf("Attempt %v/%v: test still running", i, maxAttempts))
			case "fail":
				if body, err = kubeClient.Get().Prefix("proxy").Resource("services").Name(svc.Name).Suffix("read").Do().Raw(); err != nil {
				Fail(fmt.Sprintf("Failed on attempt %v. Cleaning up. Error reading details: %v", i, err))
				} else {
					Fail(fmt.Sprintf("Failed on attempt %v. Cleaning up. Details:\n%v", i, string(body)))
				}
				break Loop
			}
		}

		if !passed {
			if body, err = kubeClient.Get().Prefix("proxy").Resource("services").Name(svc.Name).Suffix("read").Do().Raw(); err != nil {
				Fail(fmt.Sprintf("Timed out. Cleaning up. Error reading details: %v", err))
			} else {
				Fail(fmt.Sprintf("Timed out. Cleaning up. Details:\n%v", string(body)))
			}
		}
		Expect(string(body)).To(Equal("pass"))
	}
  })
})
