package kubernetes

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewConfigMap creates a Config Map
func NewConfigMap(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, labels map[string]string, data map[string]string) *corev1.ConfigMap {

	return NewConfigMapAnnotate(workshop, scheme, name, namespace, labels, data, nil);
}

// NewConfigMap creates a Config Map with annotations
func NewConfigMapAnnotate(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, labels map[string]string, data map[string]string, annotations map[string]string) *corev1.ConfigMap {

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			Annotations: annotations,
		},
		Data: data,
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, configmap, scheme)

	return configmap
}
