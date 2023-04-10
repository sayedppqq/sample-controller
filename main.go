package main

import (
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/code-generator"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/appscodepc/.kube/config", "my kubeconfig file location")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {

		// in cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error while building config file %s", err.Error())
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error %s while building clientset", err.Error())
	}

	fmt.Println(clientset)
}
