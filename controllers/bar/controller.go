package bar

import (
	"context"
	clientset "crds/pkg/generated/clientset/versioned"
	informers "crds/pkg/generated/informers/externalversions"
	"crds/pkg/generated/informers/externalversions/mycrds/v1alpha1"
	"golang.org/x/time/rate"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"log"
	"time"
)

type BarController struct {
	kubeInterface      kubernetes.Interface
	BarInterface       clientset.Interface
	barInformer        v1alpha1.BarInformer
	deploymentInformer v1.DeploymentInformer
	workqueue          workqueue.TypedRateLimitingInterface[cache.ObjectName]
}

func NewBarController(
	ctx context.Context,
	kubeInterface kubernetes.Interface,
	barInterface clientset.Interface,
	barInformer v1alpha1.BarInformer,
	deploymentInformer v1.DeploymentInformer) *BarController {

	logger := klog.FromContext(ctx)

	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[cache.ObjectName](5*time.Millisecond, 1000*time.Second),
		&workqueue.TypedBucketRateLimiter[cache.ObjectName]{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &BarController{
		kubeInterface: kubeInterface,
		BarInterface:  barInterface,
		workqueue:     workqueue.NewTypedRateLimitingQueue(ratelimiter),
	}

	officalFactory := kubeinformers.NewSharedInformerFactory(kubeInterface, time.Second*20)
	barFactory := informers.NewSharedInformerFactory(barInterface, 0)

	//barsInformer := barFactory.Roger().V1alpha1().Bars()

	//podInformer := officalFactory.Core().V1().Pods()
	//
	//podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc: func(obj interface{}) {
	//		//log.Println("Add pod")
	//	},
	//})

	barInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.enqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.enqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			logger.Info("Bar Crd deleted: ", "Object:", obj)
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

	return controller

}

func (c BarController) Run(ctx context.Context, worker int) {

	log.Println("Waiting for pod&deployment cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.deploymentInformer.Informer().HasSynced); !ok {
		log.Fatal("Failed to wait for pod or deploymentcache")
	}
	log.Println("Pod & deployment cache synced")

	log.Println("Waiting for example cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.barInformer.Informer().HasSynced); !ok {
		log.Fatal("Failed to sync")
	}
	log.Println("Cache synced")

	for i := 0; i < worker; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}
}

func (c BarController) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c BarController) processNextWorkItem(ctx context.Context) bool {
	logger := klog.FromContext(ctx)

	return true
}

func (c BarController) enqueue(obj interface{}) {
	if objRef, err := cache.ObjectToName(obj); err != nil {
		utilruntime.HandleError(err)
	} else {
		c.workqueue.Add(objRef)
	}
}
