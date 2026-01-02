package controllers

import (
	"context"
	"fmt"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/kubernetes"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/log"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/istio"
//	maistrav1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/maistra/v1"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"

	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	//	v1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/maistra/v1"
	//	kialiConfig "github.com/kiali/kiali/config"
	//	ctl "github.com/argoproj/gitops-engine/pkg/utils/kube/ctl"
)

// Reconciling ServiceMesh
func (r *WorkshopReconciler) reconcileServiceMesh(workshop *workshopv1.Workshop, users int) (reconcile.Result, error) {
	enabledServiceMesh := workshop.Spec.Infrastructure.ServiceMesh.Enabled
	enabledServerless := workshop.Spec.Infrastructure.Serverless.Enabled

	if enabledServiceMesh || enabledServerless {

//		if result, err := r.addElasticSearchOperator(workshop); util.IsRequeued(result, err) {
//			return result, err
//		}

//		if result, err := r.addJaegerOperator(workshop); util.IsRequeued(result, err) {
//			return result, err
//		}

		if result, err := r.addKialiOperator(workshop); util.IsRequeued(result, err) {
			return result, err
		}

		if result, err := r.addServiceMesh(workshop, users); util.IsRequeued(result, err) {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addServiceMesh(workshop *workshopv1.Workshop, users int) (reconcile.Result, error) {

	const operatorNamespace = "openshift-operators"
	const kialiName = "kiali-user-workload-monitoring"
	const istioName = "default"
	const istioNamespace = "istio-system"
	const cniNamespace = "istio-cni"

	// Service Mesh Operator
	channel := workshop.Spec.Infrastructure.ServiceMesh.ServiceMeshOperatorHub.Channel
	clusterserviceversion := workshop.Spec.Infrastructure.ServiceMesh.ServiceMeshOperatorHub.ClusterServiceVersion

	subscription := kubernetes.NewRedHatSubscription(workshop, r.Scheme, "servicemeshoperator3", operatorNamespace,
		"servicemeshoperator3", channel, clusterserviceversion)
	if err := r.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "servicemeshoperator3", operatorNamespace); err != nil {
		log.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	// Wait for Operator to be running
	if !kubernetes.GetK8Client().GetDeploymentStatus("servicemesh-operator3", operatorNamespace) {
		return reconcile.Result{Requeue: true}, nil
	}

	// Create namespace with labels
	// oc label namespace istio-system istio-discovery=enabled
	// oc label namespace istio-cni istio-discovery=enabled
	discoveryLabels := map[string]string{
		"istio-discovery":   "enabled",
	}

	annotations := map[string]string{
	}


	// Deploy Service Mesh Projects
	istioSystemNamespace := kubernetes.NewNamespaceAnnotate(workshop, r.Scheme, istioNamespace, discoveryLabels, annotations)
	if err := r.Create(context.TODO(), istioSystemNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Namespace", istioSystemNamespace.Name)
	}

	cniSystemNamespace := kubernetes.NewNamespaceAnnotate(workshop, r.Scheme, cniNamespace, discoveryLabels, annotations)
	if err := r.Create(context.TODO(), cniSystemNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Namespace", cniSystemNamespace.Name)
	}

	zNamespace := kubernetes.NewNamespaceAnnotate(workshop, r.Scheme, "ztunnel", discoveryLabels, annotations)
	if err := r.Create(context.TODO(), zNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Namespace", zNamespace.Name)
	}


/*
	istioMembers := []string{}
	istioUsers := []rbac.Subject{}

	if workshop.Spec.Infrastructure.GitOps.Enabled {
		argocdSubject := rbac.Subject{
			Kind:     rbac.UserKind,
			Name:     "system:serviceaccount:argocd:argocd-argocd-application-controller",
			APIGroup: "rbac.authorization.k8s.io",
		}
		istioUsers = append(istioUsers, argocdSubject)
	}

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		stagingProjectName := fmt.Sprintf("%s%d", workshop.Spec.Infrastructure.Project.StagingName, id)
		userSubject := rbac.Subject{
			Kind:     rbac.UserKind,
			Name:     username,
			APIGroup: "rbac.authorization.k8s.io", 
		}

		istioMembers = append(istioMembers, stagingProjectName)
		istioUsers = append(istioUsers, userSubject)
	}


	jaegerRole := kubernetes.NewRole(workshop, r.Scheme,
		"jaeger-user", "istio-system", labels, kubernetes.JaegerUserRules())
	if err := r.Create(context.TODO(), jaegerRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role", jaegerRole.Name)
	}

	jaegerRoleBinding := kubernetes.NewRoleBindingUsers(workshop, r.Scheme,
		"jaeger-users", "istio-system", labels, istioUsers, jaegerRole.Name, "Role")
	if err := r.Create(context.TODO(), jaegerRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role Binding", jaegerRoleBinding.Name)
	} else if errors.IsAlreadyExists(err) {
		found := &rbac.RoleBinding{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: jaegerRoleBinding.Name, Namespace: istioSystemNamespace.Name}, found); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(istioUsers, found.Subjects) {
				found.Subjects = istioUsers
				if err := r.Update(context.TODO(), found); err != nil {
					return reconcile.Result{}, err
				}
				log.Infof("Updated %s Role Binding", found.Name)
			}
		}
	}
*/
	/*
	labels := map[string]string{
		"app.kubernetes.io/part-of": "istio",
	}

	meshUserRoleBinding := kubernetes.NewRoleBindingUsers(workshop, r.Scheme,
		"mesh-users", "istio-system", labels, istioUsers, "mesh-user", "Role")	

	if err := r.Create(context.TODO(), meshUserRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role Binding", meshUserRoleBinding.Name)
	}

	// To avoid UI errors in Kiali fetching the deployment status of the ingressgateway we need this extra Role 
	// and Binding
	istioSysDeployRole := kubernetes.NewRole(workshop, r.Scheme,
		"istio-system-deploy-status", "istio-system", labels, kubernetes.KialiUserRules())
	if err := r.Create(context.TODO(), istioSysDeployRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role", istioSysDeployRole.Name)
	}

	meshUserViewRoleBinding := kubernetes.NewRoleBindingUsers(workshop, r.Scheme,
		"mesh-users-view", "istio-system", labels, istioUsers, "istio-system-deploy-status", "Role")
	if err := r.Create(context.TODO(), meshUserViewRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role Binding", meshUserViewRoleBinding.Name)
	}
	*/

	istioSailCR := istio.NewSailCR(workshop, r.Scheme, istioName, istioSystemNamespace.Name, discoveryLabels)
	if err := r.Create(context.TODO(), istioSailCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Istio SailOperator Custom Resource", istioSailCR.Name)
	}

	istioCniCR := istio.NewCNICR(workshop, r.Scheme, istioName, cniSystemNamespace.Name, discoveryLabels)
	if err := r.Create(context.TODO(), istioCniCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Istio CNI Custom Resource", istioCniCR.Name)
	}

	kialiCR := istio.NewKialiCR(workshop, r.Scheme, kialiName, istioSystemNamespace.Name)
	if err := r.Create(context.TODO(), kialiCR); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Kiali Custom Resource", kialiName)
	}

	// create cluster role binding for kiali to monitor
	kialiMonitorRoleBinding := kubernetes.NewClusterRoleBindingSA(workshop, r.Scheme,
		"kiali-monitoring-rbac", istioNamespace, nil, "kiali-service-account", "cluster-monitoring-view", "ClusterRole")
	if err := r.Create(context.TODO(), kialiMonitorRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Cluster Role Binding", kialiMonitorRoleBinding.Name)
	}

	// enable monitoring CM for user workload
	data := map[string]string{
		"config.yaml": `enableUserWorkload: true`,
	}

	monitorCM := kubernetes.NewConfigMapAnnotate(workshop, r.Scheme, "cluster-monitoring-config", "openshift-monitoring", nil, data, nil)
	if err := r.Create(context.TODO(), monitorCM); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created config map for cluster monitoring %s", monitorCM.Name)
	}

	// enable telemetry
	telemetry := istio.NewTelemetryCR(workshop, r.Scheme, "enable-prometheus-metrics", istioNamespace)
	if err := r.Create(context.TODO(), telemetry); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created Telemetry cluster monitoring")
	}

	// enable Service monitor
	serviceMonitor := istio.NewServiceMonitorCR(workshop, r.Scheme, "istiod-monitor", istioNamespace)
	if err := r.Create(context.TODO(), serviceMonitor); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created ServiceMonitor in %s", istioNamespace)
	}

	// add PodMonitoring to istio-system 
	podMonitor := istio.NewPodMonitorCR(workshop, r.Scheme, "proxies-monitor", istioNamespace)
	if err := r.Create(context.TODO(), podMonitor); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created PodMonitor for %s", istioNamespace)
	}


	// loop through all the user projects we want to make discoverable by labelling the namespace
	// and a podmonitor to every user project so Kiali can see them
	// aslo add rbac so kiali can edit istio resources in the project
	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		stagingProjectName := fmt.Sprintf("%s%d", workshop.Spec.Infrastructure.Project.StagingName, id)
		istioUsers := []rbac.Subject{}

		var podMonitor = istio.NewPodMonitorCR(workshop, r.Scheme, "proxies-monitor", stagingProjectName)
		if err := r.Create(context.TODO(), podMonitor); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			log.Infof("Created PodMonitor for %s", stagingProjectName)
		}


		userSubject := rbac.Subject{
			Kind:     rbac.UserKind,
			Name:     username,
			APIGroup: "rbac.authorization.k8s.io", 
		}

		istioUsers = append(istioUsers, userSubject)

		meshUserRoleBinding := kubernetes.NewRoleBindingUsers(workshop, r.Scheme,
		"kiali-write", stagingProjectName, nil, istioUsers, "kiali-write-privileges", "ClusterRole")	

		if err := r.Create(context.TODO(), meshUserRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			log.Infof("Created %s Role Binding", meshUserRoleBinding.Name)
		}

	}


	
	// Now patch the Kiali CR (if ready) to disable some of the warning features
	// we use unstructured patch here because we have no Go struct definition for this object
	patchBytes := []byte(`{ "spec":{"kiali_feature_flags":{"validations":{"ignore":["KIA0302","KIA1301"]}}}}`)

	u := &unstructured.Unstructured{}
	u.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      kialiName,
			"namespace": istioNamespace,
		},
	}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kiali.io",
		Version: "v1alpha1",
		Kind:    "Kiali",
	})

	if err := r.Client.Patch(context.TODO(), u, client.RawPatch(types.MergePatchType, patchBytes)); err != nil {
		log.Infof("Kiali Custom Resource not ready")
		return reconcile.Result{}, err
	} 
    
	// embed kiali in the web console 
	embedKiali := istio.NewOSSMConsoleCR(workshop, r.Scheme)
	if err := r.Create(context.TODO(), embedKiali); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created embedded Kiali console")
	}

	//Success
	return reconcile.Result{}, nil
}
 
