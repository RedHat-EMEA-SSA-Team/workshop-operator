package istio

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ist "github.com/istio-ecosystem/sail-operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)


func NewSailCR(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, discoveryLabels map[string]string) *ist.Istio {

	labelSelector := &metav1.LabelSelector{
		MatchLabels: discoveryLabels,
	}

	var discoverySelectors []*metav1.LabelSelector = make([]*metav1.LabelSelector, 1)

	discoverySelectors[0] = labelSelector

	meshConfig := &ist.MeshConfig{
					DiscoverySelectors: discoverySelectors,
				}

    values := &ist.Values{
				MeshConfig: meshConfig,
				}

	cr := &ist.Istio{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sailoperator.io/v1",
			Kind:       "Istio",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
		},
		Spec: ist.IstioSpec{
			Namespace: namespace,
			Values: values,
			Version: "v1.27-latest",
			},
		}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}

func NewCNICR(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, discoveryLabels map[string]string) *ist.IstioCNI {


	cr := &ist.IstioCNI{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "sailoperator.io/v1",
			Kind:       "IstioCNI",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
		},
		Spec: ist.IstioCNISpec{
			Namespace: namespace,
			Version: "v1.27-latest",
			},
		}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}

func NewKialiMonitoringUnstructured(name string, namespace string) *unstructured.Unstructured {
	kiali := &unstructured.Unstructured{}
	kiali.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kiali.io/v1alpha1",
		"kind":       "Kiali",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"external_services": map[string]interface{}{
				"prometheus": map[string]interface{}{
					"url": "https://thanos-querier.openshift-monitoring.svc.cluster.local:9091",
					"auth": map[string]interface{}{
						"type":            "bearer",
						"use_kiali_token": true,
					},
					"thanos_proxy": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			"deployment": map[string]interface{}{
				"discovery_selectors": map[string]interface{}{
					"default": []interface{}{
						map[string]interface{}{
							"matchLabels": map[string]interface{}{
								"istio-discovery": "enabled",
							},
						},
					},
				},
			},
		},
	})
	return kiali
}

func NewKialiCR(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string) *unstructured.Unstructured  {

	cr := NewKialiMonitoringUnstructured(name, namespace)

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}

func NewTelemetryUnstructured(name string, namespace string) *unstructured.Unstructured {
	tel := &unstructured.Unstructured{}
	tel.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "telemetry.istio.io/v1",
		"kind":       "Telemetry",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
					"metrics": []interface{}{
						map[string]interface{}{
							"providers": []interface{}{
								map[string]interface{}{
									"name": "prometheus",
								},
							},
						},
					},
				},
	})
	return tel
}

func NewTelemetryCR(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string) *unstructured.Unstructured {

	cr := NewTelemetryUnstructured(name, namespace)

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}

func NewIstiodServiceMonitorUnstructured(name string, namespace string) *unstructured.Unstructured {
	sm := &unstructured.Unstructured{}

	sm.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "monitoring.coreos.com/v1",
		"kind":       "ServiceMonitor",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"targetLabels": []interface{}{
				"app",
			},
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"istio": "pilot",
				},
			},
			"endpoints": []interface{}{
				map[string]interface{}{
					"port":     "http-monitoring",
					"interval": "30s",
				},
			},
		},
	})

	return sm
}

func NewServiceMonitorCR(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string) *unstructured.Unstructured {

	cr := NewIstiodServiceMonitorUnstructured(name, namespace)

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}

func NewPodMonitorUnstructured(name string, namespace string) *unstructured.Unstructured {
	pm := &unstructured.Unstructured{}

	pm.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "monitoring.coreos.com/v1",
		"kind":       "PodMonitor",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"matchExpressions": []interface{}{
					map[string]interface{}{
						"key":      "istio-prometheus-ignore",
						"operator": "DoesNotExist",
					},
				},
			},
			"podMetricsEndpoints": []interface{}{
				map[string]interface{}{
					"path":     "/stats/prometheus",
					"interval": "30s",
					"relabelings": []interface{}{
						map[string]interface{}{
							"action":       "keep",
							"sourceLabels": []interface{}{"__meta_kubernetes_pod_container_name"},
							"regex":        "istio-proxy",
						},
						map[string]interface{}{
							"action":       "keep",
							"sourceLabels": []interface{}{"__meta_kubernetes_pod_annotationpresent_prometheus_io_scrape"},
						},
						map[string]interface{}{
							"action":       "replace",
							"regex":        `(\d+);(([A-Fa-f0-9]{1,4}::?){1,7}[A-Fa-f0-9]{1,4})`,
							"replacement":  `[$2]:$1`,
							"sourceLabels": []interface{}{"__meta_kubernetes_pod_annotation_prometheus_io_port", "__meta_kubernetes_pod_ip"},
							"targetLabel":  "__address__",
						},
						map[string]interface{}{
							"action":       "replace",
							"regex":        `(\d+);((([0-9]+?)(\.|$)){4})`,
							"replacement":  `$2:$1`,
							"sourceLabels": []interface{}{"__meta_kubernetes_pod_annotation_prometheus_io_port", "__meta_kubernetes_pod_ip"},
							"targetLabel":  "__address__",
						},
						map[string]interface{}{
							"action": "labeldrop",
							"regex":  "__meta_kubernetes_pod_label_(.+)",
						},
						map[string]interface{}{
							"sourceLabels": []interface{}{"__meta_kubernetes_namespace"},
							"action":       "replace",
							"targetLabel":  "namespace",
						},
						map[string]interface{}{
							"sourceLabels": []interface{}{"__meta_kubernetes_pod_name"},
							"action":       "replace",
							"targetLabel":  "pod_name",
						},
					},
				},
			},
		},
	})

	return pm
}

func NewPodMonitorCR(workshop *workshopv1.Workshop, scheme *runtime.Scheme, name string, namespace string) *unstructured.Unstructured {
	cr := NewPodMonitorUnstructured(name, namespace)

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr

}

func NewOSSMConsoleUnstructured(name string, namespace string) *unstructured.Unstructured {
	console := &unstructured.Unstructured{}

	console.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "kiali.io/v1alpha1",
		"kind":       "OSSMConsole",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"version": "default",
		},
	})

	return console
}

func NewOSSMConsoleCR(workshop *workshopv1.Workshop, scheme *runtime.Scheme) *unstructured.Unstructured {

	cr := NewOSSMConsoleUnstructured("ossmconsole", "openshift-operators")

		// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}