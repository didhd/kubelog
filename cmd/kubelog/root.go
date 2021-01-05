package main

import (
	"github.com/spf13/cobra"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	context    string
	namespace  string
	container  string
	kubeconfig string
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "kubelog",
	Short: "Stream the logs for deployment, daemonset, statefulset or a container in a pod.",
	Long: `Stream the logs for deployment, daemonset, statefulset or a container in a pod.`,
}
