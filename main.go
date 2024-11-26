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

	barClient := clientTools.GetExampleClientSet()

	officalFactory := kubeinformers.NewSharedInformerFactory(officalClient, time.Second*20)
	barFactory := informers.NewSharedInformerFactory(barClient, 0)

	barsInformer := barFactory.Roger().V1alpha1().Bars()
	deploymentInformer := officalFactory.Apps().V1().Deployments()
	podInformer := officalFactory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Add pod")
		},
	})

	barsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Add mycrds:", obj)
		},
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Add deployment:", obj)
		},
	})

	log.Println("Starting Informers")
	officalFactory.Start(ctx.Done())
	barFactory.Start(ctx.Done())

	log.Println("Waiting for pod&deployment cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), podInformer.Informer().HasSynced, deploymentInformer.Informer().HasSynced); !ok {
		log.Fatal("Failed to wait for pod or deploymentcache")
	}
	log.Println("Pod & deployment cache synced")

	log.Println("Waiting for example cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), barsInformer.Informer().HasSynced); !ok {
		log.Fatal("Failed to sync")
	}
	log.Println("Cache synced")

	<-time.After(time.Minute * 5)

}