func (r *WorkshopReconciler) addElasticSearchOperator(workshop *workshopv1.Workshop) (reconcile.Result, error) {

	channel := workshop.Spec.Infrastructure.ServiceMesh.ElasticSearchOperatorHub.Channel
	clusterserviceversion := workshop.Spec.Infrastructure.ServiceMesh.ElasticSearchOperatorHub.ClusterServiceVersion
	subcriptionName := fmt.Sprintf("elasticsearch-operator-%s", channel)

	redhatOperatorsNamespace := kubernetes.NewNamespace(workshop, r.Scheme, "openshift-operators-redhat")
	if err := r.Create(context.TODO(), redhatOperatorsNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Namespace", redhatOperatorsNamespace.Name)

		redhatOperatorGroup := kubernetes.NewOperatorGroup(workshop, r.Scheme, "operators-redhat-group", redhatOperatorsNamespace.Name, "")
		if err := r.Create(context.TODO(), redhatOperatorGroup); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			log.Infof("Created %s OperatorGroup", redhatOperatorGroup.Name)
		}
	}

	subscription := kubernetes.NewRedHatSubscription(workshop, r.Scheme, subcriptionName, redhatOperatorsNamespace.Name,
		"elasticsearch-operator", channel, clusterserviceversion)
	if err := r.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, subcriptionName, redhatOperatorsNamespace.Name); err != nil {
		log.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addJaegerOperator(workshop *workshopv1.Workshop) (reconcile.Result, error) {

	channel := workshop.Spec.Infrastructure.ServiceMesh.JaegerOperatorHub.Channel
	clusterserviceversion := workshop.Spec.Infrastructure.ServiceMesh.JaegerOperatorHub.ClusterServiceVersion

	redhatOperatorsNamespace := kubernetes.NewNamespace(workshop, r.Scheme, "openshift-distibuted-tracing")
	if err := r.Create(context.TODO(), redhatOperatorsNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Namespace", redhatOperatorsNamespace.Name)

		redhatOperatorGroup := kubernetes.NewOperatorGroup(workshop, r.Scheme, "distributed-tracing-group", redhatOperatorsNamespace.Name, "")
		if err := r.Create(context.TODO(), redhatOperatorGroup); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			log.Infof("Created %s OperatorGroup", redhatOperatorGroup.Name)
		}
	}

	subscription := kubernetes.NewRedHatSubscription(workshop, r.Scheme, "jaeger-product", redhatOperatorsNamespace.Name,
		"jaeger-product", channel, clusterserviceversion)
	if err := r.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "jaeger-product", redhatOperatorsNamespace.Name); err != nil {
		log.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addKialiOperator(workshop *workshopv1.Workshop) (reconcile.Result, error) {

	channel := workshop.Spec.Infrastructure.ServiceMesh.KialiOperatorHub.Channel
	clusterserviceversion := workshop.Spec.Infrastructure.ServiceMesh.KialiOperatorHub.ClusterServiceVersion

	subscription := kubernetes.NewRedHatSubscription(workshop, r.Scheme, "kiali-ossm", "openshift-operators",
		"kiali-ossm", channel, clusterserviceversion)
	if err := r.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "kiali-ossm", "openshift-operators"); err != nil {
		log.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	//Success
	return reconcile.Result{}, nil
}
