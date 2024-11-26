package bar

import (
	"context"
	clientset "crds/pkg/generated/clientset/versioned"
	"golang.org/x/time/rate"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

type BarController struct {
	kubeInterface kubernetes.Interface
	BarInterface  clientset.Interface
	workqueue     workqueue.TypedRateLimitingInterface[cache.ObjectName]
}

func NewBarController(
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

	return controller

}

func (c BarController) Run(ctx context.Context) {

}
