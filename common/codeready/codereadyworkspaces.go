package codeready

import (
	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	che "github.com/eclipse-che/che-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type codeReadyUser struct {
	Username    string       `json:"username"`
	Enabled     bool         `json:"enabled"`
	Email       string       `json:"email"`
	Credentials []credential `json:"credentials"`
	ClientRoles clientRoles  `json:"clientRoles"`
}

type credential struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type clientRoles struct {
	RealmManagement []string `json:"realm-management"`
}

//
// THIS CHE/V1 API CRD is now deprecated and will be removed in the future
//
// FIXME
//
// NewCustomResource creates a Custom Resource
func NewCustomResource(workshop *workshopv1.Workshop, scheme *runtime.Scheme,
	name string, namespace string) *che.CheCluster {

	i := int32(-1)

//	openShiftoAuth := workshop.Spec.Infrastructure.CodeReadyWorkspace.OpenshiftOAuth

	pluginRegistryImage := workshop.Spec.Infrastructure.CodeReadyWorkspace.PluginRegistryImage.Name +
		":" + workshop.Spec.Infrastructure.CodeReadyWorkspace.PluginRegistryImage.Tag

	if pluginRegistryImage == ":" {
		pluginRegistryImage = ""
	}

	cr := &che.CheCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CheCluster",
			APIVersion: "v1",   // THIS CHE/V1 API CRD is now deprecated and will be removed in the future
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: che.CheClusterSpec{
			Server: che.CheClusterSpecServer{
				CustomCheProperties: map[string]string{
					"CHE_LIMITS_USER_WORKSPACES_RUN_COUNT":   "2",  // max 2 workspaces running
					"CHE_LIMITS_WORKSPACE_IDLE_TIMEOUT":  "30000000",  // timeout in milliseconds, this is about about 8 hours; @60x60x1000
					"CHE_LIMITS_WORKSPACE_RUN_TIMEOUT":   "0",
					},
				TlsSupport:           true,
				SelfSignedCert:       false,
			},
			Database: che.CheClusterSpecDB{
				ExternalDb:          false,
			},
			Auth: che.CheClusterSpecAuth{
				ExternalIdentityProvider:      false,
				IdentityProviderAdminUserName: "admin",
				IdentityProviderPassword:      "admin",
			},
			Storage: che.CheClusterSpecStorage{
				PvcStrategy:       "common",
				PvcClaimSize:      "1Gi",
			},
			DevWorkspace: che.CheClusterSpecDevWorkspace{
				SecondsOfInactivityBeforeIdling: &i,  // unlimited before timeout
			},
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, cr, scheme)

	return cr
}

// NewUser creates a user
func NewUser(username string, password string) *codeReadyUser {
	return &codeReadyUser{
		Username: username,
		Enabled:  true,
		Email:    username + "@none.com",
		Credentials: []credential{
			{
				Type:  "password",
				Value: password,
			},
		},
		ClientRoles: clientRoles{
			RealmManagement: []string{
				"user",
			},
		},
	}
}
