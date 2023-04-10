package controller

import (
	"context"
	"fmt"

	rokibulv1alpha1 "github.com/RokibulHasan7/crd/pkg/apis/rokibul.hasan/v1alpha1"
	clientset "github.com/RokibulHasan7/crd/pkg/client/clientset/versioned"
	informer "github.com/RokibulHasan7/crd/pkg/client/informers/externalversions/rokibul.hasan/v1alpha1"
	lister "github.com/RokibulHasan7/crd/pkg/client/listers/rokibul.hasan/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	// client for custom resource RokibulHasan
	sampleclientset clientset.Interface

	// RokibulHasan has synced
	rokibulhasanSynced cache.InformerSynced

	// lister
	rokibulhasanLister lister.RokibulHasanLister

	// queue
	workQueue workqueue.RateLimitingInterface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformer.DeploymentInformer,
	rokibulhasanInformer informer.RokibulHasanInformer) *Controller {

	ctrl := &Controller{
		kubeclientset:      kubeclientset,
		sampleclientset:    sampleclientset,
		deploymentsLister:  deploymentInformer.Lister(),
		deploymentsSynced:  deploymentInformer.Informer().HasSynced,
		rokibulhasanLister: rokibulhasanInformer.Lister(),
		rokibulhasanSynced: rokibulhasanInformer.Informer().HasSynced,
		workQueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "RokibulHasan"),
	}

	log.Println("Setting up event handlers")

	// Set up an event handler for when RokibulHasan resources change
	rokibulhasanInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.enqueueRokibulHasan,
		UpdateFunc: func(oldObj, newObj interface{}) {
			ctrl.enqueueRokibulHasan(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			ctrl.enqueueRokibulHasan(obj)
		},
	})

	// whatif deployment resources changes??

	return ctrl
}

// enqueueRokibulHasan takes an RokibulHasan resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than RokibulHasan.
func (c *Controller) enqueueRokibulHasan(obj interface{}) {
	log.Println("Enqueueing RokibulHasan...")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.AddRateLimited(key)
}

func (c *Controller) Run(ch chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	if ok := cache.WaitForCacheSync(ch, c.rokibulhasanSynced, c.deploymentsSynced); !ok {
		log.Println("cache was not synched")
	}
	log.Println("RokibulHasan started . . .")

	log.Println("Worker started...")
	go wait.Until(c.runWorker, time.Second, ch)
	<-ch
	log.Println("Worker shutting down...")
	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {

	}
}

func (c *Controller) processNextWorkItem() bool {
	// The Get() call gets the new item from the queue and deletes it
	item, shutdown := c.workQueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so, we can defer controller.workQueue.Done
	err := func(item interface{}) error {
		defer c.workQueue.Done(item)

		var key string
		var ok bool
		if key, ok = item.(string); !ok {
			c.workQueue.Forget(item)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", item))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			fmt.Println("add rate limit")
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workQueue.Forget(item)
		return nil
	}(item)

	if err != nil {
		utilruntime.HandleError(err)
	}
	return true
}

func (c *Controller) syncHandler(key string) error {
	fmt.Println("Reconcile -------> ")
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invald resource key: %s", key))
		return nil
	}

	rokibul, err := c.rokibulhasanLister.RokibulHasans(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("RokibulHasan '%s' in work queue, no longer exists", key))
			return nil
		}
		return err
	}

	deploymentName := rokibul.Spec.DeploymentName
	log.Println("Deployment Name: ", deploymentName)

	if deploymentName == "" {
		utilruntime.HandleError(fmt.Errorf("%s: deployment name can not be empty", key))
		return nil
	}

	deployment, err := c.deploymentsLister.Deployments(namespace).Get(deploymentName)
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Create(context.TODO(), newDeployment(rokibul), metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	if rokibul.Spec.Replicas != nil && *rokibul.Spec.Replicas != *deployment.Spec.Replicas {
		log.Printf("Rokibul %s replicas: %d, deployment replicas: %d", name, *rokibul.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Update(context.TODO(), newDeployment(rokibul), metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	if rokibul.Status.AvailableReplicas != deployment.Status.Replicas {
		err = c.updateRokibulHasanStatus(rokibul, deployment)
		if err != nil {
			return err
		}
	}
	serviceName := rokibul.Spec.DeploymentName + "-service"

	service, err := c.kubeclientset.CoreV1().Services(rokibul.Namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		service, err = c.kubeclientset.CoreV1().Services(rokibul.Namespace).Create(context.TODO(), newService(rokibul), metav1.CreateOptions{})
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("\nservice %s created .....\n", service.Name)
	} else if err != nil {
		log.Println(err)
		return err
	}

	_, err = c.kubeclientset.CoreV1().Services(rokibul.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Controller) updateRokibulHasanStatus(rokibul *rokibulv1alpha1.RokibulHasan, deployment *appsv1.Deployment) error {
	rokibulCopy := rokibul.DeepCopy()
	rokibulCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// fmt.Println("CHECK ->>>> ")

	_, err := c.sampleclientset.RokibulV1alpha1().RokibulHasans(rokibul.Namespace).Update(context.TODO(), rokibulCopy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
func newService(rokibul *rokibulv1alpha1.RokibulHasan) *corev1.Service {
	labels := map[string]string{
		"app": "my-app",
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: rokibul.Spec.DeploymentName + "-service",
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(rokibul, rokibulv1alpha1.SchemeGroupVersion.WithKind("RokibulHasan")),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       rokibul.Spec.DeploymentName,
					Port:       rokibul.Spec.Container.Port,
					TargetPort: intstr.FromInt(int(rokibul.Spec.Container.Port)),
					NodePort:   30009,
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

func newDeployment(rokibul *rokibulv1alpha1.RokibulHasan) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind: "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      rokibul.Spec.DeploymentName,
			Namespace: rokibul.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(rokibul, rokibulv1alpha1.SchemeGroupVersion.WithKind("RokibulHasan")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: rokibul.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "api-server",
							Image: rokibul.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: rokibul.Spec.Container.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}
