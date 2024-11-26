package client

import (
	"context"
	clientset "crds/pkg/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/signal"
)

var (
	KubeConfigPath = "D:\\projects\\goprojects\\cobratest\\k8stests\\configs\\config"
)

func getConfig(masterUrl, kConfig string) (*restclient.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kConfig)
	return config, err
}

func GetExampleClientSet() *clientset.Clientset {
	config, _ := getConfig("", KubeConfigPath)
	client, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return client

}

func GetOfficialClientSet() *kubernetes.Clientset {
	config, _ := getConfig("", KubeConfigPath)
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return client
}

var onlyOneSignalHandler = make(chan struct{})
var shutdownSignals = []os.Signal{os.Interrupt}

// SetupSignalHandler registered for SIGTERM and SIGINT. A context is returned
// which is cancelled on one of these signals. If a second signal is caught,
// the program is terminated with exit code 1.
func SetupSignalHandler() context.Context {
	close(onlyOneSignalHandler) // panics when called twice

	c := make(chan os.Signal, 2)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
