package bar

import (
	"context"
	barv1alpha1 "crds/pkg/apis/mycrds/v1alpha1"
	clientset "crds/pkg/generated/clientset/versioned"
	"crds/pkg/generated/informers/externalversions/mycrds/v1alpha1"
	"fmt"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

const controllerAgentName = "crds"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Bar is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Bar fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Bar"
	// MessageResourceSynced is the message used for an Event fired when a Bar
	// is synced successfully
	MessageResourceSynced = "Bar synced successfully"
	// FieldManager distinguishes this controller from other things writing to API objects
	FieldManager = controllerAgentName
)

type BarController struct {
	kubeInterface      kubernetes.Interface
	barInterface       clientset.Interface
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
		kubeInterface:      kubeInterface,
		barInterface:       barInterface,
		barInformer:        barInformer,
		deploymentInformer: deploymentInformer,
		workqueue:          workqueue.NewTypedRateLimitingQueue(ratelimiter),
	}

	//barsInformer := barFactory.Roger().V1alpha1().Bars()

	//podInformer := officalFactory.Core().V1().Pods()
	//
	//podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc: func(obj interface{}) {
	//		//log.Println("Add pod")
	//	},
	//})

	barInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueue,
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

	return controller

}

func (c BarController) Run(ctx context.Context, worker int) error {

	logger := klog.FromContext(ctx)
	logger.Info("Waiting for deployment cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.deploymentInformer.Informer().HasSynced); !ok {
		return fmt.Errorf("Failed to wait for deployment to be cached")
	}
	logger.Info("Pod & deployment cache synced")

	logger.Info("Waiting for example cache sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.barInformer.Informer().HasSynced); !ok {
		return fmt.Errorf("Failed to sync")
	}
	logger.Info("Cache synced")

	for i := 0; i < worker; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

func (c BarController) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c BarController) processNextWorkItem(ctx context.Context) bool {
	logger := klog.FromContext(ctx)

	objRef, shutdown := c.workqueue.Get()
	if shutdown {
		logger.Info("Worker shutting down")
		return false
	}

	defer c.workqueue.Done(objRef)

	err := c.syncHandler(ctx, objRef)
	if err != nil {
		c.workqueue.Forget(objRef)
		logger.Info("Successfully synced", "objectName", objRef)
		return true
	}

	utilruntime.HandleError(err)
	c.workqueue.AddRateLimited(objRef)
	return true
}

func (c BarController) syncHandler(ctx context.Context, objectRef cache.ObjectName) error {
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "Object", objectRef.Name)

	bar, err := c.barInformer.Lister().Bars(objectRef.Namespace).Get(objectRef.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleErrorWithContext(ctx, err, "Bar referenced by item in work queue no longer exists", "objectReference", objectRef)
			return nil
		}

		return err
	}

	deploymentName := bar.Spec.DeploymentName
	if deploymentName == "" {
		utilruntime.HandleErrorWithContext(ctx, nil, "Deployment name missing from object reference", "objectReference", objectRef)
		return nil
	}

	deployment, err := c.deploymentInformer.Lister().Deployments(bar.Namespace).Get(deploymentName)
	if errors.IsNotFound(err) {
		logger.Info("Can not find the deployment, Creating it")
		//deployment, err = c.kubeInterface.AppsV1().Deployments(bar.Namespace).Create(ctx, newDeployment(bar), metav1.CreateOptions{FieldManager: FieldManager})
	}

	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(bar, deployment) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		//c.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf("%s", msg)
	}

	if bar.Spec.Replicas != nil && *bar.Spec.Replicas != *deployment.Spec.Replicas {
		logger.V(4).Info("Update deployment resource", "currentReplicas", *bar.Spec.Replicas, "desiredReplicas", *deployment.Spec.Replicas)
		//deployment, err = c.kubeInterface.AppsV1().Deployments(bar.Namespace).Update(ctx, newDeployment(bar), metav1.UpdateOptions{FieldManager: FieldManager})
	}

	if err != nil {
		return err
	}
	//c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func newDeployment(bar *barv1alpha1.Bar) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "nginx",
		"controller": bar.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bar.Spec.DeploymentName,
			Namespace: bar.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(bar, barv1alpha1.SchemeGroupVersion.WithKind("Bar")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: bar.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}

func (c BarController) enqueue(obj interface{}) {
	if objRef, err := cache.ObjectToName(obj); err != nil {
		utilruntime.HandleError(err)
	} else {
		c.workqueue.Add(objRef)
	}
}
