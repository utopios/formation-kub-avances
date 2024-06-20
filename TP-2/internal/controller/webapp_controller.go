/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	utopiosnetv1 "github.com/utopios/webapp/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// WebAppReconciler reconciles a WebApp object
type WebAppReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=utopios.net.utopios.net,resources=webapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=utopios.net.utopios.net,resources=webapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=utopios.net.utopios.net,resources=webapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WebApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *WebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := log.FromContext(ctx)

	var webApp utopiosnetv1.WebApp
	if err := r.Get(ctx, req.NamespacedName, &webApp); err != nil {
		log.Error(err, "unable to fetch WebApp")
		//A corriger
		return ctrl.Result{RequeueAfter: time.Minute}, client.IgnoreNotFound(err)
	}

	deployment := &appsv1.Deployment{}
	deploymentName := types.NamespacedName{Name: webApp.Spec.AppName + "-frontend", Namespace: webApp.Namespace}
	if err := r.Get(ctx, deploymentName, deployment); err != nil {
		if errors.IsNotFound(err) {
			deployment = r.frontendDeployment(webApp)
			log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			if err := r.Create(ctx, deployment); err != nil {
				log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
				r.Recorder.Event(&webApp, corev1.EventTypeWarning, "ReconcileError", err.Error())
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	updated := false
	if *deployment.Spec.Replicas != webApp.Spec.Replicas {
		deployment.Spec.Replicas = &webApp.Spec.Replicas
		updated = true
	}
	if deployment.Spec.Template.Spec.Containers[0].Image != webApp.Spec.Image {
		deployment.Spec.Template.Spec.Containers[0].Image = webApp.Spec.Image
		updated = true
	}
	if updated {
		if err := r.Update(ctx, deployment); err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}
	}

	// Manage auto-scaling
	if webApp.Spec.AutoScaleEnabled {
		if err := r.ensureHPA(webApp, *deployment); err != nil {
			log.Error(err, "Failed to ensure HPA", "HPA.Namespace", deployment.Namespace, "HPA.Name", deployment.Name)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&utopiosnetv1.WebApp{}).
		Complete(r)
}

func (r *WebAppReconciler) frontendDeployment(webApp utopiosnetv1.WebApp) *appsv1.Deployment {
	labels := map[string]string{"app": webApp.Spec.AppName, "tier": "frontend"}
	replicas := webApp.Spec.Replicas

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webApp.Spec.AppName + "-frontend",
			Namespace: webApp.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "web",
						Image: webApp.Spec.Image,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
						}},
					}},
				},
			},
		},
	}
}

func (r *WebAppReconciler) ensureHPA(webApp utopiosnetv1.WebApp, deployment appsv1.Deployment) error {
	hpaName := webApp.Spec.AppName + "-hpa"
	var hpa autoscalingv2.HorizontalPodAutoscaler
	err := r.Get(context.TODO(), types.NamespacedName{
		Name:      hpaName,
		Namespace: deployment.Namespace,
	}, &hpa)

	if errors.IsNotFound(err) {
		hpa = autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      hpaName,
				Namespace: deployment.Namespace,
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind:       "Deployment",
					Name:       deployment.Name,
					APIVersion: "apps/v1",
				},
				MinReplicas: &webApp.Spec.Replicas,
				MaxReplicas: webApp.Spec.Replicas * 2,
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: corev1.ResourceCPU,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: pointer.Int32Ptr(50), // Utilisation moyenne de CPU à 50%
							},
						},
					},
					{
						Type: autoscalingv2.PodsMetricSourceType,
						Pods: &autoscalingv2.PodsMetricSource{
							Metric: autoscalingv2.MetricIdentifier{
								Name: "http_requests",
							},
							Target: autoscalingv2.MetricTarget{
								Type:         autoscalingv2.AverageValueMetricType,
								AverageValue: resource.NewQuantity(int64(webApp.Spec.TrafficThreshold), resource.DecimalSI),
							},
						},
					},
				},
			},
		}

		err = r.Create(context.TODO(), &hpa)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
