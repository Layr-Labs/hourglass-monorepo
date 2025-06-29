package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/Layr-Labs/hourglass-monorepo/hourglass-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// HourglassExecutorReconciler reconciles a HourglassExecutor object
type HourglassExecutorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hourglass.eigenlayer.io,resources=hourglassexecutors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hourglass.eigenlayer.io,resources=hourglassexecutors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hourglass.eigenlayer.io,resources=hourglassexecutors/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *HourglassExecutorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the HourglassExecutor instance
	var executor v1alpha1.HourglassExecutor
	if err := r.Get(ctx, req.NamespacedName, &executor); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("HourglassExecutor resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get HourglassExecutor")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if executor.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &executor)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&executor, "hourglass.eigenlayer.io/finalizer") {
		controllerutil.AddFinalizer(&executor, "hourglass.eigenlayer.io/finalizer")
		if err := r.Update(ctx, &executor); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconcile ConfigMap
	if err := r.reconcileConfigMap(ctx, &executor); err != nil {
		logger.Error(err, "Failed to reconcile ConfigMap")
		return ctrl.Result{}, err
	}

	// Reconcile Deployment
	if err := r.reconcileDeployment(ctx, &executor); err != nil {
		logger.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}

	// Update status
	if err := r.updateStatus(ctx, &executor); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

func (r *HourglassExecutorReconciler) handleDeletion(ctx context.Context, executor *v1alpha1.HourglassExecutor) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling HourglassExecutor deletion")

	// Remove finalizer
	controllerutil.RemoveFinalizer(executor, "hourglass.eigenlayer.io/finalizer")
	if err := r.Update(ctx, executor); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *HourglassExecutorReconciler) reconcileConfigMap(ctx context.Context, executor *v1alpha1.HourglassExecutor) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      executor.Name + "-config",
			Namespace: executor.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Data = map[string]string{
			"config.yaml": r.buildExecutorConfig(executor),
		}
		return controllerutil.SetControllerReference(executor, configMap, r.Scheme)
	})

	return err
}

func (r *HourglassExecutorReconciler) reconcileDeployment(ctx context.Context, executor *v1alpha1.HourglassExecutor) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      executor.Name,
			Namespace: executor.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		replicas := int32(1)
		if executor.Spec.Replicas != nil {
			replicas = *executor.Spec.Replicas
		}

		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":                          "hourglass-executor",
					"hourglass.eigenlayer.io/name": executor.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                          "hourglass-executor",
						"hourglass.eigenlayer.io/name": executor.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "executor",
							Image: executor.Spec.Image,
							Args: []string{
								"--config=/etc/hourglass/config.yaml",
							},
							Resources: executor.Spec.Resources,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/etc/hourglass",
									ReadOnly:  true,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 9090,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/readyz",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: executor.Name + "-config",
									},
								},
							},
						},
					},
					NodeSelector:                  executor.Spec.NodeSelector,
					Tolerations:                   executor.Spec.Tolerations,
					ImagePullSecrets:              executor.Spec.ImagePullSecrets,
					ServiceAccountName:            "hourglass-executor",
					TerminationGracePeriodSeconds: &[]int64{30}[0],
				},
			},
		}

		return controllerutil.SetControllerReference(executor, deployment, r.Scheme)
	})

	return err
}

func (r *HourglassExecutorReconciler) updateStatus(ctx context.Context, executor *v1alpha1.HourglassExecutor) error {
	// Get the deployment to check status
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      executor.Name,
		Namespace: executor.Namespace,
	}, deployment); err != nil {
		return err
	}

	// Update status based on deployment state
	executor.Status.Replicas = deployment.Status.Replicas
	executor.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	executor.Status.LastConfigUpdate = &metav1.Time{Time: time.Now()}

	if deployment.Status.ReadyReplicas == deployment.Status.Replicas && deployment.Status.Replicas > 0 {
		executor.Status.Phase = "Running"
	} else if deployment.Status.Replicas == 0 {
		executor.Status.Phase = "Stopped"
	} else {
		executor.Status.Phase = "Pending"
	}

	return r.Status().Update(ctx, executor)
}

func (r *HourglassExecutorReconciler) buildExecutorConfig(executor *v1alpha1.HourglassExecutor) string {
	// Build YAML configuration for the executor
	// This is a simplified version - in practice, you'd use a proper YAML library
	config := fmt.Sprintf(`
aggregator_endpoint: "%s"
performer_mode: "%s"
log_level: "%s"
chains:
`, executor.Spec.Config.AggregatorEndpoint,
		executor.Spec.Config.PerformerMode,
		executor.Spec.Config.LogLevel)

	for _, chain := range executor.Spec.Config.Chains {
		config += fmt.Sprintf(`
  - name: "%s"
    rpc: "%s"
    chain_id: %d
    task_mailbox_address: "%s"
`, chain.Name, chain.RPC, chain.ChainID, chain.TaskMailboxAddress)
	}

	if executor.Spec.Config.Kubernetes != nil {
		config += fmt.Sprintf(`
kubernetes:
  namespace: "%s"
`, executor.Spec.Config.Kubernetes.Namespace)
	}

	return config
}

// SetupWithManager sets up the controller with the Manager.
func (r *HourglassExecutorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.HourglassExecutor{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}