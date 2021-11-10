package controllers

import (
	"context"
	"fmt"
	"reflect"

	argocdv2 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	commonargocdoperator "github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/argocdoperator"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/kubernetes"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"
	argocdoperator "github.com/argoproj-labs/argocd-operator/api/v1alpha1"
	"github.com/prometheus/common/log"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling GitOps
func (r *WorkshopReconciler) reconcileGitOps(workshop *workshopv1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {
	enabledGitOps := workshop.Spec.Infrastructure.GitOps.Enabled

	if enabledGitOps {
		if result, err := r.addGitOps(workshop, users, appsHostnameSuffix, openshiftConsoleURL); util.IsRequeued(result, err) {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addGitOps(workshop *workshopv1.Workshop, users int,
	appsHostnameSuffix string, openshiftConsoleURL string) (reconcile.Result, error) {

	name := "openshift-gitops-operator"
	operatorNamespace := "openshift-operators"
	channel := workshop.Spec.Infrastructure.GitOps.OperatorHub.Channel
	clusterServiceVersion := workshop.Spec.Infrastructure.GitOps.OperatorHub.ClusterServiceVersion

	labels := map[string]string{
		"app.kubernetes.io/part-of": "argocd",
	}

	subscription := kubernetes.NewRedHatSubscription(workshop, r.Scheme, name, operatorNamespace,
		name, channel, clusterServiceVersion)
	if err := r.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", subscription.Name)
	}

	// Approve the installation
	if err := r.ApproveInstallPlan(clusterServiceVersion, name, operatorNamespace); err != nil {
		log.Infof("Waiting for Subscription to create InstallPlan for %s", name)
		return reconcile.Result{Requeue: true}, nil
	}

	// Wait for Operator to be running
	if !kubernetes.GetK8Client().GetDeploymentStatus("gitops-operator-controller-manager", operatorNamespace) {
		return reconcile.Result{Requeue: true}, nil
	}

	namespace := kubernetes.NewNamespace(workshop, r.Scheme, "argocd")
	if err := r.Create(context.TODO(), namespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Project", namespace.Name)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(workshop.Spec.User.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("Error when Bcrypt encrypt password for Argo CD: %v", err)
		return reconcile.Result{}, err
	}
	bcryptPassword := string(hashedPassword)

	argocdPolicy := ""
	namespaceList := ""
	secretData := map[string]string{}
	configMapData := map[string]string{}

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		userRole := fmt.Sprintf("role:%s", username)
		projectName := fmt.Sprintf("%s%d", workshop.Spec.Infrastructure.Project.StagingName, id)
		if id == 1 {
			namespaceList = projectName
		} else {
			namespaceList = fmt.Sprintf("%s,%s", namespaceList, projectName)
		}

		userPolicy := `p, ` + userRole + `, applications, *, ` + projectName + `/*, allow
p, ` + userRole + `, clusters, get, https://kubernetes.default.svc, allow
p, ` + userRole + `, projects, *,` + projectName + `, allow
p, ` + userRole + `, repositories, *, http://gitea-server.gitea.svc:3000/` + username + `/*, allow
g, ` + username + `, ` + userRole + `
`
		argocdPolicy = fmt.Sprintf("%s%s", argocdPolicy, userPolicy)

		secretData[fmt.Sprintf("accounts.%s.password", username)] = bcryptPassword

		configMapData[fmt.Sprintf("accounts.%s", username)] = "login"

		labels["app.kubernetes.io/name"] = "appproject-cr"
		appProjectCustomResource := commonargocdoperator.NewAppProjectCustomResource(workshop, r.Scheme, projectName, namespace.Name, labels, argocdPolicy)
		if err := r.Create(context.TODO(), appProjectCustomResource); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			log.Infof("Created %s Custom Resource", appProjectCustomResource.Name)
		} else if errors.IsAlreadyExists(err) {
			customResourceFound := &argocdv2.AppProject{}
			if err := r.Get(context.TODO(), types.NamespacedName{Name: appProjectCustomResource.Name, Namespace: namespace.Name}, customResourceFound); err != nil {
				return reconcile.Result{}, err
			} else if err == nil {
				if !reflect.DeepEqual(appProjectCustomResource.Spec, customResourceFound.Spec) {
					customResourceFound.Spec = appProjectCustomResource.Spec
					if err := r.Update(context.TODO(), customResourceFound); err != nil {
						return reconcile.Result{}, err
					}
					log.Infof("Updated %s Custom Resource", customResourceFound.Name)
				}
			}
		}
	}

	labels["app.kubernetes.io/name"] = "argocd-secret"
	secret := kubernetes.NewStringDataSecret(workshop, r.Scheme, "argocd-secret", namespace.Name, labels, secretData)
	if err := r.Create(context.TODO(), secret); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Secret", secret.Name)
		// } else if errors.IsAlreadyExists(err) {
		// 	secretFound := &corev1.Secret{}
		// 	if err := r.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: namespace.Name}, secretFound); err != nil {
		// 		return reconcile.Result{}, err
		// 	} else if err == nil {
		// 		if !util.IsIntersectMap(secretData, secretFound.StringData) {
		// 			secretFound.StringData = secretData
		// 			if err := r.Update(context.TODO(), secretFound); err != nil {
		// 				return reconcile.Result{}, err
		// 			}
		// 			log.Infof("Updated %s Secret", secretFound.Name)
		// 		}
		// 	}
	}

	labels["app.kubernetes.io/name"] = "argocd-cm"
	configmap := kubernetes.NewConfigMap(workshop, r.Scheme, "argocd-cm", namespace.Name, labels, configMapData)
	if err := r.Create(context.TODO(), configmap); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s ConfigMap", configmap.Name)
	} else if errors.IsAlreadyExists(err) {
		configmapFound := &corev1.ConfigMap{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: configmap.Name, Namespace: namespace.Name}, configmapFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !util.IsIntersectMap(configMapData, configmapFound.Data) {
				configmapFound.Data = configMapData
				if err := r.Update(context.TODO(), configmapFound); err != nil {
					return reconcile.Result{}, err
				}
				log.Infof("Updated %s ConfigMap", configmapFound.Name)
			}
		}
	}

	labels["app.kubernetes.io/name"] = "argocd-cr"
	argoCDCustomResource := commonargocdoperator.NewArgoCDCustomResource(workshop, r.Scheme, "argocd", namespace.Name, labels, argocdPolicy)
	if err := r.Create(context.TODO(), argoCDCustomResource); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Custom Resource", argoCDCustomResource.Name)
	} else if errors.IsAlreadyExists(err) {
		customResourceFound := &argocdoperator.ArgoCD{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: argoCDCustomResource.Name, Namespace: namespace.Name}, customResourceFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !reflect.DeepEqual(&argocdPolicy, customResourceFound.Spec.RBAC.Policy) {
				customResourceFound.Spec.RBAC.Policy = &argocdPolicy
				if err := r.Update(context.TODO(), customResourceFound); err != nil {
					return reconcile.Result{}, err
				}
				log.Infof("Updated %s Custom Resource", customResourceFound.Name)
			}
		}
	}

	// Wait for ArgoCD Dex Server to be running
	// if !kubernetes.GetK8Client().GetDeploymentStatus("argocd-dex-server", namespace.Name) {
	// 	return reconcile.Result{Requeue: true}, nil
	// }

	// Wait for ArgoCD Server to be running
	if !kubernetes.GetK8Client().GetDeploymentStatus("argocd-server", namespace.Name) {
		return reconcile.Result{Requeue: true}, nil
	}

	labels["app.kubernetes.io/name"] = "argocd-default-cluster-config"

	if result, err := r.manageArgocdDefaultClusterConfigSecret(workshop, namespace.Name, labels, namespaceList); util.IsRequeued(result, err) {
		return result, err
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) manageArgocdDefaultClusterConfigSecret(workshop *workshopv1.Workshop, namespaceName string,
	labels map[string]string, namespaceList string) (reconcile.Result, error) {

	secretName := "argocd-default-cluster-config"
	clusterConfigSecretData := map[string]string{}

	clusterConfigSecretData["config"] = "{\"tlsClientConfig\":{\"insecure\":false}}"
	clusterConfigSecretData["name"] = "in-cluster"
	clusterConfigSecretData["namespaces"] = namespaceList
	clusterConfigSecretData["server"] = "https://kubernetes.default.svc"

	clusterConfigSecret := kubernetes.NewStringDataSecret(workshop, r.Scheme, secretName, namespaceName, labels, clusterConfigSecretData)
	if err := r.Create(context.TODO(), clusterConfigSecret); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Secret", clusterConfigSecret.Name)
	} else if errors.IsAlreadyExists(err) {
		clusterConfigSecretFound := &corev1.Secret{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: clusterConfigSecret.Name, Namespace: namespaceName}, clusterConfigSecretFound); err != nil {
			return reconcile.Result{}, err
		} else if err == nil {
			if !util.IsIntersectMap(clusterConfigSecretData, clusterConfigSecretFound.StringData) {
				clusterConfigSecretFound.StringData = clusterConfigSecretData
				if err := r.Update(context.TODO(), clusterConfigSecretFound); err != nil {
					return reconcile.Result{}, err
				}
				log.Infof("Updated %s Secret", clusterConfigSecretFound.Name)
			}
		}
	}

	//Success
	return reconcile.Result{}, nil
}
