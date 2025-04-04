package webapp

import (
	"context"
	v1 "crds/pkg/apis/webapp/v1"
	clientset "crds/pkg/generated/clientset/versioned"
	webappv1 "crds/pkg/generated/informers/externalversions/webapp/v1"
	"fmt"
	"golang.org/x/time/rate"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	infomersv1 "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

const (
	controllerAgentName = "webapps"

	// FileBeat Consumer Producer Are Deployment Components
	FileBeat = "filebeat"
	Consumer = "consumer"
	Producer = "producer"

	// LogVolume FileBeatVolume are Volumes Components
	LogVolume      = "logVolumeComponent"
	FileBeatVolume = "filebeatVolumeComponent"

	HostPathType = coreV1.HostPathDirectoryOrCreate

	FilebeatConfigMapName = "webapp-filebeat-yaml-config"
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

var (
	webAppComponents = []string{FileBeat, Consumer, Producer}
	hostPathType     = HostPathType
	// Volume
	logVolume = coreV1.Volume{
		Name: "log",
		VolumeSource: coreV1.VolumeSource{
			HostPath: &coreV1.HostPathVolumeSource{
				Path: "/mnt2/crds/log",
				Type: &hostPathType,
			},
		},
	}

	filebeatVolume = coreV1.Volume{
		Name: FilebeatConfigMapName,
		VolumeSource: coreV1.VolumeSource{
			ConfigMap: &coreV1.ConfigMapVolumeSource{
				LocalObjectReference: coreV1.LocalObjectReference{
					Name: FilebeatConfigMapName,
				},
			},
		},
	}

	// VolumeMounts
	logVolumeMounts = coreV1.VolumeMount{
		Name:      "log",
		MountPath: "/home/web/webapp/logs",
	}
	filebeatVolumeMounts = coreV1.VolumeMount{
		Name:      FilebeatConfigMapName,
		MountPath: "/usr/share/filebeat/filebeat.yml",
		SubPath:   "filebeat.yml",
	}
)

type Controller struct {
	kubeInterface      kubernetes.Interface
	crdInterface       clientset.Interface
	deploymentInformer infomersv1.DeploymentInformer
	webappInformer     webappv1.WebappInformer
	workqueue          workqueue.TypedRateLimitingInterface[cache.ObjectName]
}

func NewController(ctx context.Context,
	kubeInterface kubernetes.Interface,
	crdInterface clientset.Interface,
	deploymentInformer infomersv1.DeploymentInformer,
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
	defer c.workqueue.Done(objRef)

	err := c.syncHandler(ctx, objRef)
	if err != nil {
		logger.Error(err, "Error syncing object")
	} else {
		c.workqueue.Forget(objRef)
		return true
	}

	utilruntime.HandleError(err)
	c.workqueue.AddRateLimited(objRef)

	return true
}

func (c Controller) syncHandler(ctx context.Context, objectRef cache.ObjectName) error {
	// objectRef usually is namespace/objectName
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "objectRef", objectRef)

	webapp, err := c.webappInformer.Lister().Webapps(objectRef.Namespace).Get(objectRef.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleErrorWithContext(ctx, err, "webapp referenced by item in work queue no longer exists", "objectReference", objectRef)
			return nil
		}
		return err
	}
	logger.V(4).Info("Got webapp")
	logger.V(5).Info("Got webapp. And its details:", "webapp", webapp)

	// Pre Check
	//// Filebeat config
	_, err = c.kubeInterface.CoreV1().ConfigMaps(webapp.Namespace).Get(ctx, FilebeatConfigMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		logger.V(4).Info("Can not find filebeat cm. Creating it.")
		fileContent, err := ioutil.ReadFile("artifacts/examples/webapp/config/filebeat.yml")
		if err != nil {
			return err

		}
		cm := &coreV1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      FilebeatConfigMapName,
				Namespace: webapp.Namespace,
			},
			Data: map[string]string{
				"filebeat.yml": string(fileContent),
			},
		}
		_, err = c.kubeInterface.CoreV1().ConfigMaps(webapp.Namespace).Create(ctx, cm, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	for _, webAppComponent := range webAppComponents {
		logger.V(4).Info(fmt.Sprintf("Processing %s Component", webAppComponent))
		deploymentName := fmt.Sprintf("webapp-%s-%s", webapp.Spec.Env, webAppComponent)
		replicas := int32(1)
		switch webAppComponent {
		case Consumer:
			replicas = *webapp.Spec.ConsumerReplicas
		case Producer:
			replicas = *webapp.Spec.ProducerReplicas
		}
		err = c.checkDeployment(ctx, webapp, deploymentName, webAppComponent, replicas)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Controller) checkDeployment(ctx context.Context, webapp *v1.Webapp, deploymentName, webAppComponent string, replicas int32) error {
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "Component", webAppComponent)
	deployment, err := c.deploymentInformer.Lister().Deployments(webapp.Namespace).Get(deploymentName)
	if errors.IsNotFound(err) {
		logger.Info(fmt.Sprintf("Can not find the %s Deployment. Creating", deploymentName))
		deployment, err = c.kubeInterface.AppsV1().Deployments(webapp.Namespace).Create(ctx, newWebComponentDeployment(ctx, webapp, webAppComponent, deploymentName, replicas), metav1.CreateOptions{FieldManager: FieldManager})
	}

	if err != nil {
		logger.V(4).Info("Encounter error: Failed to create Deployment", "Error", err)
		return err
	}

	if !metav1.IsControlledBy(deployment, webapp) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		logger.V(4).Info("Encounter error: Deployment is not controlled by ", "msg", msg)
		return fmt.Errorf("%s", msg)
	}

	if *deployment.Spec.Replicas != replicas {
		logger.V(4).Info("Update deployment resource", "currentReplicas", *deployment.Spec.Replicas, "desiredReplicas", replicas)
		deployment, err = c.kubeInterface.AppsV1().Deployments(webapp.Namespace).Update(ctx, newWebComponentDeployment(ctx, webapp, webAppComponent, deploymentName, replicas), metav1.UpdateOptions{FieldManager: FieldManager})
	}
	if err != nil {
		return err
	}

	return nil

}

