package bar

import (
	"context"
	clientset "crds/pkg/generated/clientset/versioned"
	informers "crds/pkg/generated/informers/externalversions"
	"golang.org/x/time/rate"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

type BarController struct {
	kubeInterface kubernetes.Interface
	BarInterface  clientset.Interface
	workqueue     workqueue.TypedRateLimitingInterface[cache.ObjectName]
}

func NewBarController(
	ctx context.Context,
	kubeInterface kubernetes.Interface,
	BarInterface clientset.Interface) *BarController {

	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[cache.ObjectName](5*time.Millisecond, 1000*time.Second),
		&workqueue.TypedBucketRateLimiter[cache.ObjectName]{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &BarController{
		kubeInterface: kubeInterface,
		BarInterface:  BarInterface,
		workqueue:     workqueue.NewTypedRateLimitingQueue(ratelimiter),
	}

	officalFactory := kubeinformers.NewSharedInformerFactory(kubeInterface, time.Second*20)
	barFactory := informers.NewSharedInformerFactory(BarInterface, 0)

	barsInformer := barFactory.Roger().V1alpha1().Bars()
	deploymentInformer := officalFactory.Apps().V1().Deployments()
	podInformer := officalFactory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//log.Println("Add pod")
		},
	})

	barsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Println("Add mycrds:", obj)
		},
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//log.Println("Add deployment:")
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

	return controller

}

func (c BarController) Run(ctx context.Context) {

	<-ctx.Done()
}
