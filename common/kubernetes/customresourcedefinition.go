package kubernetes

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewCustomResourceDefinition creates a Custom Resource Definition (CRD)
func NewCustomResourceDefinition(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, group string, kind string, listKind string, plural string, singular string, version string, shortNames []string, additionalPrinterColumns []apiextensionsv1.CustomResourceColumnDefinition) *apiextensionsv1.CustomResourceDefinition {

	crd := &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Scope: apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:       kind,
				ListKind:   listKind,
				Plural:     plural,
				Singular:   singular,
				ShortNames: shortNames,
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: version,
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
					AdditionalPrinterColumns: additionalPrinterColumns,
					Served:                   true,
					Storage:                  true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, crd, scheme)

	return crd
}
