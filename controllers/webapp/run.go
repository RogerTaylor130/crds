package webapp

import (
	"context"
	crdinformers "crds/pkg/generated/informers/externalversions"
	clientTools "crds/tools/client"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"time"
)

func RunWebappController(ctx context.Context) {
	logger := klog.FromContext(ctx)

	crdClient := clientTools.GetExampleClientSet()
	officialClient := clientTools.GetOfficialClientSet()

	officialInformerFactory := kubeinformers.NewSharedInformerFactory(officialClient, time.Second*30)
	crdInformerFactory := crdinformers.NewSharedInformerFactory(crdClient, time.Second*30)

	contrller := NewController(ctx, officialClient, crdClient,
		officialInformerFactory.Apps().V1().Deployments(),
		crdInformerFactory.Custom().V1().Webapps())

	logger.Info("Controller initialized. Starting informers & controller")

	officialInformerFactory.Start(ctx.Done())
	crdInformerFactory.Start(ctx.Done())

	if err := contrller.Run(ctx, 1); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}
