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
	"github.com/prometheus/common/log"

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

var jsonCheCodeEclipse = v1.JSON {Raw: []byte(`"che-code.eclipse.org"`) }
var jsonFalse = v1.JSON {Raw: []byte(`false`) }
var jsonTrue = v1.JSON {Raw: []byte(`true`) }
var jsonCommon = v1.JSON {Raw: []byte(`"common"`) }
var jsonMain = v1.JSON {Raw: []byte(`"main"`) }

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

	container.Container.Command = []string {"/checode/entrypoint-volume.sh"}
	container.Container.VolumeMounts = append(container.Container.VolumeMounts, workspaces.VolumeMount {
		Name: "checode",
		Path: "/checode",
		})

	cheEnvs = []workspaces.EnvVar {
			{
				Name: "CHE_DASHBOARD_URL",
				Value: "https://devspaces."+appsHostnameSuffix,
			},
			{
				Name: "CHE_PLUGIN_REGISTRY_URL",
				Value: "https://devspaces."+appsHostnameSuffix+"/plugin-registry/v3",
			},
			{
				Name: "CHE_PLUGIN_REGISTRY_INTERNAL_URL",
				Value: "http://plugin-registry.openshift-devspaces.svc:8080/v3",
			},
			{
				Name: "OPENVSX_REGISTRY_URL",
				Value: "https://open-vsx.org",
			},
		}

	// extend the devfile enVar array with the builtin Che values
	container.Env = append(container.Env, cheEnvs[0], cheEnvs[1], cheEnvs[2], cheEnvs[3])
	
	secure := false

	// append endpoints to the existing container values
	endpoints := []workspaces.Endpoint{
		{
			Name: "che-code",
			Exposure: workspaces.PublicEndpointExposure,
			Protocol: workspaces.HTTPSEndpointProtocol,
			TargetPort: 3100,
			Path: "?tkn=eclipse-che",
			Secure: &secure,
			Attributes: attributes.Attributes{
				"contributed-by": jsonCheCodeEclipse,
				"discoverable": jsonFalse,
				"urlRewriteSupported": jsonTrue,
				"type": jsonMain,
				"cookiesAuthEnabled": jsonTrue,
				},
		},
		{
			Name: "code-redirect-1",
			Exposure: workspaces.PublicEndpointExposure,
			Protocol: workspaces.HTTPEndpointProtocol,
			TargetPort: 13131,
			Attributes: attributes.Attributes{
				"contributed-by": jsonCheCodeEclipse,
				"discoverable": jsonFalse,
				"urlRewriteSupported": jsonTrue,
				},
		},
		{
			Name: "code-redirect-2",
			Exposure: workspaces.PublicEndpointExposure,
			Protocol: workspaces.HTTPEndpointProtocol,
			TargetPort: 13132,
			Attributes: attributes.Attributes{
				"contributed-by": jsonCheCodeEclipse,
				"discoverable": jsonFalse,
				"urlRewriteSupported": jsonTrue,
				},
		},
		{
			Name: "code-redirect-3",
			Exposure: workspaces.PublicEndpointExposure,
			Protocol: workspaces.HTTPEndpointProtocol,
			TargetPort: 13133,
			Attributes: attributes.Attributes{
				"contributed-by": jsonCheCodeEclipse,
				"discoverable": jsonFalse,
				"urlRewriteSupported": jsonTrue,
				},
		},
	}
	container.Endpoints = append(container.Endpoints, endpoints[0], endpoints[1], endpoints[2], endpoints[3])	

	return d, reconcile.Result{}, nil;
}

