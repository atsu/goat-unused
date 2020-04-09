package k8s

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getLocalKubeConf(t *testing.T) *rest.Config {
	t.Helper()
	h := os.Getenv("HOME")
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(h, ".kube", "config"))
	if err != nil {
		t.Fatal(err)
	}
	return config
}

// Initial manual testing
func TestDefaultOpenshiftController_Manual(t *testing.T) {
	t.SkipNow()

	osc := NewDefaultOpenshiftController()
	config := getLocalKubeConf(t)
	if err := osc.Authenticate(config); err != nil {
		t.Fatal(err)
	}
	list, err := osc.ListObjects("", v1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) < 1 {
		t.FailNow()
	}
	fmt.Printf("found %d items\n", len(list.Items))
	watcher, err := osc.WatchEvents("", metav1.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	doneCh := make(chan int)
	events := make(map[watch.EventType]int)
	go func() {
		for {
			select {
			case <-time.After(time.Second * 30):
				watcher.Stop()
				close(doneCh)
				return
			case evt := <-watcher.ResultChan():
				if cnt, ok := events[evt.Type]; ok {
					events[evt.Type] = cnt + 1
				} else {
					events[evt.Type] = 1
				}
			}
		}
	}()
	<-doneCh
	fmt.Println("Events: ", events)

}
