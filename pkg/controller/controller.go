package controller

import (
	"context"
	"fmt"
	controllerv1alpha1 "github.com/sayedppqq/sample-controller/pkg/apis/sayedppqq.dev/v1alpha1"
	klient "github.com/sayedppqq/sample-controller/pkg/client/clientset/versioned"
	informer "github.com/sayedppqq/sample-controller/pkg/client/informers/externalversions/sayedppqq.dev/v1alpha1"
	lister "github.com/sayedppqq/sample-controller/pkg/client/listers/sayedppqq.dev/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

// Controller Structure for Kluster controller
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// klusterclientset is a clientset for our own API group
	klusterclientset klient.Interface

	// For deployment resource
	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced

	// For Kluster resource
	KlusterLister lister.KlusterLister
	KlusterSynced cache.InformerSynced

	// work queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens.
	workQueue workqueue.RateLimitingInterface
}

// NewController Constructor for Kluster and return a new controller
func NewController(
	kubeclientset kubernetes.Interface,
	klusterclientset klient.Interface,
	deploymentInformer appsinformer.DeploymentInformer,
	klusterInformer informer.KlusterInformer) *Controller {

	controller := &Controller{
		kubeclientset:    kubeclientset,
		klusterclientset: klusterclientset,

		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,

		KlusterLister: klusterInformer.Lister(),
		KlusterSynced: klusterInformer.Informer().HasSynced,

		workQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Klusters"),
	}
	log.Println("setting up event handler")

	//set up event handler when Kluster resource changes
	klusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueKluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			controller.enqueueKluster(newObj)
		},
		DeleteFunc: controller.enqueueKluster,
	})
	return controller
}

// enqueueKluster takes a Kluster resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Kluster and only can call by a
// Controller object.
func (c *Controller) enqueueKluster(obj interface{}) {
	log.Println("enqueueing Kluster in work queue....")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	c.workQueue.AddRateLimited(key)
}

func (c *Controller) Run(thread int, ch <-chan struct{}) error {

	defer runtime.HandleCrash()
	defer c.workQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Println("Starting Kluster controller....")

	// Wait for the caches to be synced before starting workers
	log.Println("Waiting for the informer caches to sync....")
	if ok := cache.WaitForCacheSync(ch, c.deploymentsSynced, c.KlusterSynced); !ok {
		return fmt.Errorf("error while sync in cache")
	}

	log.Println("Starting workers")
	// Launch two workers to process Kluster resources
	for i := 0; i < thread; i++ {
		// Until loops until ch channel is closed, running runWorker every time period.
		go wait.Until(c.runWorker, time.Second, ch)
	}

	// Here channel holds the execution of this function so that
	// wat.Until can call continuously after time period
	<-ch

	log.Println("Shutting down worker")
	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {

	}
}
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workQueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func, so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off period.
		defer c.workQueue.Done(obj)
		var key string
		var ok bool

		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workQueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Kluster resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		// Finally, if no error occurs we Forget this item, so it does not
		// get queued again until another change happens.
		c.workQueue.Forget(obj)
		log.Printf("successfully synced %s\n", key)
		return nil
	}(obj)
	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Kluster resource
// with the current status of the resource.
// implement the business logic here.
func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	log.Println("syncHandler called")
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Kluster resource with this namespace/name
	kluster, err := c.KlusterLister.Klusters(namespace).Get(name)
	if err != nil {
		// The Kluster resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("Kluster '%s' in work queue no longer exists", key))
			return nil
		}
		return nil
	}

	deploymentName := kluster.Spec.Name
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s : deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Kluster.spec
	deployment, err := c.deploymentsLister.Deployments(namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Create(context.TODO(), newDeployment(kluster), metav1.CreateOptions{})
	}

	// If an error occurs during Get/Create, we'll requeue the item, so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If this number of the replicas on the Kluster resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if kluster.Spec.Replicas != nil && *kluster.Spec.Replicas != *deployment.Spec.Replicas {
		log.Printf("%s replicas: %d but Deployment replicas: %d\n", name, *kluster.Spec.Replicas, *deployment.Spec.Replicas)

		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Update(context.TODO(), newDeployment(kluster), metav1.UpdateOptions{})

		// If an error occurs during Update, we'll requeue the item, so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}
	}

	// Finally, we update the status block of the Kluster resource to reflect the
	// current state of the world
	err = c.updateKlusterStatus(kluster, deployment)

	if err != nil {
		return err
	}
	return nil
}

// newDeployment creates a new Deployment for a Kluster resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Kluster resource that 'owns' it.
func newDeployment(kluster *controllerv1alpha1.Kluster) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kluster.Spec.Name,
			Namespace: kluster.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: kluster.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-app",
					},
				},
				Spec: v1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-app",
							Image: kluster.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: kluster.Spec.Container.Port,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (c *Controller) updateKlusterStatus(kluster *controllerv1alpha1.Kluster, deployment *appsv1.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	klusterCopy := kluster.DeepCopy()
	klusterCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Kluster resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.klusterclientset.SayedppqqV1alpha1().Klusters(kluster.Namespace).Update(context.TODO(), klusterCopy, metav1.UpdateOptions{})

	return err
}
