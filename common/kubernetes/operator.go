package kubernetes

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewAnsibleOperatorDeployment creates an Ansible Operator Deployment
func NewAnsibleOperatorDeployment(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, labels map[string]string, image string, serviceAccountName string) *appsv1.Deployment {

	operator := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{
						{
							Name:            "ansible",
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							Command: []string{
								"/usr/local/bin/ao-logs",
								"/tmp/ansible-operator/runner",
								"stdout",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "runner",
									MountPath: "/tmp/ansible-operator/runner",
									ReadOnly:  true,
								},
							},
						},
						{
							Name:            "operator",
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "runner",
									MountPath: "/tmp/ansible-operator/runner",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "WATCH_NAMESPACE",
									Value: "",
								},
								{
									Name:  "OPERATOR_NAME",
									Value: name,
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "ANSIBLE_GATHERING",
									Value: "explicit",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "runner",
						},
					},
				},
			},
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, operator, scheme)

	return operator
}

// NewOperatorDeployment creates an Operator Deployment
func NewOperatorDeployment(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, labels map[string]string, image string, serviceAccountName string, metricsPort int32, commands []string, args []string, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) *appsv1.Deployment {

	operator := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					Containers: []corev1.Container{
						{
							Name:            "operator",
							Image:           image,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts:    volumeMounts,
							Ports: []corev1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: metricsPort,
									Protocol:      "TCP",
								},
							},
							Command: commands,
							Args:    args,
							Env: []corev1.EnvVar{
								{
									Name:  "WATCH_NAMESPACE",
									Value: "",
								},
								{
									Name:  "OPERATOR_NAME",
									Value: name,
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, operator, scheme)

	return operator
}
