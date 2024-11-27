package main

import (
	"crds/controllers/bar"
	informers "crds/pkg/generated/informers/externalversions"
	clientTools "crds/tools/client"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"log"
	"time"
)

func main() {

	officalClient := clientTools.GetOfficialClientSet()

	ctx := clientTools.SetupSignalHandler()
	logger := klog.FromContext(ctx)

	barClient := clientTools.GetExampleClientSet()

	officalFactory := kubeinformers.NewSharedInformerFactory(officalClient, time.Second*20)
	barFactory := informers.NewSharedInformerFactory(barClient, 0)

	barInformer := barFactory.Roger().V1alpha1().Bars()
	deploymentInformer := officalFactory.Apps().V1().Deployments()

	controller := bar.NewBarController(ctx, officalClient, barClient, barInformer, deploymentInformer)

	log.Println("Starting Informers")
	officalFactory.Start(ctx.Done())
	barFactory.Start(ctx.Done())

	if err := controller.Run(ctx, 1); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
