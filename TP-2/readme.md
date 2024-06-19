#### Resources webapp_types.go

```go
type WebAppSpec struct {
    AppName          string `json:"appName"`
    Image            string `json:"image"`
    DBImage          string `json:"dbImage"`
    Replicas         int32  `json:"replicas"`
    DBSize           string `json:"dbSize"`
    AutoScaleEnabled bool   `json:"autoScaleEnabled"`
    TrafficThreshold int32  `json:"trafficThreshold"`
}

type WebAppStatus struct {
    AvailableReplicas int32 `json:"availableReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AppName",type="string",JSONPath=".spec.appName"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="AutoScale",type="boolean",JSONPath=".spec.autoScaleEnabled"
```

### Resource webapp_controller 

```go
type WebAppReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=web.example.com,resources=webapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=web.example.com,resources=webapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=web.example.com,resources=webapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

func (r *WebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)


    var webApp webappv1.WebApp
    if err := r.Get(ctx, req.NamespacedName, &webApp); err != nil {
        log.Error(err, "unable to fetch WebApp")
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    deployment := &appsv1.Deployment{}
    deploymentName := types.NamespacedName{Name: webApp.Spec.AppName + "-frontend", Namespace: webApp.Namespace}
    if err := r.Get(ctx, deploymentName, deployment); err != nil {
        if errors.IsNotFound(err) {
            deployment = r.frontendDeployment(webApp)
            log.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
            if err := r.Create(ctx, deployment); err != nil {
                log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
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


func (r *WebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&webappv1.WebApp{}).
        Owns(&appsv1.Deployment{}).
        Complete(r)
}

func (r *WebAppReconciler) frontendDeployment(webApp webappv1.WebApp) *appsv1.Deployment {
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


func (r *WebAppReconciler) ensureHPA(webApp webv1.WebApp, deployment appsv1.Deployment) error {
    hpaName := webApp.Spec.AppName + "-hpa"
    var hpa autoscalingv2beta2.HorizontalPodAutoscaler
    err := r.Get(context.TODO(), types.NamespacedName{
        Name:      hpaName,
        Namespace: deployment.Namespace,
    }, &hpa)

    if errors.IsNotFound(err) {
        hpa = autoscalingv2beta2.HorizontalPodAutoscaler{
            ObjectMeta: metav1.ObjectMeta{
                Name:      hpaName,
                Namespace: deployment.Namespace,
            },
            Spec: autoscalingv2beta2.HorizontalPodAutoscalerSpec{
                ScaleTargetRef: autoscalingv2beta2.CrossVersionObjectReference{
                    Kind:       "Deployment",
                    Name:       deployment.Name,
                    APIVersion: "apps/v1",
                },
                MinReplicas: &webApp.Spec.Replicas,
                MaxReplicas: webApp.Spec.Replicas * 2,
                Metrics: []autoscalingv2beta2.MetricSpec{
                    {
                        Type: autoscalingv2beta2.ResourceMetricSourceType,
                        Resource: &autoscalingv2beta2.ResourceMetricSource{
                            Name: corev1.ResourceCPU,
                            Target: autoscalingv2beta2.MetricTarget{
                                Type:               autoscalingv2beta2.UtilizationMetricType,
                                AverageUtilization: pointer.Int32Ptr(50), // Utilisation moyenne de CPU à 50%
                            },
                        },
                    },
                    {
                        Type: autoscalingv2beta2.PodsMetricSourceType,
                        Pods: &autoscalingv2beta2.PodsMetricSource{
                            Metric: autoscalingv2beta2.MetricIdentifier{
                                Name: "http_requests",
                            },
                            Target: autoscalingv2beta2.MetricTarget{
                                Type:               autoscalingv2beta2.AverageValueMetricType,
                                AverageValue:       resource.NewQuantity(int64(webApp.Spec.TrafficThreshold), resource.DecimalSI),
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
```