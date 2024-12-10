package webapp

import (
	"context"
	clientset "crds/pkg/generated/clientset/versioned"
	webappv1 "crds/pkg/generated/informers/externalversions/webapp/v1"
	"fmt"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/util/wait"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

type Controller struct {
	kubeInterface      kubernetes.Interface
	crdInterface       clientset.Interface
	deploymentInformer appsv1.DeploymentInformer
	webappInformer     webappv1.WebappInformer
	workqueue          workqueue.TypedRateLimitingInterface[cache.ObjectName]
}

func NewController(ctx context.Context,
	kubeInterface kubernetes.Interface,
	crdInterface clientset.Interface,
	deploymentInformer appsv1.DeploymentInformer,
	webappInformer webappv1.WebappInformer) *Controller {

	logger := klog.FromContext(ctx)
	logger.Info("Initializing controller")

	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[cache.ObjectName](5*time.Millisecond, 1000*time.Second),
		&workqueue.TypedBucketRateLimiter[cache.ObjectName]{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	// init
	controller := &Controller{
		kubeInterface:      kubeInterface,
		crdInterface:       crdInterface,
		deploymentInformer: deploymentInformer,
		webappInformer:     webappInformer,
		workqueue:          workqueue.NewTypedRateLimitingQueue(ratelimiter),
	}

	// event handler func
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	webappInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueue(new)
		},
		DeleteFunc: controller.enqueue,
	})

	return controller
}

func (c Controller) Run(ctx context.Context, i int) error {

	logger := klog.FromContext(ctx)

	logger.Info("Waiting for caches to be synced")
	//wait for cache to be synced
	if ok := cache.WaitForCacheSync(ctx.Done(), c.deploymentInformer.Informer().HasSynced, c.webappInformer.Informer().HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	logger.Info("Caches synced")

	for j := 0; j < i; j++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Controller started")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

func (c Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

// TODO processNextItem
func (c Controller) processNextItem(ctx context.Context) bool {
	logger := klog.FromContext(ctx)

	// Get blocks until it can return an item to be processed. If shutdown = true,
	// the caller should end their goroutine. You must call Done with item when you
	// have finished processing it.
	objRef, shutdown := c.workqueue.Get() // This will block if can not get item from work queue
	if shutdown {
		logger.Info("Worker shutting down")
		return false
	}
	logger.Info("Processing item", "Name of ObjRef", objRef.Name)
	defer c.workqueue.Done(objRef)

	return true
}

// TODO: handleObject
func (c Controller) handleObject(obj interface{}) {
	logger := klog.FromContext(context.TODO())
	logger.Info("Handling object")

}

// TODO: enqueue
func (c Controller) enqueue(obj interface{}) {
	logger := klog.FromContext(context.TODO())
	logger.Info("Enqueuing object")
}
