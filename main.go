package main

import (
	"flag"
	"fmt"
	klient "github.com/sayedppqq/sample-controller/pkg/client/clientset/versioned"
	klusterinformers "github.com/sayedppqq/sample-controller/pkg/client/informers/externalversions"
	"github.com/sayedppqq/sample-controller/pkg/controller"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/code-generator"
	"time"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/home/appscodepc/.kube/config", "my kubeconfig file location")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {

		// in cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error while building config file %s\n", err.Error())
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error while building kubernetes clientset %s\n", err.Error())
	}
	klientset, err := klient.NewForConfig(config)
	if err != nil {
		fmt.Printf("error while building custom klientset %s\n", err.Error())
	}

	// Initialise the informer resource and here we will be using sharedinformer factory instead of simple informers
	// because in case if we need to query / watch multiple Group versions, and it’s a good practise as well
	// NewSharedInformerFactory will create a new ShareInformerFactory for "all namespaces"
	// 20*time.Second is the re-sync period to update the in-memory cache of informer //
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(clientset, time.Second*5)
	klusterInformerFactory := klusterinformers.NewSharedInformerFactory(klientset, time.Second*5)

	// From this informerfactory we can create specific informers for every group version resource
	// that are default available in k8s environment such as Pods, deployment, etc
	// podInformer := kubeInformationFactory.Core().V1().Pods()
	ctrl := controller.NewController(
		clientset,
		klientset,
		kubeInformerFactory.Apps().V1().Deployments(),
		klusterInformerFactory.Sayedppqq().V1alpha1().Klusters(),
	)

	// creating an unbuffered channel to synchronize the update
	ch := make(chan struct{})
	defer close(ch)

	// These will run until the channel is closed.
	kubeInformerFactory.Start(ch)
	klusterInformerFactory.Start(ch)

	if err = ctrl.Run(2, ch); err != nil {
		fmt.Printf("error while running controller\n")
	}
}
