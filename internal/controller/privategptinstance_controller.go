package controller

import (
	"context"
    "fmt"
    "time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
    "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	privategptv1alpha1 "github.com/msimonelli331/privategpt-operator/api/v1alpha1"
)

// PrivateGPTInstanceReconciler reconciles a PrivateGPTInstance object
type PrivateGPTInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=privategpt.eirl,resources=privategptinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=privategpt.eirl,resources=privategptinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=privategpt.eirl,resources=privategptinstances/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=deployments/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PrivateGPTInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *PrivateGPTInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

    privateGPTInstance := &privategptv1alpha1.PrivateGPTInstance{}
    err := r.Get(ctx, req.NamespacedName, privateGPTInstance)
    if err != nil {
        if apierrors.IsNotFound(err) {
            // If the custom resource is not found then it usually means that it was deleted or not created
            // In this way, we will stop the reconciliation
            log.Info("privateGPTInstance resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        // Error reading the object - requeue the request.
        log.Error(err, "Failed to get privateGPTInstance")
        return ctrl.Result{}, err
    }
	
	log.Info("Processing PrivateGPTInstance", "name", privateGPTInstance.Name)

    // Let's just set the status as Unknown when no status is available
    if len(privateGPTInstance.Status.Conditions) == 0 {
        meta.SetStatusCondition(&privateGPTInstance.Status.Conditions, metav1.Condition{Type: "Available", Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
        if err = r.Status().Update(ctx, privateGPTInstance); err != nil {
            log.Error(err, "Failed to update privateGPTInstance status")
            return ctrl.Result{}, err
        }

        // Let's re-fetch the privateGPTInstance Custom Resource after updating the status
        // so that we have the latest state of the resource on the cluster and we will avoid
        // raising the error "the object has been modified, please apply
        // your changes to the latest version and try again" which would re-trigger the reconciliation
        // if we try to update it again in the following operations
        if err := r.Get(ctx, req.NamespacedName, privateGPTInstance); err != nil {
            log.Error(err, "Failed to re-fetch privateGPTInstance")
            return ctrl.Result{}, err
        }
    }

	// Check if a Deployment for the PrivateGPTInstance exists, if not, create one
	deploymentFound := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: privateGPTInstance.Name, Namespace: privateGPTInstance.Namespace}, deploymentFound)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new deployment
		dep, err := r.deploymentForInstance(privateGPTInstance)
		if err != nil {
			log.Error(err, "Failed to define new Deployment resource for privateGPTInstance")
			
            // The following implementation will update the status
            meta.SetStatusCondition(&privateGPTInstance.Status.Conditions, metav1.Condition{Type: "Available",
                Status: metav1.ConditionFalse, Reason: "Reconciling",
                Message: fmt.Sprintf("Failed to create Deployment for the custom resource (%s): (%s)", privateGPTInstance.Name, err)})

            if err := r.Status().Update(ctx, privateGPTInstance); err != nil {
                log.Error(err, "Failed to update privateGPTInstance status")
                return ctrl.Result{}, err
            }

            return ctrl.Result{}, err
		}

        log.Info("Creating a new Deployment",
            "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
        if err = r.Create(ctx, dep); err != nil {
            log.Error(err, "Failed to create new Deployment",
                "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
            return ctrl.Result{}, err
        }

        // Deployment created successfully
        // We will requeue the reconciliation so that we can ensure the state
        // and move forward for the next operations
        return ctrl.Result{RequeueAfter: time.Minute}, nil
    } else if err != nil {
        log.Error(err, "Failed to get Deployment")
        // Let's return the error for the reconciliation be re-trigged again
        return ctrl.Result{}, err
    }

	// Check if a Service for the PrivateGPTInstance exists, if not, create one
	serviceFound := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: privateGPTInstance.Name, Namespace: privateGPTInstance.Namespace}, serviceFound)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new service
		svc, err := r.serviceForInstance(privateGPTInstance)
		if err != nil {
			log.Error(err, "Failed to define new Service resource for privateGPTInstance")
			
            // The following implementation will update the status
            meta.SetStatusCondition(&privateGPTInstance.Status.Conditions, metav1.Condition{Type: "Available",
                Status: metav1.ConditionFalse, Reason: "Reconciling",
                Message: fmt.Sprintf("Failed to create Service for the custom resource (%s): (%s)", privateGPTInstance.Name, err)})

            if err := r.Status().Update(ctx, privateGPTInstance); err != nil {
                log.Error(err, "Failed to update privateGPTInstance status")
                return ctrl.Result{}, err
            }

            return ctrl.Result{}, err
		}

        log.Info("Creating a new Service",
            "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
        if err = r.Create(ctx, dep); err != nil {
            log.Error(err, "Failed to create new Service",
                "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
            return ctrl.Result{}, err
        }

        // Service created successfully
        // We will requeue the reconciliation so that we can ensure the state
        // and move forward for the next operations
        return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
        log.Error(err, "Failed to get Service")
        // Let's return the error for the reconciliation be re-trigged again
        return ctrl.Result{}, err
    }

	// Check if an Ingress for the PrivateGPTInstance exists, if not, create one
	ingressFound := &networkingv1.Ingress{}
	err = r.Get(ctx, types.NamespacedName{Name: privateGPTInstance.Name, Namespace: privateGPTInstance.Namespace}, ingressFound)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new ingress
		ing, err := r.ingressForInstance(privateGPTInstance)
		if err != nil {
			log.Error(err, "Failed to define new Ingress resource for privateGPTInstance")
			
            // The following implementation will update the status
            meta.SetStatusCondition(&privateGPTInstance.Status.Conditions, metav1.Condition{Type: "Available",
                Status: metav1.ConditionFalse, Reason: "Reconciling",
                Message: fmt.Sprintf("Failed to create Ingress for the custom resource (%s): (%s)", privateGPTInstance.Name, err)})

            if err := r.Status().Update(ctx, privateGPTInstance); err != nil {
                log.Error(err, "Failed to update privateGPTInstance status")
                return ctrl.Result{}, err
            }

            return ctrl.Result{}, err
		}

        log.Info("Creating a new Ingress",
            "Ingress.Namespace", ing.Namespace, "Ingress.Name", ing.Name)
        if err = r.Create(ctx, dep); err != nil {
            log.Error(err, "Failed to create new Ingress",
                "Ingress.Namespace", ing.Namespace, "Ingress.Name", ing.Name)
            return ctrl.Result{}, err
        }

        // Ingress created successfully
        // We will requeue the reconciliation so that we can ensure the state
        // and move forward for the next operations
        return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
        log.Error(err, "Failed to get Ingress")
        // Let's return the error for the reconciliation be re-trigged again
        return ctrl.Result{}, err
    }

	// The following implementation will update the status
    meta.SetStatusCondition(&privateGPTInstance.Status.Conditions, metav1.Condition{Type: "Available",
        Status: metav1.ConditionTrue, Reason: "Reconciling",
        Message: fmt.Sprintf("Deployment, Service, and Ingress for custom resource (%s) created successfully", privateGPTInstance.Name)})

    if err := r.Status().Update(ctx, privateGPTInstance); err != nil {
        log.Error(err, "Failed to update privateGPTInstance status")
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrivateGPTInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&privategptv1alpha1.PrivateGPTInstance{}).
        Owns(&appsv1.Deployment{}).
        Owns(&corev1.Service{}).
        Owns(&networkingv1.Ingress{}).
		Named("privategptinstance").
		Complete(r)
}

