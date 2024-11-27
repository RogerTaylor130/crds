package main

import (
	"crds/controllers/bar"
	informers "crds/pkg/generated/informers/externalversions"
	clientTools "crds/tools/client"
	kubeinformers "k8s.io/client-go/informers"
	"log"
	"time"
)

func main() {

	officalClient := clientTools.GetOfficialClientSet()

	ctx := clientTools.SetupSignalHandler()

	barClient := clientTools.GetExampleClientSet()

	officalFactory := kubeinformers.NewSharedInformerFactory(officalClient, time.Second*20)
	barFactory := informers.NewSharedInformerFactory(barClient, 0)

	barInformer := barFactory.Roger().V1alpha1().Bars()
	deploymentInformer := officalFactory.Apps().V1().Deployments()

	log.Println("Starting Informers")
	officalFactory.Start(ctx.Done())
	barFactory.Start(ctx.Done())

	controller := bar.NewBarController(ctx, officalClient, barClient, barInformer, deploymentInformer)

	controller.Run(ctx, 1)
}
