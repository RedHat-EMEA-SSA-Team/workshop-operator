package controllers

import (
	"context"
	"fmt"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/kubernetes"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/prometheus/common/log"

	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciling IstioWorkspace
func (r *WorkshopReconciler) reconcileIstioWorkspace(workshop *workshopv1.Workshop, users int) (reconcile.Result, error) {
	enabled := workshop.Spec.Infrastructure.IstioWorkspace.Enabled

	if enabled {

		if result, err := r.addIstioWorkspace(workshop, users); util.IsRequeued(result, err) {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addIstioWorkspace(workshop *workshopv1.Workshop, users int) (reconcile.Result, error) {

	channel := workshop.Spec.Infrastructure.IstioWorkspace.OperatorHub.Channel
	clusterserviceversion := workshop.Spec.Infrastructure.IstioWorkspace.OperatorHub.ClusterServiceVersion

	labels := map[string]string{
		"app.kubernetes.io/part-of": "istio-workspace",
	}

	for id := 1; id <= users; id++ {
		username := fmt.Sprintf("user%d", id)
		stagingProjectName := fmt.Sprintf("%s%d", workshop.Spec.Infrastructure.Project.StagingName, id)

		role := kubernetes.NewRole(workshop, r.Scheme,
			username+"-istio-workspace", stagingProjectName, labels, kubernetes.IstioWorkspaceUserRules())
		if err := r.Create(context.TODO(), role); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			log.Infof("Created %s Role", role.Name)
		}

		users := []rbac.Subject{
			{
				Kind: rbac.UserKind,
				Name: username,
			},
		}

		roleBinding := kubernetes.NewRoleBindingUsers(workshop, r.Scheme,
			username+"-istio-workspace", stagingProjectName, labels, users, username+"-istio-workspace", "Role")
		if err := r.Create(context.TODO(), roleBinding); err != nil && !errors.IsAlreadyExists(err) {
			return reconcile.Result{}, err
		} else if err == nil {
			log.Infof("Created %s Role Binding", roleBinding.Name)
		}

		// Create SCC
		serviceAccountUser := "system:serviceaccount:" + stagingProjectName + ":default"

		privilegedSCCFound := &securityv1.SecurityContextConstraints{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: "privileged"}, privilegedSCCFound); err != nil {
			return reconcile.Result{}, err
		}

		if !util.StringInSlice(serviceAccountUser, privilegedSCCFound.Users) {
			privilegedSCCFound.Users = append(privilegedSCCFound.Users, serviceAccountUser)
			if err := r.Update(context.TODO(), privilegedSCCFound); err != nil {
				return reconcile.Result{}, err
			} else if err == nil {
				log.Infof("Updated %s SCC", privilegedSCCFound.Name)
			}
		}

		anyuidSCCFound := &securityv1.SecurityContextConstraints{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: "anyuid"}, anyuidSCCFound); err != nil {
			return reconcile.Result{}, err
		}

		if !util.StringInSlice(serviceAccountUser, anyuidSCCFound.Users) {
			anyuidSCCFound.Users = append(anyuidSCCFound.Users, serviceAccountUser)
			if err := r.Update(context.TODO(), anyuidSCCFound); err != nil {
				return reconcile.Result{}, err
			} else if err == nil {
				log.Infof("Updated %s SCC", anyuidSCCFound.Name)
			}
		}
	}

	subscription := kubernetes.NewCommunitySubscription(workshop, r.Scheme, "istio-workspace-operator", "openshift-operators",
		"istio-workspace-operator", channel, clusterserviceversion)
	if err := r.Create(context.TODO(), subscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", subscription.Name)
	}

	if err := r.ApproveInstallPlan(clusterserviceversion, "istio-workspace-operator", "openshift-operators"); err != nil {
		log.Infof("Waiting for Subscription to create InstallPlan for %s", subscription.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	//Success
	return reconcile.Result{}, nil
}