func (r *WorkshopReconciler) initWorkspace(workshop *workshopv1.Workshop, username string,
	codeflavor string, devfile string, devObj parser.DevfileObj, appsHostnameSuffix string) (reconcile.Result, error) {

	const userNameAppend = "-devspaces"
	const settingsCMName = "settings-xml"

	// Create namespace with dev workspace annotations
	labels := map[string]string{
		"app.kubernetes.io/part-of": "che.eclipse.org",
		"app.kubernetes.io/component": "workspaces-namespace",
		}

	annotations := map[string]string{
		"che.eclipse.org/username": username,
		"openshift.io/requester": "system:serviceaccount:openshift-devspaces:che",
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
		"controller.devfile.io/watch-configmap": "true",
	}

	annotations = map[string]string{
		"controller.devfile.io/mount-as": "subpath",
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


	// Create settings secret inside
	settingsCM := kubernetes.NewConfigMapAnnotate(workshop, r.Scheme, settingsCMName, username+userNameAppend, labels, data, annotations)
	if err := r.Create(context.TODO(), settingsCM); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created Settings.xml config map for user %s", username)
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

	commands := []workspaces.Command {
			{
			Id: "init-container-command",
			CommandUnion: workspaces.CommandUnion {
				Apply:  &workspaces.ApplyCommand{
					Component: "che-code-injector",
				},
			},
		},
	}
	
	container := &workspaces.ContainerComponent{
		Container: workspaces.Container{
			CpuRequest: "30m",
			Command: []string {"/entrypoint-init-container.sh"},
			Env: cheEnvs,
			MemoryRequest: "32Mi",
			SourceMapping: "/projects",
			CpuLimit: "500m",
			VolumeMounts: []workspaces.VolumeMount {
				{
				Name: "checode",
				Path: "/checode",
				},
			},
			MemoryLimit: "128Mi",
			Image: "registry.redhat.io/devspaces/code-rhel8",
		},
	}

	components := []workspaces.Component {
		{
			Name: "checode",
			ComponentUnion: workspaces.ComponentUnion{
				Volume: &workspaces.VolumeComponent {
					Volume: workspaces.Volume{
					},
				},
			},
		},
		{
			Name: "che-code-injector",
			ComponentUnion: workspaces.ComponentUnion{
				Container: container,
			},
		},
	}

	template := &workspaces.DevWorkspaceTemplate{

		ObjectMeta: metav1.ObjectMeta{
			Name: "che-code-workspace",
			Namespace: namespace,
		},

		TypeMeta: metav1.TypeMeta{
			Kind: "DevWorkspaceTemplate",
			APIVersion: "workspace.devfile.io/v1alpha2",
		},
		
		Spec: workspaces.DevWorkspaceTemplateSpec{
			DevWorkspaceTemplateSpecContent: workspaces.DevWorkspaceTemplateSpecContent {
				Commands:  commands,
				Components: components,
				Events: &workspaces.Events{
					DevWorkspaceEvents: workspaces.DevWorkspaceEvents{
						PreStart: []string {"init-container-command"},
					},
				},
			},
		},
	}


/*	
	apiVersion: workspace.devfile.io/v1alpha2
	kind: DevWorkspaceTemplate
	metadata:
	  name: che-code-workspace
	  namespace: user2-devspaces
	spec:
	  commands:
		- apply:
			component: che-code-injector
		  id: init-container-command
	  components:
		- name: checode
		  volume: {}
		- container:
			cpuRequest: 30m
			command:
			  - /entrypoint-init-container.sh
			env:
			  - name: CHE_DASHBOARD_URL
	#            value: 'https://devspaces.apps.<hostprefix>'
				value: 'https://devspaces.apps.cluster-48rld.48rld.sandbox388.opentlc.com'
			  - name: CHE_PLUGIN_REGISTRY_URL
	#            value: 'https://devspaces.apps.<hostprefix>/plugin-registry/v3'
				value: >-
				  https://devspaces.apps.cluster-48rld.48rld.sandbox388.opentlc.com/plugin-registry/v3
			  - name: CHE_PLUGIN_REGISTRY_INTERNAL_URL
				value: 'http://plugin-registry.openshift-devspaces.svc:8080/v3'
			  - name: OPENVSX_REGISTRY_URL
				value: 'https://open-vsx.org'
			memoryRequest: 32Mi
			sourceMapping: /projects
			cpuLimit: 500m
			volumeMounts:
			  - name: checode
				path: /checode
			memoryLimit: 128Mi
			image: >-
			  registry.redhat.io/devspaces/code-rhel8
		  name: che-code-injector
	  events:
		preStart:
		  - init-container-command
*/	

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

	components := []workspaces.Component {
		{
			Name: "workshop-tools",
			ComponentUnion: workspaces.ComponentUnion{
				Container: container,
			},
			Attributes: attributes.Attributes{
				"che-code.eclipse.org/contribute-cpuLimit": jsonTrue,
				"che-code.eclipse.org/contribute-cpuRequest": jsonTrue,
				"che-code.eclipse.org/contribute-endpoint/che-code": v1.JSON{Raw: []byte(`3100`) },
				"che-code.eclipse.org/contribute-endpoint/code-redirect-1": v1.JSON{Raw: []byte(`13131`) },
				"che-code.eclipse.org/contribute-endpoint/code-redirect-2": v1.JSON{Raw: []byte(`13132`) },
				"che-code.eclipse.org/contribute-endpoint/code-redirect-3": v1.JSON{Raw: []byte(`13133`) },
				"che-code.eclipse.org/contribute-entry-point": jsonTrue,
				"che-code.eclipse.org/contribute-memoryLimit": jsonTrue,
				"che-code.eclipse.org/contribute-memoryRequest": jsonTrue,
				"che-code.eclipse.org/contribute-volume-mount/checode": v1.JSON{Raw: []byte(`"/checode"`) },
				"che-code.eclipse.org/contributed-container": v1.JSON{Raw: []byte(`"workshop-tools"`) },
				"che-code.eclipse.org/original-cpuLimit": v1.JSON{Raw: []byte(`"1000m"`) },
				"che-code.eclipse.org/original-cpuRequest": v1.JSON{Raw: []byte(`"50m"`) },
				"che-code.eclipse.org/original-memoryLimit": v1.JSON{Raw: []byte(`"2048Mi"`) },
				"che-code.eclipse.org/original-memoryRequest": v1.JSON{Raw: []byte(`"256Mi"`) },
			},
		},
		{
			Name: "m2",
			ComponentUnion: workspaces.ComponentUnion{
				Volume: &workspaces.VolumeComponent {
					Volume: workspaces.Volume{
						Size: "1G",
					},
				},
			},
		},

		{
			Name: "che-code-workspace",
			ComponentUnion: workspaces.ComponentUnion{
				Plugin: &workspaces.PluginComponent{
					ImportReference: workspaces.ImportReference{
						ImportReferenceUnion: workspaces.ImportReferenceUnion{
							Kubernetes: &workspaces.KubernetesCustomResourceImportReference{
								Name: "che-code-workspace",
								Namespace: namespace,
							},
						},
					},
				},
			},
		},
	}

	annotations := map[string]string {
		"che.eclipse.org/che-editor": "che-incubator/che-code/insiders",
		"che.eclipse.org/devfile-source" : `"url:\n location: \u003e-\n    ` + devfile + `\nfactory:\n  params: \u003e-\n    url=` + devfile + `\n"`,
	}

	workspace := &workspaces.DevWorkspace{

		ObjectMeta: metav1.ObjectMeta{
			Name: "wksp-end-to-end-dev",
			Namespace: namespace,
			Annotations: annotations,
		},

		TypeMeta: metav1.TypeMeta{
			Kind: "DevWorkspace",
			APIVersion: "workspace.devfile.io/v1alpha2",
		},
		
		Spec: workspaces.DevWorkspaceSpec{
			Started: true,
			RoutingClass: "che",
			Template: workspaces.DevWorkspaceTemplateSpec{
				DevWorkspaceTemplateSpecContent: workspaces.DevWorkspaceTemplateSpecContent{
					Attributes: attributes.Attributes{
						"controller.devfile.io/devworkspace-config": v1.JSON{Raw: []byte(`{
							"name": "devworkspace-config",
							"namespace": "openshift-devspaces"
						}`) },
		
						"controller.devfile.io/storage-type": jsonCommon,
						"dw.metadata.annotations": v1.JSON{Raw: []byte(`{
							"che.eclipse.org/che-editor": "che-incubator/che-code/insiders",
							"che.eclipse.org/devfile-source": "url:\n location: \u003e-\n    ` + devfile + `\nfactory:\n  params: \u003e-\n    url=` + devfile + `\n"
						}`) },
					}, 
					Commands: commands,
					Components: components,
					Projects: projects,
				},
			},
		}, 
	}

	// Set Workshop instance as the owner and controller
	ctrl.SetControllerReference(workshop, workspace, scheme)

	return workspace
}

/*
apiVersion: workspace.devfile.io/v1alpha2
kind: DevWorkspace
metadata:
  annotations:
    che.eclipse.org/che-editor: che-incubator/che-code/insiders
    che.eclipse.org/devfile-source: |
      url:
        location: >-
          https://raw.githubusercontent.com/RedHat-EMEA-SSA-Team/end-to-end-developer-workshop/6.4/devfile.yaml
      factory:
        params: >-
          url=https://raw.githubusercontent.com/RedHat-EMEA-SSA-Team/end-to-end-developer-workshop/6.4/devfile.yaml
  name: wksp-end-to-end-dev
#  namespace: user1-devspaces
  namespace: user2-devspaces
  finalizers:
    - storage.controller.devfile.io
spec:
  routingClass: che
  started: true
  template:
    attributes:
      controller.devfile.io/devworkspace-config:
        name: devworkspace-config
        namespace: openshift-devspaces
      controller.devfile.io/storage-type: common
      dw.metadata.annotations:
        che.eclipse.org/che-editor: che-incubator/che-code/insiders
        che.eclipse.org/devfile-source: |
          url:
            location: >-
              https://raw.githubusercontent.com/RedHat-EMEA-SSA-Team/end-to-end-developer-workshop/6.4/devfile.yaml
          factory:
            params: >-
              url=https://raw.githubusercontent.com/RedHat-EMEA-SSA-Team/end-to-end-developer-workshop/6.4/devfile.yaml
    commands:
      - exec:
          commandLine: >-
            odo login $(oc whoami --show-server)
            --username=${DEVWORKSPACE_NAMESPACE%-devspaces} --password=openshift
            --insecure-skip-tls-verify
          component: workshop-tools
          label: OpenShift - Login
          workingDir: /projects/workshop
        id: openshift---login
      - exec:
          commandLine: >-
            odo project create
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: OpenShift - Create Development Project
          workingDir: /projects/workshop
        id: openshift---create-development-project
      - exec:
          commandLine: >-
            [[ ! -z "$(ps aux | grep -v grep | grep "compile quarkus:dev" | awk
            '{print $2}')" ]] &&  echo '!! Application already running in Dev
            Mode !!' ||  mvn compile quarkus:dev -Ddebug=false
          component: workshop-tools
          label: Inventory - Compile (Dev Mode)
          workingDir: /projects/workshop/labs/inventory-quarkus
        id: inventory---compile-dev-mode
      - exec:
          commandLine: mvn clean package -DskipTests
          component: workshop-tools
          label: Inventory - Build
          workingDir: /projects/workshop/labs/inventory-quarkus
        id: inventory---build
      - exec:
          commandLine: >-
            odo create --app coolstore --project
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: Inventory - Create Component
          workingDir: /projects/workshop/labs/inventory-quarkus
        id: inventory---create-component
      - exec:
          commandLine: odo push
          component: workshop-tools
          label: Inventory - Push
          workingDir: /projects/workshop/labs/inventory-quarkus
        id: inventory---push
      - exec:
          commandLine: mvn clean package -DskipTests
          component: workshop-tools
          label: Catalog - Build
          workingDir: /projects/workshop/labs/catalog-spring-boot
        id: catalog---build
      - exec:
          commandLine: 'mvn spring-boot:run'
          component: workshop-tools
          label: Catalog - Run
          workingDir: /projects/workshop/labs/catalog-spring-boot
        id: catalog---run
      - exec:
          commandLine: >-
            odo create --app coolstore --project
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: Catalog - Create Component
          workingDir: /projects/workshop/labs/catalog-spring-boot
        id: catalog---create-component
      - exec:
          commandLine: odo push
          component: workshop-tools
          label: Catalog - Push
          workingDir: /projects/workshop/labs/catalog-spring-boot
        id: catalog---push
      - exec:
          commandLine: >-
            odo create --app coolstore --project
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: Gateway - Create Component
          workingDir: /projects/workshop/labs/gateway-dotnet
        id: gateway---create-component
      - exec:
          commandLine: odo push
          component: workshop-tools
          label: Gateway - Push
          workingDir: /projects/workshop/labs/gateway-dotnet
        id: gateway---push
      - exec:
          commandLine: >-
            for i in {1..60}; do if [ $(curl -s -w "%{http_code}" -o /dev/null
            http://catalog-coolstore.my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}.svc:8080/actuator/health)
            == "200" ]; then MSG="\\033[0;32mThe request to Catalog Service has
            succeeded\\033[0m"; else MSG="\\033[0;31mERROR - The request to
            Catalog Service has failed\\033[0m"; fi;echo -e $MSG;sleep 2s; done
          component: workshop-tools
          label: Catalog - Generate Traffic
          workingDir: /projects/workshop/labs/catalog-spring-boot
        id: catalog---generate-traffic
      - exec:
          commandLine: >-
            oc patch deployment/catalog-coolstore -n
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
            --patch '{"spec": {"template": {"spec": {"affinity": {"podAffinity":
            {"requiredDuringSchedulingIgnoredDuringExecution":
            [{"labelSelector": { "matchExpressions": [{"key" : "component",
            "operator" : "In", "values": ["catalog"]}]}, "topologyKey" :
            "kubernetes.io/hostname"}]}}}'
          component: workshop-tools
          label: Catalog - Add PodAffinity
          workingDir: /projects/workshop/labs/catalog-spring-boot
        id: catalog---add-podaffinity
      - exec:
          commandLine: >-
            oc project
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14};
            oc set probe deployment/gateway-coolstore  --liveness --readiness
            --period-seconds=5 --get-url=http://:8080/health;oc set probe
            deployment/web-coolstore  --liveness --readiness --period-seconds=5
            --get-url=http://:8080/;echo "Health Probes Done"
          component: workshop-tools
          label: Probes - Configure Gateway & Web
          workingDir: /projects/workshop
        id: probes---configure-gateway--web
      - exec:
          commandLine: >-
            ./gateway_generate_traffic.sh
            cn-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: Gateway - Generate Traffic
          workingDir: /projects/workshop/.tasks
        id: gateway---generate-traffic
      - exec:
          commandLine: >-
            ./inner_loop_deploy_coolstore.sh
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: Inner Loop - Deploy Coolstore
          workingDir: /projects/workshop/.tasks
        id: inner-loop---deploy-coolstore
      - exec:
          commandLine: >-
            git init; git remote add origin
            http://gitea-server.gitea.svc:3000/${DEVWORKSPACE_NAMESPACE%-devspaces}/inventory-quarkus.git;
            git add *; git commit -m "Initial"; git push
            http://${DEVWORKSPACE_NAMESPACE%-devspaces}:openshift@gitea-server.gitea.svc:3000/${DEVWORKSPACE_NAMESPACE%-devspaces}/inventory-quarkus.git
          component: workshop-tools
          label: Inventory - Commit
          workingDir: /projects/workshop/labs/inventory-quarkus
        id: inventory---commit
      - exec:
          commandLine: >-
            ./gitops_export_coolstore.sh
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
            cn-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: GitOps - Export Coolstore
          workingDir: /projects/workshop/.tasks
        id: gitops---export-coolstore
      - exec:
          commandLine: >-
            oc project
            cn-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
            && ./pipeline_deploy_coolstore.sh
            cn-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: Pipeline - Deploy Coolstore
          workingDir: /projects/workshop/.tasks
        id: pipeline---deploy-coolstore
      - exec:
          commandLine: >-
            git init; git remote add origin
            http://gitea-server.gitea.svc:3000/${DEVWORKSPACE_NAMESPACE%-devspaces}/inventory-gitops.git
            2> /dev/null; git add *; git commit -m "Initial Inventory GitOps";
            git push
            http://${DEVWORKSPACE_NAMESPACE%-devspaces}:openshift@gitea-server.gitea.svc:3000/${DEVWORKSPACE_NAMESPACE%-devspaces}/inventory-gitops.git
          component: workshop-tools
          label: GitOps - Commit Inventory
          workingDir: /projects/workshop/labs/gitops/inventory-coolstore
        id: gitops---commit-inventory
      - exec:
          commandLine: >-
            ./gitops_commit_configure_coolstore.sh
            ${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: GitOps - Commit & Configure Coolstore
          workingDir: /projects/workshop/.tasks
        id: gitops---commit--configure-coolstore
      - exec:
          commandLine: >-
            oc patch deployment/catalog-coolstore --patch '{"spec": {"template":
            {"metadata": {"annotations": {"sidecar.istio.io/inject": "true"}}}'
            -n
            cn-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
            && oc patch deployment/gateway-coolstore --patch '{"spec":
            {"template": {"metadata": {"annotations":
            {"sidecar.istio.io/inject": "true"}}}' -n
            cn-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14} 
          component: workshop-tools
          label: Service Mesh - Deploy Catalog and Gateway
        id: service-mesh---deploy-catalog-and-gateway
      - exec:
          commandLine: >-
            git checkout .; git clean -fd; git clean -f; oc delete project
            my-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14};
            oc delete
            deployment,deploymentconfig,buildconfig,imagestream,route,secret,configmap,pvc,service,pipeline,pipelinerun
            --all --namespace
            cn-project${DEVWORKSPACE_NAMESPACE:4:${#DEVWORKSPACE_NAMESPACE}-14}
          component: workshop-tools
          label: OpenShift - Cleanup
          workingDir: /projects/workshop
        id: openshift---cleanup
    components:
      - attributes:
          che-code.eclipse.org/contribute-endpoint/code-redirect-1: 13131
          che-code.eclipse.org/contribute-memoryLimit: true
          che-code.eclipse.org/contribute-endpoint/code-redirect-2: 13132
          che-code.eclipse.org/contribute-cpuRequest: true
          che-code.eclipse.org/contribute-endpoint/code-redirect-3: 13133
          che-code.eclipse.org/original-memoryLimit: 2048Mi
          che-code.eclipse.org/contributed-container: workshop-tools
          che-code.eclipse.org/original-cpuLimit: 1000m
          che-code.eclipse.org/contribute-cpuLimit: true
          che-code.eclipse.org/contribute-memoryRequest: true
          che-code.eclipse.org/original-memoryRequest: 256Mi
          che-code.eclipse.org/contribute-endpoint/che-code: 3100
          che-code.eclipse.org/contribute-entry-point: true
          che-code.eclipse.org/original-cpuRequest: 50m
          che-code.eclipse.org/contribute-volume-mount/checode: /checode
          che-code.eclipse.org/original-entry-point:
            - /home/developer/entrypoint.sh
        container:
          cpuRequest: 80m
          command:
            - /checode/entrypoint-volume.sh
          env:
            - name: MAVEN_OPTS
              value: '-Xmx2048m -Duser.home=/home/developer'
            - name: MAVEN_MIRROR_URL
              value: 'http://nexus.opentlc-shared.svc:8081/repository/maven-all-public'
          memoryRequest: 512Mi
          sourceMapping: /projects
          cpuLimit: 1500m
          volumeMounts:
            - name: m2
              path: /home/developer/.m2
            - name: checode
              path: /checode
          memoryLimit: 3Gi
          image: 'quay.io/redhat-emea-ssa-team/workshop-tools:6.4'
          args:
            - sh
            - '-c'
            - '${PLUGIN_REMOTE_ENDPOINT_EXECUTABLE}'
          endpoints:
            - attributes:
                protocol: http
              exposure: public
              name: 8080-port
              protocol: http
              targetPort: 8080
            - attributes:
                protocol: http
              exposure: public
              name: 9000-port
              protocol: http
              targetPort: 9000
            - attributes:
                protocol: http
                public: 'false'
              exposure: internal
              name: 5005-port
              protocol: http
              targetPort: 5005
            - attributes:
                contributed-by: che-code.eclipse.org
                cookiesAuthEnabled: true
                discoverable: false
                type: main
                urlRewriteSupported: true
              exposure: public
              name: che-code
              path: '?tkn=eclipse-che'
              protocol: https
              secure: false
              targetPort: 3100
            - attributes:
                contributed-by: che-code.eclipse.org
                discoverable: false
                urlRewriteSupported: true
              exposure: public
              name: code-redirect-1
              protocol: http
              targetPort: 13131
            - attributes:
                contributed-by: che-code.eclipse.org
                discoverable: false
                urlRewriteSupported: true
              exposure: public
              name: code-redirect-2
              protocol: http
              targetPort: 13132
            - attributes:
                contributed-by: che-code.eclipse.org
                discoverable: false
                urlRewriteSupported: true
              exposure: public
              name: code-redirect-3
              protocol: http
              targetPort: 13133
          mountSources: true
        name: workshop-tools
      - name: m2
        volume:
          size: 1G
      - name: che-code-workspace
        plugin:
          kubernetes:
            name: che-code-workspace
            namespace: user2-devspaces
    projects:
      - attributes:
          source-origin: branch
        git:
          checkoutFrom:
            revision: '6.4'
          remotes:
            origin: >-
              https://github.com/RedHat-EMEA-SSA-Team/end-to-end-developer-workshop.git
        name: workshop

*/		