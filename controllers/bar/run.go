package bar

import (
	"context"
	informers "crds/pkg/generated/informers/externalversions"
	clientTools "crds/tools/client"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/klog/v2"
)

func RunBarController(ctx context.Context) {
	officalClient := clientTools.GetOfficialClientSet()

	logger := klog.FromContext(ctx)

	barClient := clientTools.GetExampleClientSet()

	officalFactory := kubeinformers.NewSharedInformerFactory(officalClient, 0)
	barFactory := informers.NewSharedInformerFactory(barClient, 0)

	barInformer := barFactory.Example().V1alpha1().Bars()
	deploymentInformer := officalFactory.Apps().V1().Deployments()

	controller := NewBarController(ctx, officalClient, barClient, barInformer, deploymentInformer)

	logger.Info("Starting Informers")
	officalFactory.Start(ctx.Done())
	barFactory.Start(ctx.Done())

	if err := controller.Run(ctx, 1); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
