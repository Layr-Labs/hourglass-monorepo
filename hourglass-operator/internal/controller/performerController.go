package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hourglass-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PerformerReconciler reconciles a Performer object
type PerformerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hourglass.eigenlayer.io,resources=performers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hourglass.eigenlayer.io,resources=performers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hourglass.eigenlayer.io,resources=performers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PerformerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Performer instance
	var performer v1alpha1.Performer
	if err := r.Get(ctx, req.NamespacedName, &performer); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Performer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Performer")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if performer.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &performer)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&performer, "hourglass.eigenlayer.io/performer-finalizer") {
		controllerutil.AddFinalizer(&performer, "hourglass.eigenlayer.io/performer-finalizer")
		if err := r.Update(ctx, &performer); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile Pod
	if err := r.reconcilePod(ctx, &performer); err != nil {
		logger.Error(err, "Failed to reconcile Pod")
		return ctrl.Result{}, err
	}

	// Reconcile Service
	if err := r.reconcileService(ctx, &performer); err != nil {
		logger.Error(err, "Failed to reconcile Service")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, &performer); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute * 2}, nil
}

func (r *PerformerReconciler) handleDeletion(ctx context.Context, performer *v1alpha1.Performer) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling Performer deletion", "performer", performer.Name)

	// Remove finalizer
	controllerutil.RemoveFinalizer(performer, "hourglass.eigenlayer.io/performer-finalizer")
	if err := r.Update(ctx, performer); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PerformerReconciler) reconcilePod(ctx context.Context, performer *v1alpha1.Performer) error {
	podName := r.getPodName(performer)

	// Check if pod already exists
	existingPod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      podName,
		Namespace: performer.Namespace,
	}, existingPod)

	if err == nil {
		// Pod exists - check if it needs to be recreated
		if r.shouldRecreatePod(existingPod, performer) {
			// Delete existing pod, it will be recreated on next reconcile
			return r.Delete(ctx, existingPod)
		}
		// Pod is fine, do nothing
		return nil
	}

	if !errors.IsNotFound(err) {
		// Some other error occurred
		return err
	}

	// Pod doesn't exist, create it
	pod := r.buildPodSpec(performer)

	// Set controller reference
	if err := controllerutil.SetControllerReference(performer, pod, r.Scheme); err != nil {
		return err
	}

	return r.Create(ctx, pod)
}

// shouldRecreatePod determines if a pod should be recreated
func (r *PerformerReconciler) shouldRecreatePod(existingPod *corev1.Pod, performer *v1alpha1.Performer) bool {
	// Recreate if pod is in a failed state
	if existingPod.Status.Phase == corev1.PodFailed {
		return true
	}

	// Recreate if pod is stuck in pending for too long (more than 5 minutes)
	if existingPod.Status.Phase == corev1.PodPending {
		if time.Since(existingPod.CreationTimestamp.Time) > 5*time.Minute {
			return true
		}
	}

	// Check if the image has changed
	if len(existingPod.Spec.Containers) > 0 {
		if existingPod.Spec.Containers[0].Image != performer.Spec.Image {
			return true
		}
	}

	// Pod is healthy, keep it
	return false
}

// buildPodSpec builds the pod specification
func (r *PerformerReconciler) buildPodSpec(performer *v1alpha1.Performer) *corev1.Pod {
	// Generate labels for pod
	labels := map[string]string{
		"app":                               "hourglass-performer",
		"hourglass.eigenlayer.io/performer": performer.Name,
		"hourglass.eigenlayer.io/avs":       r.sanitizeLabel(performer.Spec.AVSAddress),
	}

	// Build container spec
	container := corev1.Container{
		Name:            "performer",
		Image:           performer.Spec.Image,
		ImagePullPolicy: performer.Spec.ImagePullPolicy,
	}

	// Set command and args if specified
	if len(performer.Spec.Config.Command) > 0 {
		container.Command = performer.Spec.Config.Command
	}
	if len(performer.Spec.Config.Args) > 0 {
		container.Args = performer.Spec.Config.Args
	}

	// Configure gRPC port
	grpcPort := int32(8080) // Default to 8080 to match hello-performer
	if performer.Spec.Config.GRPCPort != 0 {
		grpcPort = performer.Spec.Config.GRPCPort
	}
	container.Ports = []corev1.ContainerPort{
		{
			Name:          "grpc",
			ContainerPort: grpcPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	// Set environment variables
	if performer.Spec.Config.Environment != nil {
		for key, value := range performer.Spec.Config.Environment {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  key,
				Value: value,
			})
		}
	}

	// Set resources
	container.Resources = performer.Spec.Resources

	// Apply hardware requirements to resources
	if performer.Spec.HardwareRequirements != nil {
		r.applyHardwareRequirements(&container, performer.Spec.HardwareRequirements)
	}

	// Configure probes
	container.LivenessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt32(grpcPort),
			},
		},
		InitialDelaySeconds: 30,
		PeriodSeconds:       10,
	}
	container.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromInt32(grpcPort),
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       5,
	}

	// Build pod spec
	podSpec := corev1.PodSpec{
		Containers:       []corev1.Container{container},
		ImagePullSecrets: performer.Spec.ImagePullSecrets,
		RestartPolicy:    corev1.RestartPolicyAlways,
	}

	// Apply scheduling constraints
	if performer.Spec.Scheduling != nil {
		podSpec.NodeSelector = performer.Spec.Scheduling.NodeSelector
		podSpec.Affinity = performer.Spec.Scheduling.Affinity
		podSpec.Tolerations = performer.Spec.Scheduling.Tolerations
		if performer.Spec.Scheduling.RuntimeClass != nil {
			podSpec.RuntimeClassName = performer.Spec.Scheduling.RuntimeClass
		}
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.getPodName(performer),
			Namespace: performer.Namespace,
			Labels:    labels,
		},
		Spec: podSpec,
	}
}

