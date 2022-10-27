package controllers

import (
	//	"bytes"
	"context"
	"crypto/tls"

	//	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/codeready"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/kubernetes"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"
	"github.com/prometheus/common/log"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)


// Reconciling CodeReadyWorkspace
func (r *WorkshopReconciler) reconcileCodeReadyWorkspace(workshop *workshopv1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {
	enabled := workshop.Spec.Infrastructure.CodeReadyWorkspace.Enabled

	if enabled {
		if result, err := r.addCodeReadyWorkspace(workshop, users, appsHostnameSuffix, openshiftConsoleURL); util.IsRequeued(result, err) {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addCodeReadyWorkspace(workshop *workshopv1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {

	//const InstallNameSpace = "openshift-operators"
	const InstallNameSpace = "openshift-devspaces"
	const CheNameSpace = "openshift-devspaces"
	const OperatorGroupName = "devspaces"
	const OperatorDeployment = "devspaces-operator"
	const DevSpacesDeployment = "devspaces"
	const SubscriptionName = "devspaces"
	const PackageName = "devspaces"
	const InstallPlan = "devspaces"
	const CheClusterCustomResource = "devspaces"
	const CheURLCodeFlavour = "devspaces"

	channel := workshop.Spec.Infrastructure.CodeReadyWorkspace.OperatorHub.Channel
	clusterServiceVersion := workshop.Spec.Infrastructure.CodeReadyWorkspace.OperatorHub.ClusterServiceVersion

	codeReadyWorkspacesInstall := kubernetes.NewNamespace(workshop, r.Scheme, InstallNameSpace)
	if err := r.Create(context.TODO(), codeReadyWorkspacesInstall); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created DevSpace %s Project", codeReadyWorkspacesInstall.Name)
	}

	codeReadyWorkspacesOperatorGroup := kubernetes.NewOperatorGroup(workshop, r.Scheme, OperatorGroupName, codeReadyWorkspacesInstall.Name, "")
	if err := r.Create(context.TODO(), codeReadyWorkspacesOperatorGroup); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s OperatorGroup", codeReadyWorkspacesOperatorGroup.Name)
	}

	codeReadyWorkspacesSubscription := kubernetes.NewRedHatSubscription(workshop, r.Scheme, SubscriptionName, codeReadyWorkspacesInstall.Name,
		PackageName, channel, clusterServiceVersion)
	if err := r.Create(context.TODO(), codeReadyWorkspacesSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", codeReadyWorkspacesSubscription.Name)
	}

	// Approve the installation
	if err := r.ApproveInstallPlan(clusterServiceVersion, InstallPlan, codeReadyWorkspacesInstall.Name); err != nil {
		log.Warnf("Waiting for Subscription to create InstallPlan for %s", InstallPlan)
		return reconcile.Result{Requeue: true}, nil
	}

	// Wait for CodeReadyWorkspace Operator to be running
	if !kubernetes.GetK8Client().GetDeploymentStatus(OperatorDeployment, codeReadyWorkspacesInstall.Name) {
		return reconcile.Result{Requeue: true}, nil
	}

	codeReadyWorkspacesNamespace := kubernetes.NewNamespace(workshop, r.Scheme, CheNameSpace)
	if err := r.Create(context.TODO(), codeReadyWorkspacesNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created Che Custom resource %s Project", codeReadyWorkspacesNamespace.Name)
	}

	codeReadyWorkspacesCustomResource := codeready.NewCustomResource(workshop, r.Scheme, CheClusterCustomResource, codeReadyWorkspacesNamespace.Name)
	err := r.Create(context.TODO(), codeReadyWorkspacesCustomResource); 
	
	if (err == nil) {
		log.Infof("Created %s Custom Resource", codeReadyWorkspacesCustomResource.Name)

	} else if (errors.ReasonForError(err) == "only one CheCluster is allowed") {
		// Now Dev Spaces only allows one instance of the Che CR at the moment, so reports an attempt to add an extra one 
		// as forbidden 403 and not as "AlreadyExists". So we need to ignore that
//		log.Infof("An instance of %s Custom Resource already exists", codeReadyWorkspacesCustomResource.Name)
		err = nil

	} else if (err != nil && !errors.IsAlreadyExists(err)) {
		return reconcile.Result{}, err
	}

	// Wait for CodeReadyWorkspace to be running
	if !kubernetes.GetK8Client().GetDeploymentStatus(DevSpacesDeployment, codeReadyWorkspacesNamespace.Name) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, nil
	}

/*	
	// Initialize Workspaces from devfile
	devfile, result, err := getDevFile(workshop)
	if err != nil {
		return result, err
	}
*/
	//no keycloak in DevSpaces

	// Users and Workspaces
	// NO keycloak option

	/*
		if !workshop.Spec.Infrastructure.CodeReadyWorkspace.OpenshiftOAuth {
			masterAccessToken, result, err := getKeycloakAdminToken(workshop, codeReadyWorkspacesNamespace.Name, appsHostnameSuffix)
			if err != nil {
				return result, err
			}

			labels := map[string]string{
				"app.kubernetes.io/part-of": "devspaces",
			}

			// Che Cluster Role
			cheClusterRole :=
				kubernetes.NewClusterRole(workshop, r.Scheme, "che", codeReadyWorkspacesNamespace.Name, labels, kubernetes.CheRules())
			if err := r.Create(context.TODO(), cheClusterRole); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			} else if err == nil {
				log.Infof("Created %s Cluster Role", cheClusterRole.Name)
			}

			cheClusterRoleBinding := kubernetes.NewClusterRoleBindingSA(workshop, r.Scheme, "che", codeReadyWorkspacesNamespace.Name, labels, "che", cheClusterRole.Name, "ClusterRole")
			if err := r.Create(context.TODO(), cheClusterRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
				return reconcile.Result{}, err
			} else if err == nil {
				log.Infof("Created %s Cluster Role Binding", cheClusterRoleBinding.Name)
			}

			for id := 1; id <= users; id++ {
				username := fmt.Sprintf("user%d", id)

				if result, err := createUser(workshop, username, "codeready", codeReadyWorkspacesNamespace.Name, appsHostnameSuffix, masterAccessToken); err != nil {
					return result, err
				}

				userAccessToken, result, err := getUserToken(workshop, username, "codeready", codeReadyWorkspacesNamespace.Name, appsHostnameSuffix)
				if err != nil {
					return result, err
				}

				if result, err := initWorkspace(workshop, username, "codeready", codeReadyWorkspacesNamespace.Name, userAccessToken, devfile, appsHostnameSuffix); err != nil {
					return result, err
				}

			}
		} else {
		/*	
		// loop through the users to try and activate their workspace

		for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)

		userAccessToken, result, err := getOAuthUserToken(workshop, username, CheURLCodeFlavour, codeReadyWorkspacesNamespace.Name, appsHostnameSuffix)
		if err != nil {
			return result, err
		}

		//			if result, err := updateUserEmail(workshop, username, CheURLCodeFlavour, codeReadyWorkspacesNamespace.Name, appsHostnameSuffix); err != nil {
		//				return result, err
		//			}

		if result, err := initWorkspace(workshop, username, CheURLCodeFlavour, codeReadyWorkspacesNamespace.Name, userAccessToken, devfile, appsHostnameSuffix); err != nil {
			return result, err
		}

	}
	*/
	//	}

	//Success
	return reconcile.Result{}, nil
}

func getDevFile(workshop *workshopv1.Workshop) (string, reconcile.Result, error) {

	var (
		httpResponse *http.Response
		httpRequest  *http.Request
		devfile      string
		client       = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	gitURL, err := url.Parse(workshop.Spec.Source.GitURL)
	if err != nil {
		return "", reconcile.Result{}, err
	}
	devfileRawURL := fmt.Sprintf("https://raw.githubusercontent.com%s/%s/devfile.yaml", gitURL.Path, workshop.Spec.Source.GitBranch)
	httpRequest, err = http.NewRequest("GET", devfileRawURL, nil)

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		log.Errorf("Error when getting Devfile from %s", devfileRawURL)
		return "", reconcile.Result{}, err
	}

	if httpResponse.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
		if err != nil {
			log.Errorf("Error when reading %s", devfileRawURL)
			return "", reconcile.Result{}, err
		}

		bodyJSON, err := yaml.YAMLToJSON(bodyBytes)
		if err != nil {
			log.Errorf("Error to converting %s to JSON", devfileRawURL)
			return "", reconcile.Result{}, err
		}
		devfile = string(bodyJSON)
	} else {
		log.Errorf("Error (%v) when getting Devfile from %s", httpResponse.StatusCode, devfileRawURL)
		return "", reconcile.Result{}, err
	}

	return devfile, reconcile.Result{}, nil
}

/*
func createUser(workshop *workshopv1.Workshop, username string, codeflavor string,
	namespace string, appsHostnameSuffix string, masterToken string) (reconcile.Result, error) {

	var (
		openshiftUserPassword = workshop.Spec.User.Password
		body                  []byte
		err                   error
		httpResponse          *http.Response
		httpRequest           *http.Request
		keycloakCheUserURL    = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/admin/realms/" + codeflavor + "/users"

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	body, err = json.Marshal(codeready.NewUser(username, openshiftUserPassword))
	if err != nil {
		return reconcile.Result{}, err
	}

	httpRequest, err = http.NewRequest("POST", keycloakCheUserURL, bytes.NewBuffer(body))
	httpRequest.Header.Set("Authorization", "Bearer "+masterToken)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		return reconcile.Result{}, err
	}
	if httpResponse.StatusCode == http.StatusCreated {
		log.Infof("Created %s in OpenShift Dev Spaces", username)
	}

	return reconcile.Result{}, nil
}
*/

/*
func getUserToken(workshop *workshopv1.Workshop, username string, codeflavor string, namespace string, appsHostnameSuffix string) (string, reconcile.Result, error) {
	var (
		openshiftUserPassword = workshop.Spec.User.Password
		err                   error
		httpResponse          *http.Response
		httpRequest           *http.Request
		keycloakCheTokenURL   = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/realms/" + codeflavor + "/protocol/openid-connect/token"

		userToken util.Token
		client    = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	// Get User Access Token
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", openshiftUserPassword)
	data.Set("client_id", codeflavor+"-public")
	data.Set("grant_type", "password")

	httpRequest, err = http.NewRequest("POST", keycloakCheTokenURL, strings.NewReader(data.Encode()))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		log.Errorf("Error to get the user access  token from %s keycloak (%v)", codeflavor, err)
		return "", reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&userToken); err != nil {
			log.Errorf("Error to get the user access  token from %s keycloak (%v)", codeflavor, err)
			return "", reconcile.Result{}, err
		}
	} else {
		log.Errorf("Error to get the user access token from %s keycloak (%d)", codeflavor, httpResponse.StatusCode)
		return "", reconcile.Result{}, err
	}

	return userToken.AccessToken, reconcile.Result{}, nil
}
*/

func getOAuthUserToken(workshop *workshopv1.Workshop, username string,
	codeflavor string, namespace string, appsHostnameSuffix string) (string, reconcile.Result, error) {
	var (
		openshiftUserPassword = workshop.Spec.User.Password
		err                   error
		httpResponse          *http.Response
		httpRequest           *http.Request
		//		keycloakCheTokenURL   = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/realms/" + codeflavor + "/protocol/openid-connect/token"
		oauthOpenShiftURL = "https://oauth-openshift." + appsHostnameSuffix + "/oauth/authorize?client_id=openshift-challenging-client&response_type=token"

		//		userToken util.Token
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	// GET TOKEN
	httpRequest, err = http.NewRequest("GET", oauthOpenShiftURL, nil)
	httpRequest.Header.Set("Authorization", "Basic "+util.GetBasicAuth(username, openshiftUserPassword))
	httpRequest.Header.Set("X-CSRF-Token", "xxx")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		log.Errorf("Error when getting Token Exchange for %s: %v", username, err)
		return "", reconcile.Result{}, err
	}

	if httpResponse.StatusCode == http.StatusFound {
		locationURL, err := url.Parse(httpResponse.Header.Get("Location"))
		if err != nil {
			return "", reconcile.Result{}, err
		}

		regex := regexp.MustCompile("access_token=([^&]+)")
		subjectToken := regex.FindStringSubmatch(locationURL.Fragment)

		return subjectToken[1], reconcile.Result{}, nil
	}

	log.Errorf("Error parsing token for %s: %v", username, httpResponse.StatusCode)

	return "", reconcile.Result{}, err
}

/*
		// Get User Access Token
		data := url.Values{}
		data.Set("client_id", codeflavor+"-public")
		data.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
		data.Set("subject_token", subjectToken[1])
		data.Set("subject_issuer", "openshift-v4")
		data.Set("subject_token_type", "urn:ietf:params:oauth:token-type:access_token")

		httpRequest, err = http.NewRequest("POST", "FIXME", strings.NewReader(data.Encode()))
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		httpResponse, err = client.Do(httpRequest)
		if err != nil {
			log.Errorf("Error to get the oauth user access  token from %s keycloak (%v)", codeflavor, err)
			return "", reconcile.Result{}, err
		}
		defer httpResponse.Body.Close()
		if httpResponse.StatusCode == http.StatusOK {
			if err := json.NewDecoder(httpResponse.Body).Decode(&userToken); err != nil {
				log.Errorf("Error to get the oauth user access  token from %s keycloak (%v)", codeflavor, err)
				return "", reconcile.Result{}, err
			}
		} else {
			log.Errorf("Error to get the oauth user access token from %s keycloak (%d)", codeflavor, httpResponse.StatusCode)
			return "", reconcile.Result{}, err
		}
	} else {
		log.Errorf("Error when getting Token Exchange for %s (%d)", username, httpResponse.StatusCode)
		return "", reconcile.Result{}, err
	}

	return userToken.AccessToken, reconcile.Result{}, nil

}
*/

/*
func getKeycloakAdminToken(workshop *workshopv1.Workshop, namespace string, appsHostnameSuffix string) (string, reconcile.Result, error) {
	var (
		err                 error
		httpResponse        *http.Response
		httpRequest         *http.Request
		keycloakCheTokenURL = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"

		masterToken util.Token
		client      = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	// GET TOKEN
	httpRequest, err = http.NewRequest("POST", keycloakCheTokenURL, strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		return "", reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
			return "", reconcile.Result{}, err
		}
	}

	return masterToken.AccessToken, reconcile.Result{}, nil
}
*/

/*
func updateUserEmail(workshop *workshopv1.Workshop, username string,
	codeflavor string, namespace string, appsHostnameSuffix string) (reconcile.Result, error) {
	var (
		err                    error
		httpResponse           *http.Response
		httpRequest            *http.Request
		keycloakMasterTokenURL = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/realms/master/protocol/openid-connect/token"
		keycloakUserURL        = "https://keycloak-" + namespace + "." + appsHostnameSuffix + "/auth/admin/realms/" + codeflavor + "/users"
		masterToken            util.Token
		client                 = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		cheUser []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
	)

	// Get Keycloak Admin Token
	httpRequest, err = http.NewRequest("POST", keycloakMasterTokenURL, strings.NewReader("username=admin&password=admin&grant_type=password&client_id=admin-cli"))
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		log.Errorf("Error when getting the master token from %s keycloak (%v)", codeflavor, err)
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&masterToken); err != nil {
			log.Errorf("Error when reading the master token: %v", err)
			return reconcile.Result{}, err
		}
	} else {
		log.Errorf("Error when getting the master token from %s keycloak (%d)", codeflavor, httpResponse.StatusCode)
		return reconcile.Result{}, err
	}

	// GET USER
	httpRequest, err = http.NewRequest("GET", keycloakUserURL+"?username="+username, nil)
	httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		log.Errorf("Error when getting %s user: %v", username, err)
		return reconcile.Result{}, err
	}

	defer httpResponse.Body.Close()
	if httpResponse.StatusCode == http.StatusOK {
		if err := json.NewDecoder(httpResponse.Body).Decode(&cheUser); err != nil {
			log.Errorf("Error to get the user info (%v)", err)
			return reconcile.Result{}, err
		}

		if cheUser[0].Email == "" {
			httpRequest, err = http.NewRequest("PUT", keycloakUserURL+"/"+cheUser[0].ID,
				strings.NewReader(`{"email":"`+username+`@none.com"}`))
			httpRequest.Header.Set("Content-Type", "application/json")
			httpRequest.Header.Set("Authorization", "Bearer "+masterToken.AccessToken)
			httpResponse, err = client.Do(httpRequest)
			if err != nil {
				log.Errorf("Error when update email address for %s: %v", username, err)
				return reconcile.Result{}, err
			}
		}
	} else {
		log.Errorf("Error when getting %s user: %v", username, httpResponse.StatusCode)
		return reconcile.Result{}, err
	}

	//Success
	return reconcile.Result{}, nil
}
*/

func initWorkspace(workshop *workshopv1.Workshop, username string,
	codeflavor string, namespace string, userAccessToken string, devfile string,
	appsHostnameSuffix string) (reconcile.Result, error) {

	var (
		err                 error
		httpResponse        *http.Response
		httpRequest         *http.Request
		devfileWorkspaceURL = "https://" + codeflavor + "-" + namespace + "." + appsHostnameSuffix + "/api/workspace/devfile?start-after-create=true&namespace=" + username

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			// Do not follow Redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	)

	httpRequest, err = http.NewRequest("POST", devfileWorkspaceURL, strings.NewReader(devfile))
	httpRequest.Header.Set("Authorization", "Bearer "+userAccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")

	httpResponse, err = client.Do(httpRequest)
	if err != nil {
		log.Errorf("Error when creating the workspace for %s: %v", username, err)
		return reconcile.Result{}, err
	}
	defer httpResponse.Body.Close()

	//Success
	return reconcile.Result{}, nil

}
