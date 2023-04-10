package main

import (
	"flag"
	"fmt"
	"github.com/RokibulHasan7/crd/controller"
	clientset "github.com/RokibulHasan7/crd/pkg/client/clientset/versioned"
	informers "github.com/RokibulHasan7/crd/pkg/client/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

func main() {
	log.Println("Configuring kubeConfig...")
	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Println("Building config from flags, %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Building kubernetes clientset, %s", err.Error())
		panic(err)
	}
	fmt.Println(kubeClient)

	exampleClient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Println("Building example clientset, %s", err.Error())
		panic(err)
	}

	kubeInformationFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Minute)
	exampleInformationFactory := informers.NewSharedInformerFactory(exampleClient, time.Minute)

	ctrl := controller.NewController(kubeClient, exampleClient,
		kubeInformationFactory.Apps().V1().Deployments(),
		exampleInformationFactory.Rokibul().V1alpha1().RokibulHasans())

	ch := make(chan struct{})
	exampleInformationFactory.Start(ch)
	kubeInformationFactory.Start(ch)

	if err := ctrl.Run(ch); err != nil {
		log.Printf("Error runnning controller %s \n", err.Error())

	}
}
