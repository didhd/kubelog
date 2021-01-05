package main

import (
	"bufio"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	k8s "github.com/didhd/kubelog/pkg/kubernetes"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/homedir"
)

func init() {
	rootCmd.AddCommand(logsCmd)

	// Make sure the order of the flags is correct.
	logsCmd.Flags().SortFlags = false
	logsCmd.PersistentFlags().SortFlags = false

	logsCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "(optional) kubernetes namespace")
	logsCmd.Flags().StringVarP(&container, "container", "c", "", "(optional) kubernetes container name")

	// Load kubeconfig filepath.
	if home := homedir.HomeDir(); home != "" {
		logsCmd.Flags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		logsCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream the logs for deployment, daemonset, statefulset or a container in a pod.",
	Long: `Stream the logs for deployment, daemonset, statefulset or a container in a pod. If the pods have only one container, the container name is
optional.

Examples:
	# Stream logs from pod nginx with only one container
	kubelog logs nginx -n default

	# Stream all pod logs in a deployment, daemonset or statefulset.
	kubelog logs deployment/nginx -n default
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get Kubernetes client
		clientset, err := k8s.NewClientset(kubeconfig)
		if err != nil {
			log.Fatalf("cannot find kubeconfig")
		}

		// Parse args into podID
		if len(args) <= 0 {
			log.Fatalf("A pod, deployment, daemonset or statefulset should be specified")
		}

		if strings.Contains(args[0], "/") {
			// Deployment, daemonset or statefulset

			if len(strings.Split(args[0], "/")) <= 1 {
				log.Fatalf("cannot stream logs")
			}

			kind := strings.ToLower(strings.Split(args[0], "/")[0])
			name := strings.Split(args[0], "/")[1]
			var selector *metav1.LabelSelector

			if kind == "deployment" || kind == "deployments" {
				deployment, err := clientset.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
				if err != nil {
					log.Fatalf("cannot stream logs. %v", err)
				}
				selector = deployment.Spec.Selector
			} else if kind == "daemonset" || kind == "daemonsets" {
				daemonset, err := clientset.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
				if err != nil {
					log.Fatalf("cannot stream logs. %v", err)
				}
				selector = daemonset.Spec.Selector
			} else if kind == "statefulset" || kind == "statefulsets" {
				statefulset, err := clientset.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
				if err != nil {
					log.Fatalf("cannot stream logs. %v", err)
				}
				selector = statefulset.Spec.Selector
			}

			// Get LabelSelector.
			labelMap, err := metav1.LabelSelectorAsMap(selector)
			if err != nil {
				log.Fatalf("cannot stream logs. %v", err)
			}
			options := metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(labelMap).String(),
			}

			// Search pods using LabelSelector.
			pods, err := clientset.CoreV1().Pods(namespace).List(options)
			if err != nil {
				log.Fatalf("cannot stream logs. %v", err)
			}

			// For colors
			colorReset := "\033[0m"
			colorRed := "\033[31m"

			var wg sync.WaitGroup
			wg.Add(len(pods.Items))

			for _, pod := range pods.Items {
				go func(pod corev1.Pod) {
					podID := pod.Name

					if len(container) == 0 {
						if len(args) == 1 {
							if len(pod.Spec.Containers) != 1 {
								podContainersNames := []string{}
								for _, container := range pod.Spec.Containers {
									podContainersNames = append(podContainersNames, container.Name)
								}

								log.Fatalf("Pod %s has the following containers: %s; please specify the container to print logs for with -c", pod.ObjectMeta.Name, strings.Join(podContainersNames, ", "))
							}
							container = pod.Spec.Containers[0].Name
						} else {
							container = args[1]
						}
					}

					req := clientset.CoreV1().RESTClient().Get().
						Namespace(namespace).
						Name(podID).
						Resource("pods").
						SubResource("log").
						Param("follow", strconv.FormatBool(true)).
						Param("container", container).
						Param("timestamps", strconv.FormatBool(false))

					readCloser, err := req.Stream()
					if err != nil {
						log.Fatalf("cannot stream logs. %v", err)
					}

					defer readCloser.Close()
					defer wg.Done()

					rd := bufio.NewReader(readCloser)
					for {
						str, err := rd.ReadString('\n')
						if err != nil {
							log.Fatal("Read Error:", err)
							return
						}
						fmt.Print(string(colorRed), "["+podID+"] ", string(colorReset), str)
					}

					// _, err = io.Copy(os.Stdout, readCloser)
				}(pod)
			}

			wg.Wait()

		} else {
			// Single Pod
			podID := args[0]
			req := clientset.CoreV1().RESTClient().Get().
				Namespace(namespace).
				Name(podID).
				Resource("pods").
				SubResource("log").
				Param("follow", strconv.FormatBool(true)).
				Param("timestamps", strconv.FormatBool(false))

			readCloser, err := req.Stream()
			if err != nil {
				log.Fatalf("cannot stream logs. %v", err)
			}

			defer readCloser.Close()
			rd := bufio.NewReader(readCloser)
			for {
				str, err := rd.ReadString('\n')
				if err != nil {
					log.Fatal("Read Error:", err)
					return
				}
				fmt.Print(str)
			}
		}

	},
}
