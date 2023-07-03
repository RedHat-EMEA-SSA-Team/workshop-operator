package controllers

import (
	//	"bytes"
	"context"

	//	"encoding/json"
	"fmt"
	"net/url"

	//	"regexp"
	//	"strings"
	"time"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/codeready"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/kubernetes"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/log"

	workspaces "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
	dparse "github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var jsonCheCodeEclipse = v1.JSON{Raw: []byte(`"che-code.eclipse.org"`)}
var jsonFalse = v1.JSON{Raw: []byte(`false`)}
var jsonTrue = v1.JSON{Raw: []byte(`true`)}
var jsonCommon = v1.JSON{Raw: []byte(`"common"`)}
var jsonMain = v1.JSON{Raw: []byte(`"main"`)}

var cheEnvs = []workspaces.EnvVar{}

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
	err := r.Create(context.TODO(), codeReadyWorkspacesCustomResource)

	if err == nil {
		log.Infof("Created %s Custom Resource", codeReadyWorkspacesCustomResource.Name)

	} else if errors.ReasonForError(err) == "only one CheCluster is allowed" {
		// Now Dev Spaces only allows one instance of the Che CR at the moment, so reports an attempt to add an extra one
		// as forbidden 403 and not as "AlreadyExists". So we need to ignore that
		//		log.Infof("An instance of %s Custom Resource already exists", codeReadyWorkspacesCustomResource.Name)
		err = nil

	} else if err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	}

	// Wait for CodeReadyWorkspace to be running
	if !kubernetes.GetK8Client().GetDeploymentStatus(DevSpacesDeployment, codeReadyWorkspacesNamespace.Name) {
		return reconcile.Result{Requeue: true, RequeueAfter: time.Second * 1}, nil
	}

	// Initialize Workspaces from devfile
	devfile, result, err := getDevFileName(workshop)
	if err != nil {
		return result, err
	}

	devObj, result, err := getDevFileObj(workshop, devfile, appsHostnameSuffix)
	if err != nil {
		return result, err
	}

	// loop through the users to try and activate their workspace
	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)

		if result, err := r.initWorkspace(workshop, username, CheURLCodeFlavour, devfile, devObj, appsHostnameSuffix); err != nil {
			return result, err
		}

	}

	//Success
	return reconcile.Result{}, nil
}

func getDevFileName(workshop *workshopv1.Workshop) (string, reconcile.Result, error) {

	gitURL, err := url.Parse(workshop.Spec.Source.GitURL)
	if err != nil {
		return "", reconcile.Result{}, err
	}
	return fmt.Sprintf("https://raw.githubusercontent.com%s/%s/devfile.yaml", gitURL.Path, workshop.Spec.Source.GitBranch), reconcile.Result{}, nil
}

func getDevFileObj(workshop *workshopv1.Workshop, devfileURL string, appsHostnameSuffix string) (parser.DevfileObj, reconcile.Result, error) {

	d, err := dparse.ParseFromURLAndValidate(devfileURL)
	if err != nil {
		return d, reconcile.Result{}, err
	}

	// use supplied container (inside component)
	suppliedComponents, err := d.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return d, reconcile.Result{}, err
	}

	container := suppliedComponents[0].Container

//	container.Container.Command = []string{"/checode/entrypoint-volume.sh"}

	cheEnvs = []workspaces.EnvVar{
		{
			Name:  "CHE_DASHBOARD_URL",
			Value: "https://devspaces." + appsHostnameSuffix,
		},
		{
			Name:  "CHE_PLUGIN_REGISTRY_URL",
			Value: "https://devspaces." + appsHostnameSuffix + "/plugin-registry/v3",
		},
		{
			Name:  "CHE_PLUGIN_REGISTRY_INTERNAL_URL",
			Value: "http://plugin-registry.openshift-devspaces.svc:8080/v3",
		},
		{
			Name:  "CLUSTER_CONSOLE_URL",
			Value: "https://console-openshift-console." + appsHostnameSuffix,
		},
		{
			Name:  "CLUSTER_CONSOLE_TITLE",
			Value: "OpenShift console",
		},
		{
			Name:  "OPENVSX_REGISTRY_URL",
			Value: "",
		},
	}

	// extend the devfile enVar array with the builtin Che values
	container.Env = append(container.Env, cheEnvs[0], cheEnvs[1], cheEnvs[2], cheEnvs[3], cheEnvs[4], cheEnvs[5])

	return d, reconcile.Result{}, nil
}

