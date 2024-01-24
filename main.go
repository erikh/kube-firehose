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
	tail = flag.Bool("t", false, "new contents only; no history")
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
	logTime := metav1.NewTime(time.Unix(0, 0))

	if *tail {
		logTime = metav1.Now()
	}

	for {
		podClient := clientset.CoreV1().Pods(apiv1.NamespaceDefault)
		pods, err := podClient.List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		for _, pod := range pods.Items {
			Lock.RLock()
			if _, ok := PodMap[pod.GetName()]; !ok {
				Lock.RUnlock()
				ctx, cancel := context.WithCancel(context.Background())
				go func(pod apiv1.Pod, ctx context.Context) {
					podName := pod.GetName()

					if pod.Status.Phase != apiv1.PodRunning {
						cancel()
						deletePod(podName)
						return
					}

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
			} else {
				Lock.RUnlock()
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
