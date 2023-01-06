package kubernetes

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
)

// NewNamespace creates a new namespace/project
func NewNamespace(workshop *workshopv1.Workshop, scheme *runtime.Scheme, name string) *corev1.Namespace {

	return NewNamespaceAnnotate(workshop, scheme, name, nil, nil)
}

// NewNamespace creates a new namespace/project
func NewNamespaceAnnotate(workshop *workshopv1.Workshop, scheme *runtime.Scheme, name string, labels map[string]string, annotations map[string]string) *corev1.Namespace {

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: labels,
			Annotations: annotations,
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, namespace, scheme)

	return namespace
}
