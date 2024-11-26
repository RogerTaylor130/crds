package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	KubeConfigPath = "D:\\projects\\goprojects\\cobratest\\k8stests\\configs\\config"
)

func GetK8SClientSet() *kubernetes.Clientset {
	config, _ := clientcmd.BuildConfigFromFlags("", KubeConfigPath)
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientSet
}