func (c Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	logger := klog.FromContext(context.TODO())

	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			// TODO: understand it
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleErrorWithContext(context.Background(), nil, "Error decoding object, invalid type", "type", fmt.Sprintf("%T", obj))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			// TODO: understand it
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleErrorWithContext(context.Background(), nil, "Error decoding object tombstone, invalid type", "type", fmt.Sprintf("%T", tombstone.Obj))
			return
		}
	}
	logger.V(4).Info("Received deployment obj", "objName", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Webapp" {
			logger.V(4).Info("Skip processing deployment due to the wrong owner ref", "Deployment name", object.GetName(), "namespace", object.GetNamespace())
			return
		}
		//logger.Info("Processing Deployment", "Deployment name", object.GetName())
		webapp, err := c.webappInformer.Lister().Webapps(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			logger.V(4).Info("Ignore orphaned object", "object", klog.KObj(object), "foo", ownerRef.Name, "error", err)
			return
		}
		c.enqueue(webapp)
		logger.V(4).Info("Enqueued webapp")
		return
	}
	logger.V(6).Info("Skipped processing deployment due to no owner ref", "Deployment name", object.GetName(), "namespace", object.GetNamespace())

}

func (c Controller) enqueue(obj interface{}) {
	if objRef, err := cache.ObjectToName(obj); err != nil {
		utilruntime.HandleError(err)
	} else {
		c.workqueue.Add(objRef)
	}
}

// TODO: 277
//   - Filebeat config ---- Done
//   - Args of consumer and producer ---- Done
//   - Volumes? maybe hostPath ---- Done
//   - make different to Consumer and Producer. Currently they have not much difference
func newWebComponentDeployment(ctx context.Context, webapp *v1.Webapp, component, deploymentName string, replicas int32) *appsv1.Deployment {
	logger := klog.FromContext(ctx)
	logger.V(4).Info(fmt.Sprintf("Construct %s deployment for %s", component, webapp.Name))

	// image, hostPathType, volumes, volumeMount, args

	image := "docker.elastic.co/beats/filebeat:8.11.3"
	//hostPathType := coreV1.HostPathDirectoryOrCreate // defined in const

	var volumes []coreV1.Volume

	var volumeMount []coreV1.VolumeMount
	//var containers []coreV1.Container

	var args []string // args == nil => true

	switch component {
	case Consumer:
		image = fmt.Sprintf("192.168.38.89:30003/webapp/%s:%s", webapp.Spec.Branch, webapp.Spec.Version)
		volumes, volumeMount = getVolumesAndVolumeMounts(LogVolume)
		elm := fmt.Sprintf("--spring.profiles.active=%s,%s", webapp.Spec.Env, component)
		args = append(args, elm)
	case Producer:
		image = fmt.Sprintf("192.168.38.89:30003/webapp/%s:%s", webapp.Spec.Branch, webapp.Spec.Version)
		volumes, volumeMount = getVolumesAndVolumeMounts(LogVolume)
		elm := fmt.Sprintf("--spring.profiles.active=%s,%s", webapp.Spec.Env, component)
		args = append(args, elm)
	default:
		// default is filebeat
		volumes, volumeMount = getVolumesAndVolumeMounts(LogVolume, FileBeatVolume)
	}

	containers := getContainers(component, image, volumeMount, args)

	labels := map[string]string{
		"component":  component,
		"app":        deploymentName,
		"controller": webapp.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: webapp.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(webapp, v1.SchemeGroupVersion.WithKind("Webapp")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: coreV1.PodSpec{
					Containers: containers,
					Volumes:    volumes,
					NodeName:   "k8s-2",
				},
			},
		},
	}

}

func getVolumesAndVolumeMounts(volumesComponents ...string) ([]coreV1.Volume, []coreV1.VolumeMount) {
	// Combine volume and volumeMounts from vars

	var volumes []coreV1.Volume

	var volumeMount []coreV1.VolumeMount

	for _, component := range volumesComponents {
		switch component {
		case FileBeatVolume:
			volumes = append(volumes, filebeatVolume)
			volumeMount = append(volumeMount, filebeatVolumeMounts)
		case LogVolume:
			volumes = append(volumes, logVolume)
			volumeMount = append(volumeMount, logVolumeMounts)
		}
	}

	return volumes, volumeMount
}

func getContainers(component, image string, vm []coreV1.VolumeMount, args []string) []coreV1.Container {

	containers := []coreV1.Container{
		{
			Name:         component,
			Image:        image,
			VolumeMounts: vm,
			Env: []coreV1.EnvVar{
				{
					Name: "MY_POD_NAME",
					ValueFrom: &coreV1.EnvVarSource{
						FieldRef: &coreV1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
			},
			//ImagePullPolicy: "Never",
		},
	}
	if args != nil {
		containers[0].Args = args
	}

	return containers
}
