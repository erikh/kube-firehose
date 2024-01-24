package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var Lock = &sync.RWMutex{}
var PodMap = map[string]context.CancelFunc{}

func main() {
	var kubeconfig *string
	var tail *bool
	var wait *time.Duration
	var since *time.Duration
	tail = flag.Bool("t", false, "new contents only; no history")
	wait = flag.Duration("wait", 0, "if supplied, will exit the program after this much time")
	since = flag.Duration("since", 0, "show this much time of history")
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	tick := time.Tick(time.Second)
	start := time.Now()
	logTime := metav1.NewTime(time.Unix(0, 0))

	if *tail || *since != 0 {
		logTime = metav1.NewTime(time.Now().Add(-(*since)))
	}

	podClient := clientset.CoreV1().Pods(apiv1.NamespaceDefault)

	for {
		pods, err := podClient.List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		for _, pod := range pods.Items {
			Lock.RLock()
			_, ok := PodMap[pod.GetName()]
			Lock.RUnlock()

			if !ok && pod.Status.Phase == apiv1.PodRunning {
				ctx, cancel := context.WithCancel(context.Background())
				go func(pod apiv1.Pod, ctx context.Context) {
					podName := pod.GetName()

					for _, container := range pod.Spec.Containers {
						go func(cName string, podName string) {
							reader, err := podClient.GetLogs(podName, &apiv1.PodLogOptions{Container: cName, Follow: true, Timestamps: true, SinceTime: &logTime}).Stream(ctx)
							if err != nil {
								cancel()
								deletePod(podName)
								fmt.Fprintf(os.Stderr, "%s/%s yielded error trying to get logs: %v\n", podName, cName, err)
								return
							}

							bufReader := bufio.NewReader(reader)
							for {
								line, err := bufReader.ReadString('\n')
								if strings.TrimSpace(line) != "" {
									fmt.Printf("[%s/%s]: %s", podName, cName, line)
								}

								if err != nil {
									break
								}
							}
						}(container.Name, podName)
					}

				}(pod, ctx)

				Lock.Lock()
				PodMap[pod.GetName()] = cancel
				Lock.Unlock()
			}
		}

		if *wait != 0 {
			if start.Add(*wait).Before(time.Now()) {
				// Probably pointless, but it's probably better to be nice
				Lock.Lock()
				for _, cancel := range PodMap {
					cancel()
				}
				Lock.Unlock()
				break
			}
		}

		<-tick
	}
}

func deletePod(name string) {
	Lock.Lock()
	delete(PodMap, name)
	Lock.Unlock()
}
