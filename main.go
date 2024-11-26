package main

import (
	informers "crds/pkg/generated/informers/externalversions"
	clientTools "crds/tools/client"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"time"
)

func main() {

	officalClient := clientTools.GetOfficialClientSet()

	ctx := clientTools.SetupSignalHandler()

	januaryClient := clientTools.GetExampleClientSet()

	officalFactory := kubeinformers.NewSharedInformerFactory(officalClient, time.Second*20)
	januaryFactory := informers.NewSharedInformerFactory(januaryClient, 0)

	januaryInformer := januaryFactory.Roger().V1alpha1().Januaries()
	deploymentInformer := officalFactory.Apps().V1().Deployments()
	podInformer := officalFactory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Add pod")
		},
	})

	januaryInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Add january:", obj)
		},
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Add deployment:", obj)
		},
	})

	log.Println("Starting Informers")
	officalFactory.Start(ctx.Done())
	januaryFactory.Start(ctx.Done())

	log.Println("Waiting for pod cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), podInformer.Informer().HasSynced); !ok {
		log.Fatal("Failed to wait for pod cache")
	}
	log.Println("Pod cache synced")

	log.Println("Waiting for example cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), januaryInformer.Informer().HasSynced); !ok {
		log.Fatal("Failed to sync")
	}
	log.Println("Cache synced")

}
