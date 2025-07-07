package operator

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	avsv1alpha1 "github.com/hourglass/obsidian/api/v1alpha1"
)

// AVSReconciler reconciles a AVS object
type AVSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hourglass.io,resources=avs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hourglass.io,resources=avs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hourglass.io,resources=avs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *AVSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the AVS instance
	avs := &avsv1alpha1.AVS{}
	if err := r.Get(ctx, req.NamespacedName, avs); err != nil {
		if errors.IsNotFound(err) {
			log.Info("AVS resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get AVS")
		return ctrl.Result{}, err
	}

	// Create or update Deployment
	deployment := r.deploymentForAVS(avs)
	if err := ctrl.SetControllerReference(avs, deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	foundDeployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(ctx, deployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// Update deployment if needed
		foundDeployment.Spec = deployment.Spec
		err = r.Update(ctx, foundDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create or update Service
	service := r.serviceForAVS(avs)
	if err := ctrl.SetControllerReference(avs, service, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	foundService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(ctx, service)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Update AVS status
	avs.Status.Phase = "Running"
	avs.Status.TotalReplicas = avs.Spec.Replicas
	avs.Status.LastUpdated = metav1.Now()

	// Check attestations for each pod
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(avs.Namespace),
		client.MatchingLabels(labelsForAVS(avs.Name)),
	}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods")
		return ctrl.Result{}, err
	}

	// Update attestation status
	avs.Status.Attestations = []avsv1alpha1.AttestationStatus{}
	readyCount := int32(0)

	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodRunning {
			attestationStatus := avsv1alpha1.AttestationStatus{
				PodName:     pod.Name,
				InstanceID:  string(pod.UID),
				Measurement: "pending", // Would check actual attestation
				Valid:       true,      // Would verify attestation
				LastChecked: metav1.Now(),
			}
			avs.Status.Attestations = append(avs.Status.Attestations, attestationStatus)
			readyCount++
		}
	}

	avs.Status.ReadyReplicas = readyCount

	if err := r.Status().Update(ctx, avs); err != nil {
		log.Error(err, "Failed to update AVS status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *AVSReconciler) deploymentForAVS(avs *avsv1alpha1.AVS) *appsv1.Deployment {
	labels := labelsForAVS(avs.Name)
	replicas := avs.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      avs.Name + "-deployment",
			Namespace: avs.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"hourglass.io/attestation-required": "true",
						"hourglass.io/tee-type": avs.Spec.ComputeRequirements.TEEType,
					},
				},
				Spec: corev1.PodSpec{
					NodeSelector: avs.Spec.ComputeRequirements.NodeSelector,
					Containers: []corev1.Container{{
						Name:  "obsidian-service",
						Image: avs.Spec.ServiceImage,
						Ports: []corev1.ContainerPort{{
							Name:          "http",
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "AVS_NAME",
								Value: avs.Name,
							},
							{
								Name:  "OPERATOR",
								Value: avs.Spec.Operator,
							},
							{
								Name:  "TEE_TYPE",
								Value: avs.Spec.ComputeRequirements.TEEType,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(avs.Spec.ComputeRequirements.CPU),
								corev1.ResourceMemory: resource.MustParse(avs.Spec.ComputeRequirements.Memory),
							},
						},
						SecurityContext: &corev1.SecurityContext{
							Privileged: &[]bool{true}[0], // For TEE device access
						},
					}},
				},
			},
		},
	}

	return dep
}

func (r *AVSReconciler) serviceForAVS(avs *avsv1alpha1.AVS) *corev1.Service {
	labels := labelsForAVS(avs.Name)

	port := avs.Spec.ServicePort
	if port == 0 {
		port = 8080
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      avs.Name + "-service",
			Namespace: avs.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       port,
				TargetPort: intstr.FromInt(8080),
				Protocol:   corev1.ProtocolTCP,
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	return svc
}

func labelsForAVS(name string) map[string]string {
	return map[string]string{
		"app":          "obsidian",
		"avs":          name,
		"hourglass.io": "avs",
	}
}

func (r *AVSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&avsv1alpha1.AVS{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}