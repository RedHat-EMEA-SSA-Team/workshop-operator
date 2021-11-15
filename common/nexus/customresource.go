package nexus

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// CustomResourceValidation create a Custom Resource Validation
func NewCustomResourceValidation() *apiextensionsv1.CustomResourceValidation {

	crv := &apiextensionsv1.CustomResourceValidation{
		OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextensionsv1.JSONSchemaProps{
				"spec": {
					Type: "object",
					Properties: map[string]apiextensionsv1.JSONSchemaProps{
						"nexusVolumeSize": {
							Type: "string",
						},
						"nexusSsl": {
							Type: "bool",
						},
						"nexusImageTag": {
							Type: "string",
						},
						"nexusCpuRequest": {
							Type: "int64",
						},
						"nexusCpuLimit": {
							Type: "int64",
						},
						"nexusMemoryRequest": {
							Type: "string",
						},
						"nexusMemoryLimit": {
							Type: "string",
						},
						"nexus_repos_maven_proxy": {
							Type: "array",
							Items: &apiextensionsv1.JSONSchemaPropsOrArray{
								Schema: &apiextensionsv1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"name": {
											Type: "string",
										},
										"remote_url": {
											Type: "string",
										},
										"layout_policy": {
											Type: "string",
										},
									},
								},
							},
						},
						"nexus_repos_maven_hosted": {
							Type: "array",
							Items: &apiextensionsv1.JSONSchemaPropsOrArray{
								Schema: &apiextensionsv1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"name": {
											Type: "string",
										},
										"version_policy": {
											Type: "string",
										},
										"write_policy": {
											Type: "string",
										},
									},
								},
							},
						},
						"nexus_repos_maven_group": {
							Type: "array",
							Items: &apiextensionsv1.JSONSchemaPropsOrArray{
								Schema: &apiextensionsv1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"name": {
											Type: "string",
										},
										"member_repos": {
											Type: "string",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return crv
}

// NewCustomResource create a Custom Resource
func NewCustomResource(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string, labels map[string]string) *Nexus {

	cr := &Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: NexusSpec{
			NexusVolumeSize:    "5Gi",
			NexusSSL:           true,
			NexusImageTag:      "3.18.1-01-ubi-3",
			NexusCPURequest:    1,
			NexusCPULimit:      2,
			NexusMemoryRequest: "2Gi",
			NexusMemoryLimit:   "2Gi",
			NexusReposMavenProxy: []NexusReposMavenProxySpec{
				{
					Name:         "maven-central",
					RemoteURL:    "https://repo1.maven.org/maven2/",
					LayoutPolicy: "permissive",
				},
				{
					Name:         "redhat-ga",
					RemoteURL:    "https://maven.repository.redhat.com/ga/",
					LayoutPolicy: "permissive",
				},
				{
					Name:         "jboss",
					RemoteURL:    "https://repository.jboss.org/nexus/content/groups/public",
					LayoutPolicy: "permissive",
				},
			},
			NexusReposMavenHosted: []NexusReposMavenHostedSpec{
				{
					Name:          "releases",
					VersionPolicy: "release",
					WritePolicy:   "allow_once",
				},
			},
			NexusReposMavenGroup: []NexusReposMavenGroupSpec{
				{
					Name:        "maven-all-public",
					MemberRepos: []string{"maven-central", "redhat-ga", "jboss"},
				},
			},
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}
