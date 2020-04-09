package k8s

import (
	"log"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// More manual for now
func TestDefaultK8Controller_ListNamespaces(t *testing.T) {
	t.SkipNow()

	conf := getLocalKubeConf(t)
	osc := NewDefaultOpenshiftController()
	if err := osc.Authenticate(conf); err != nil {
		t.Fatal(err)
	}

	ns, err := osc.ListNamespaces(v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range ns.Items {
		log.Println(n.Name)
	}
}