func (r *PerformerReconciler) reconcileService(ctx context.Context, performer *v1alpha1.Performer) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.getServiceName(performer),
			Namespace: performer.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		grpcPort := int32(8080) // Default to 8080 to match hello-performer
		if performer.Spec.Config.GRPCPort != 0 {
			grpcPort = performer.Spec.Config.GRPCPort
		}

		service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{
				"hourglass.eigenlayer.io/performer": performer.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "grpc",
					Port:       grpcPort,
					TargetPort: intstr.FromString("grpc"),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		}

		return controllerutil.SetControllerReference(performer, service, r.Scheme)
	})

	return err
}

func (r *PerformerReconciler) updateStatus(ctx context.Context, performer *v1alpha1.Performer) error {
	// Retry logic for status updates to handle concurrent modifications
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// Get the latest version of the performer
		latest := &v1alpha1.Performer{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      performer.Name,
			Namespace: performer.Namespace,
		}, latest); err != nil {
			return err
		}

		// Get the pod to check status
		pod := &corev1.Pod{}
		podName := r.getPodName(latest)
		if err := r.Get(ctx, types.NamespacedName{
			Name:      podName,
			Namespace: latest.Namespace,
		}, pod); err != nil {
			if errors.IsNotFound(err) {
				latest.Status.Phase = "Pending"
				latest.Status.PodName = ""
			} else {
				return err
			}
		} else {
			latest.Status.PodName = podName
			latest.Status.Phase = string(pod.Status.Phase)

			// Check if pod is ready
			latest.Status.Ready = false
			if pod.Status.Phase == corev1.PodRunning {
				for _, condition := range pod.Status.Conditions {
					if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
						latest.Status.Ready = true
						if latest.Status.ReadyTime == nil {
							latest.Status.ReadyTime = &metav1.Time{Time: time.Now()}
						}
						break
					}
				}
			}
		}

		// Set service information
		latest.Status.ServiceName = r.getServiceName(latest)
		latest.Status.GRPCEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local",
			latest.Status.ServiceName, latest.Namespace)

		// Try to update the status
		if err := r.Status().Update(ctx, latest); err != nil {
			if errors.IsConflict(err) && i < maxRetries-1 {
				// Wait a bit before retrying
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
				continue
			}
			return err
		}

		// Success
		return nil
	}

	return fmt.Errorf("failed to update status after %d retries", maxRetries)
}

func (r *PerformerReconciler) getPodName(performer *v1alpha1.Performer) string {
	return fmt.Sprintf("performer-%s", performer.Name)
}

func (r *PerformerReconciler) getServiceName(performer *v1alpha1.Performer) string {
	return fmt.Sprintf("performer-%s", performer.Name)
}

func (r *PerformerReconciler) sanitizeLabel(value string) string {
	// Kubernetes labels must be alphanumeric with dashes and dots
	result := strings.ToLower(value)
	result = strings.ReplaceAll(result, "_", "-")
	if len(result) > 63 {
		result = result[:63]
	}
	return result
}

func (r *PerformerReconciler) applyHardwareRequirements(container *corev1.Container, hw *v1alpha1.HardwareRequirements) {
	if container.Resources.Limits == nil {
		container.Resources.Limits = make(corev1.ResourceList)
	}
	if container.Resources.Requests == nil {
		container.Resources.Requests = make(corev1.ResourceList)
	}

	// Apply GPU requirements
	if hw.GPUCount > 0 {
		gpuResource := corev1.ResourceName("nvidia.com/gpu")
		if hw.GPUType != "" {
			// Use specific GPU type resource if specified
			gpuResource = corev1.ResourceName(fmt.Sprintf("nvidia.com/%s", hw.GPUType))
		}

		gpuQuantity := resource.MustParse(fmt.Sprintf("%d", hw.GPUCount))
		container.Resources.Limits[gpuResource] = gpuQuantity
		container.Resources.Requests[gpuResource] = gpuQuantity
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PerformerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Performer{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
