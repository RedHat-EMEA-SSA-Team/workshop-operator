package controllers

import (
	"context"
	"fmt"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/kubernetes"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/log"

	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling Project
func (r *WorkshopReconciler) reconcileProject(workshop *workshopv1.Workshop, users int) (reconcile.Result, error) {
	enabledProject := workshop.Spec.Infrastructure.Project.Enabled

	id := 1
	for {
		username := fmt.Sprintf("user%d", id)
		stagingProjectName := fmt.Sprintf("%s%d", workshop.Spec.Infrastructure.Project.StagingName, id)

		if id <= users && enabledProject {
			// Project
			if workshop.Spec.Infrastructure.Project.StagingName != "" {
				if result, err := r.addProject(workshop, stagingProjectName, username); util.IsRequeued(result, err) {
					return result, err
				}
			}

		} else {
			stagingProjectNamespace := kubernetes.NewNamespace(workshop, r.Scheme, stagingProjectName)
			stagingProjectNamespaceFound := &corev1.Namespace{}
			stagingProjectNamespaceErr := r.Get(context.TODO(), types.NamespacedName{Name: stagingProjectNamespace.Name}, stagingProjectNamespaceFound)

			if stagingProjectNamespaceErr != nil && errors.IsNotFound(stagingProjectNamespaceErr) {
				break
			}

			if !(stagingProjectNamespaceErr != nil && errors.IsNotFound(stagingProjectNamespaceErr)) {
				if result, err := r.deleteProject(stagingProjectNamespace); util.IsRequeued(result, err) {
					return result, err
				}
			}
		}

		id++
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addProject(workshop *workshopv1.Workshop, projectName string, username string) (reconcile.Result, error) {

	labels := map[string]string{
		"argocd.argoproj.io/managed-by": "argocd",
	}

	projectNamespace := kubernetes.NewNamespaceAnnotate(workshop, r.Scheme, projectName, labels, nil)
	if err := r.Create(context.TODO(), projectNamespace); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Namespace", projectNamespace.Name)
	}

	if result, err := r.manageRoles(workshop, projectNamespace.Name, username); err != nil {
		return result, err
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) deleteProject(namespaces *corev1.Namespace) (reconcile.Result, error) {

	if err := r.Delete(context.TODO(), namespaces); err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Deleted %s Namespace", namespaces.Name)
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) manageRoles(workshop *workshopv1.Workshop, projectName string, username string) (reconcile.Result, error) {

	labels := map[string]string{
		"app.kubernetes.io/part-of": "project",
	}

	users := []rbac.Subject{}
	userSubject := rbac.Subject{
		Kind: rbac.UserKind,
		Name: username,
	}

	users = append(users, userSubject)

	// User
	userRoleBinding := kubernetes.NewRoleBindingUsers(workshop, r.Scheme, username+"-project", projectName, labels,
		users, "edit", "ClusterRole")
	if err := r.Create(context.TODO(), userRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role Binding", userRoleBinding.Name)
	}

	// Default
	defaultRoleBinding := kubernetes.NewRoleBindingSA(workshop, r.Scheme, username+"-default", projectName, labels,
		"default", "view", "ClusterRole")
	if err := r.Create(context.TODO(), defaultRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role Binding", defaultRoleBinding.Name)
	}

	//Argo CD
	argocdSAs := []rbac.Subject{}
	argocdApplicationControllerSubject := rbac.Subject{
		Kind: rbac.ServiceAccountKind,
		Name: "argocd-argocd-application-controller",
		Namespace: "argocd",
	}
	argocdSAs = append(argocdSAs, argocdApplicationControllerSubject)

	argocdServerSubject := rbac.Subject{
		Kind: rbac.ServiceAccountKind,
		Name: "argocd-argocd-server",
		Namespace: "argocd",
	}
	argocdSAs = append(argocdSAs, argocdServerSubject)

	argocdManagerRole := kubernetes.NewRole(workshop, r.Scheme,
		"argocd-manager", projectName, labels, kubernetes.ArgoCDRules())
	if err := r.Create(context.TODO(), argocdManagerRole); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role in %s namespace", argocdManagerRole.Name, projectName)
	}

	argocdManagerRoleBinding := kubernetes.NewRoleBindingUsers(workshop, r.Scheme,
		"argocd-manager", projectName, labels, argocdSAs, argocdManagerRole.Name, "Role")
	if err := r.Create(context.TODO(), argocdManagerRoleBinding); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Role Binding in %s namespace", argocdManagerRoleBinding.Name, projectName)
	}
	// } else if errors.IsAlreadyExists(err) {
	// 	found := &rbac.RoleBinding{}
	// 	if err := r.Get(context.TODO(), types.NamespacedName{Name: argocdManagerRoleBinding.Name, Namespace: projectName}, found); err != nil {
	// 		return reconcile.Result{}, err
	// 	} else if err == nil {
	// 		if !reflect.DeepEqual(argocdSAs, found.Subjects) {
	// 			found.Subjects = argocdSAs
	// 			if err := r.Update(context.TODO(), found); err != nil {
	// 				return reconcile.Result{}, err
	// 			}
	// 			log.Infof("Updated %s Role Binding in %s namespace", found.Name, projectName)
	// 		}
	// 	}
	// }

	//Success
	return reconcile.Result{}, nil
}
