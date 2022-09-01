package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	kubernetesConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var dryRun bool
var podStatusReason []string

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newK8SClient() (kubernetes.Interface, error) {
	kubeConfig, err := kubernetesConfig.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(kubeConfig)
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", true, "Dry run mode")
	rootCmd.PersistentFlags().StringArrayVarP(&podStatusReason, "pod-status-reason", "s", []string{"Terminated"}, "List of pod.Status.Reason")

}

var rootCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove pods",
	Run: func(cmd *cobra.Command, args []string) {
		remove()
	},
}

func remove() {
	k8sClient, err := newK8SClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	namespaces, err := k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	for _, namespace := range namespaces.Items {
		fmt.Fprintf(os.Stdout, "Working in namespace: %v\n", namespace.Name)
		listPods, err := k8sClient.CoreV1().Pods(namespace.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		for _, pod := range listPods.Items {
			for _, status := range podStatusReason {
				if pod.Status.Reason == status {
					if dryRun {
						fmt.Fprintf(os.Stdout, "Pod: %v will be deleted.\n", pod.Name)

					} else {
						err := k8sClient.CoreV1().Pods(namespace.Name).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failed to delete pod: %v %v\n", pod.Name, err)
							os.Exit(1)
						} else {
							fmt.Fprintf(os.Stdout, "Pod: %v deleted.\n", pod.Name)
						}
					}
				}
			}
		}
	}
}