func (r *WorkshopReconciler) initWorkspace(workshop *workshopv1.Workshop, username string,
	codeflavor string, devfile string, devObj parser.DevfileObj, appsHostnameSuffix string) (reconcile.Result, error) {

	const userNameAppend = "-devspaces"
	const settingsCMName = "settings-xml"
	const gitconfigCMName = "gitconfig"

	// Create namespace with dev workspace annotations
	labels := map[string]string{
		"app.kubernetes.io/part-of":   "che.eclipse.org",
		"app.kubernetes.io/component": "workspaces-namespace",
	}

	annotations := map[string]string{
		"che.eclipse.org/username": username,
		"openshift.io/requester":   "system:serviceaccount:openshift-devspaces:che",
	}

	userDevSpace := kubernetes.NewNamespaceAnnotate(workshop, r.Scheme, username+userNameAppend, labels, annotations)
	if err := r.Create(context.TODO(), userDevSpace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created user DevSpace %s Project", userDevSpace.Name)
	}

	// Create ConfigMap with dev workspace annoations
	labels = map[string]string{
		"controller.devfile.io/mount-to-devworkspace": "true",
		"controller.devfile.io/watch-configmap":       "true",
	}

	annotations = map[string]string{
		"controller.devfile.io/mount-as":   "subpath",
		"controller.devfile.io/mount-path": "/home/developer/.m2",
	}

	// pass in a maven settings.xml file to be mounted
	data := map[string]string{
		"settings.xml": `<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 https://maven.apache.org/xsd/settings-1.0.0.xsd">
	<localRepository/>
	<interactiveMode/>
	<offline/>
	<pluginGroups/>
	<servers/>
	<mirrors>
	<mirror>
		<url>${env.MAVEN_MIRROR_URL}</url>
		<mirrorOf>external:*</mirrorOf>
	</mirror>
	</mirrors>
	<proxies/>
	<profiles/>
	<activeProfiles/>
</settings>`,
	}

	// Create settings configmap inside
	settingsCM := kubernetes.NewConfigMapAnnotate(workshop, r.Scheme, settingsCMName, username+userNameAppend, labels, data, annotations)
	if err := r.Create(context.TODO(), settingsCM); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created Settings.xml config map for user %s", username)
	}

	// Create ConfigMap with dev workspace annoations
	labels = map[string]string{
		"controller.devfile.io/mount-to-devworkspace": "true",
		"controller.devfile.io/watch-configmap":       "true",
	}

	annotations = map[string]string{
		"controller.devfile.io/mount-as":   "subpath",
		"controller.devfile.io/mount-path": "/home/developer/",
	}

	// pass in a gitcofngi file to be mounted
	data = map[string]string{
		".gitconfig": "[user]\n  name = " + username + "\n  email = " + username + "@example.com\n",
	}

	// Create get config map inside
	gitconfigCM := kubernetes.NewConfigMapAnnotate(workshop, r.Scheme, gitconfigCMName, username+userNameAppend, labels, data, annotations)
	if err := r.Create(context.TODO(), gitconfigCM); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created Gitconfig config map for user %s", username)
	}

	// Create DevWorkspace Template
	dwtemp := NewDWTemplate(workshop, r.Scheme, username+userNameAppend, appsHostnameSuffix)
	if err := r.Create(context.TODO(), dwtemp); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created DWTemplate for user %s", username)
	}

	// Create DevWorkspace (DW)
	dwwork := NewDevWorkspace(workshop, r.Scheme, username+userNameAppend, devfile, devObj)
	if err := r.Create(context.TODO(), dwwork); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created DevWorkspaces for user %s", username)
	}

	//Success
	return reconcile.Result{}, nil

}