// deploymentForInstance returns a PrivateGPTInstance Deployment object
func (r *PrivateGPTInstanceReconciler) deploymentForInstance(
	privateGPTInstance privategptv1alpha1.PrivateGPTInstance) (*appsv1.Deployment, error) {
	image := privateGPTInstance.Spec.Image

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      privateGPTInstance.Name,
			Namespace: privateGPTInstance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: 1,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app.kubernetes.io/name": privateGPTInstance.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app.kubernetes.io/name": privateGPTInstance.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           image,
						Name:            "privategpt",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{
							{
								Name:  "OLLAMA_URL",
								Value: privateGPTInstance.Spec.OllamaURL,
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8001,
							Name:          "http",
							Protocol:      corev1.ProtocolTCP,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/files",
							Name:      "ingest-volume",
						}},
						Command: []string{"run", privateGPTInstance.Name},
					}},
					Volumes: []corev1.Volume{{
						Name: "ingest-volume",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "privategpt",
							},
						},
					}},
				},
			},
		},
	}

	// Set the ownerRef for the Deployment
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(&privateGPTInstance, dep, r.Scheme); err != nil {
		return nil, err
	}
	return dep, nil
}

// serviceForInstance returns a PrivateGPTInstance Service object
func (r *PrivateGPTInstanceReconciler) serviceForInstance(
	privateGPTInstance privategptv1alpha1.PrivateGPTInstance) (*corev1.Service, error) {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      privateGPTInstance.Name,
			Namespace: privateGPTInstance.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": privateGPTInstance.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app.kubernetes.io/name": privateGPTInstance.Name,
			},
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       8001,
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.FromString("http"),
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	// Set the ownerRef for the Service
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(&privateGPTInstance, svc, r.Scheme); err != nil {
		return nil, err
	}
	return svc, nil
}

// ingressForInstance returns a PrivateGPTInstance Ingress object
func (r *PrivateGPTInstanceReconciler) ingressForInstance(
	privateGPTInstance privategptv1alpha1.PrivateGPTInstance) (*networkingv1.Ingress, error) {
    hostname := privateGPTInstance.Name + "." + "pgpt" + "." + privateGPTInstance.Domain

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      privateGPTInstance.Name,
			Namespace: privateGPTInstance.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": privateGPTInstance.Name,
			},
            Annotations: map[string]string{
				"cert-manager.io/common-name": hostname,
				"cert-manager.io/issuer": "ca-issuer",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: stringPtr("traefik"),
			Rules: []networkingv1.IngressRule{{
				Host: hostname,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: networkingv1.PathTypePrefix,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: privateGPTInstance.Name,
									Port: networkingv1.ServiceBackendPort{
										Number: 8001,
									},
								},
							},
						}},
					},
				},
			}},
			TLS: []networkingv1.IngressTLS{{
				Hosts:      []string{hostname},
				SecretName: privateGPTInstance.Name,
			}},
		},
	}

	// Set the ownerRef for the Ingress
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/owners-dependents/
	if err := ctrl.SetControllerReference(&privateGPTInstance, ingress, r.Scheme); err != nil {
		return nil, err
	}
	return ingress, nil
}

func stringPtr(s string) *string {
	return &s
}
