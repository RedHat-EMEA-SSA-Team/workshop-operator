package gitea

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func NewCustomResource(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, imageTag string, labels map[string]string) *Gitea {

	if imageTag == "" {
		imageTag = "latest"
	}

	cr := &Gitea{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: GiteaSpec{
			GiteaVolumeSize:      "4Gi",
			GiteaSsl:             true,
			PostgresqlVolumeSize: "4Gi",
			GiteaImageTag: imageTag,
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}

// CustomResourceValidation create a Custom Resource Validation
func NewCustomResourceValidation() *apiextensionsv1.CustomResourceValidation {

	crv := &apiextensionsv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"spec": {
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"giteaVolumeSize": {
							Type: "string",
						},
						"giteaImageTag": {
							Type: "string",
						},
						"giteaSsl": {
							Type: "boolean",
						},
						"giteaServiceName": {
							Type: "string",
						},
						"postgresqlVolumeSize": {
							Type: "string",
						},
					},
				},
			},
		},
	}

	return crv
}