// NewDWTemplate
func NewDWTemplate(workshop *workshopv1.Workshop, scheme *runtime.Scheme, namespace string, appsHostnameSuffix string) *workspaces.DevWorkspaceTemplate {

	commands := []workspaces.Command{
		{
			Id: "init-container-command",
			CommandUnion: workspaces.CommandUnion{
				Apply: &workspaces.ApplyCommand{
					Component: "che-code-injector",
				},
			},
		},
		{
			Id: "init-che-code-command",
			CommandUnion: workspaces.CommandUnion{
				Exec: &workspaces.ExecCommand{
					CommandLine: "nohup /checode/entrypoint-volume.sh > /checode/entrypoint-logs.txt 2>&1 &",
					Component: "che-code-runtime-description",
				},				
			},
		},
	}
	secure := false

	// append endpoints to the existing container values
	endpoints := []workspaces.Endpoint{
		{
			Name:       "che-code",
			Exposure:   workspaces.PublicEndpointExposure,
			Protocol:   workspaces.HTTPSEndpointProtocol,
			TargetPort: 3100,
			Secure:     &secure,
			Attributes: attributes.Attributes{
				"discoverable":        jsonFalse,
				"urlRewriteSupported": jsonTrue,
				"type":                jsonMain,
				"cookiesAuthEnabled":  jsonTrue,
			},
		},
		{
			Name:       "code-redirect-1",
			Exposure:   workspaces.PublicEndpointExposure,
			Protocol:   workspaces.HTTPEndpointProtocol,
			TargetPort: 13131,
			Attributes: attributes.Attributes{
				"discoverable":        jsonFalse,
				"urlRewriteSupported": jsonFalse,
			},
		},
		{
			Name:       "code-redirect-2",
			Exposure:   workspaces.PublicEndpointExposure,
			Protocol:   workspaces.HTTPEndpointProtocol,
			TargetPort: 13132,
			Attributes: attributes.Attributes{
				"discoverable":        jsonFalse,
				"urlRewriteSupported": jsonFalse,
			},
		},
		{
			Name:       "code-redirect-3",
			Exposure:   workspaces.PublicEndpointExposure,
			Protocol:   workspaces.HTTPEndpointProtocol,
			TargetPort: 13133,
			Attributes: attributes.Attributes{
				"discoverable":        jsonFalse,
				"urlRewriteSupported": jsonFalse,
			},
		},
	}

	containerRuntime := &workspaces.ContainerComponent{
		Container: workspaces.Container{
			CpuRequest:    "30m",
			Env:           cheEnvs,
			MemoryRequest: "256Mi",
			SourceMapping: "/projects",
			CpuLimit:      "500m",
			VolumeMounts: []workspaces.VolumeMount{
				{
					Name: "checode",
					Path: "/checode",
				},
			},
			MemoryLimit: "1024Mi",
			Image:       "registry.redhat.io/devspaces/udi-rhel8",
		},
		Endpoints: endpoints,
	}

	containerInjector := &workspaces.ContainerComponent{
		Container: workspaces.Container{
			CpuRequest:    "30m",
			Command:       []string{"/entrypoint-init-container.sh"},
			Env:           cheEnvs,
			MemoryRequest: "32Mi",
			SourceMapping: "/projects",
			CpuLimit:      "500m",
			VolumeMounts: []workspaces.VolumeMount{
				{
					Name: "checode",
					Path: "/checode",
				},
			},
			MemoryLimit: "256Mi",
			Image:       "registry.redhat.io/devspaces/code-rhel8",
		},
	}

	components := []workspaces.Component{
		{
			Name: "che-code-runtime-description",
			ComponentUnion: workspaces.ComponentUnion{
				Container: containerRuntime,
			},
			Attributes: attributes.Attributes{
				"app.kubernetes.io/component": v1.JSON{Raw: []byte(`"che-code-runtime"`)},
				"app.kubernetes.io/part-of": v1.JSON{Raw: []byte(`"che-code.eclipse.org"`)},
				"controller.devfile.io/container-contribution": jsonTrue,
				},
		},
		{
			Name: "checode",
			ComponentUnion: workspaces.ComponentUnion{
				Volume: &workspaces.VolumeComponent{
					Volume: workspaces.Volume{},
				},
			},
		},
		{
			Name: "che-code-injector",
			ComponentUnion: workspaces.ComponentUnion{
				Container: containerInjector,
			},
		},
	}

	template := &workspaces.DevWorkspaceTemplate{

		ObjectMeta: metav1.ObjectMeta{
			Name:      "che-code-workspace",
			Namespace: namespace,
		},

		TypeMeta: metav1.TypeMeta{
			Kind:       "DevWorkspaceTemplate",
			APIVersion: "workspace.devfile.io/v1alpha2",
		},

		Spec: workspaces.DevWorkspaceTemplateSpec{
			DevWorkspaceTemplateSpecContent: workspaces.DevWorkspaceTemplateSpecContent{
				Commands:   commands,
				Components: components,
				Events: &workspaces.Events{
					DevWorkspaceEvents: workspaces.DevWorkspaceEvents{
						PreStart: []string{"init-container-command"},
						PostStart: []string{"init-che-code-command"},
					},
				},
			},
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, template, scheme)

	return template
}

// NewDevWorkspace
func NewDevWorkspace(workshop *workshopv1.Workshop, scheme *runtime.Scheme, namespace string,
	devfile string, devObj parser.DevfileObj) *workspaces.DevWorkspace {

	// use supplied commands
	commands, err := devObj.Data.GetCommands(common.DevfileOptions{})
	if err != nil {
		return nil
	}

	// use supplied projects
	projects, err := devObj.Data.GetProjects(common.DevfileOptions{})
	if err != nil {
		return nil
	}

	// use supplied container (inside component)
	suppliedComponents, err := devObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return nil
	}

	container := suppliedComponents[0].Container

	components := []workspaces.Component{
		{
			Name: "workshop-tools",
			ComponentUnion: workspaces.ComponentUnion{
				Container: container,
			},
			Attributes: attributes.Attributes{
				"controller.devfile.io/merge-contribution": jsonTrue,
			},
		},
		{
			Name: "m2",
			ComponentUnion: workspaces.ComponentUnion{
				Volume: &workspaces.VolumeComponent{
					Volume: workspaces.Volume{
						Size: "10G",
					},
				},
			},
		},

		/*
		{
			Name: "che-code-workspace",
			ComponentUnion: workspaces.ComponentUnion{
				Plugin: &workspaces.PluginComponent{
					ImportReference: workspaces.ImportReference{
						ImportReferenceUnion: workspaces.ImportReferenceUnion{
							Kubernetes: &workspaces.KubernetesCustomResourceImportReference{
								Name:      "che-code-workspace",
								Namespace: namespace,
							},
						},
					},
				},
			},
		},*/
	}

	annotations := map[string]string{
		"che.eclipse.org/che-editor":     "che-code-workspace",
	}

	componentContributions := []workspaces.ComponentContribution {
		{
			Name: "editor",
			PluginComponent: workspaces.PluginComponent{
				ImportReference: workspaces.ImportReference{
					ImportReferenceUnion: workspaces.ImportReferenceUnion{
						Kubernetes: &workspaces.KubernetesCustomResourceImportReference{
							Name:      "che-code-workspace",
						},
					},
				},
			},

		},
	};

	workspace := &workspaces.DevWorkspace{

		ObjectMeta: metav1.ObjectMeta{
			Name:        "wksp-end-to-end-dev",
			Namespace:   namespace,
			Annotations: annotations,
		},

		TypeMeta: metav1.TypeMeta{
			Kind:       "DevWorkspace",
			APIVersion: "workspace.devfile.io/v1alpha2",
		},

		Spec: workspaces.DevWorkspaceSpec{
			Contributions: componentContributions,
			Started:      true,
			RoutingClass: "che",
			Template: workspaces.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: workspaces.DevWorkspaceTemplateSpecContent{
					Attributes: attributes.Attributes{
						"controller.devfile.io/devworkspace-config": v1.JSON{Raw: []byte(`{
							"name": "devworkspace-config",
							"namespace": "openshift-devspaces"
						}`)},

//						"controller.devfile.io/scc": v1.JSON{Raw: []byte(`"container-build"`)},

						"controller.devfile.io/storage-type": jsonCommon,
						"dw.metadata.annotations": v1.JSON{Raw: []byte(`{
							"che.eclipse.org/devfile-source": "url:\n location: \u003e-\n    ` + devfile + `\nfactory:\n  params: \u003e-\n    url=` + devfile + `\n"
						}`)},
					},
					Commands:   commands,
					Components: components,
					Projects:   projects,
				},
			},
		},
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, workspace, scheme)

	return workspace
}
