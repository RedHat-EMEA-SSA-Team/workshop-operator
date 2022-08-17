package controllers

import (
	"context"
	"errors"

	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/prometheus/common/log"
	"k8s.io/apimachinery/pkg/types"
)

// ApproveInstallPlan approves manually the install of a specific CSV
func (r *WorkshopReconciler) ApproveInstallPlan(clusterServiceVersion string, subscriptionName string, namespace string) error {

	subscription := &olmv1alpha1.Subscription{}

	if err := r.Get(context.TODO(), types.NamespacedName{Name: subscriptionName, Namespace: namespace}, subscription); err != nil {
		return err
	}

	if (clusterServiceVersion == "" && subscription.Status.InstalledCSV == "") ||
		(clusterServiceVersion != "" && (subscription.Status.InstalledCSV != clusterServiceVersion)) {
		if subscription.Status.InstallPlanRef == nil {
			return errors.New("InstallPlan Approval: Subscription is not ready yet")
		}

		installPlan := &olmv1alpha1.InstallPlan{}
		if err := r.Get(context.TODO(), types.NamespacedName{Name: subscription.Status.InstallPlanRef.Name, Namespace: namespace}, installPlan); err != nil {
			return err
		}

		if (clusterServiceVersion == "" || util.StringInSlice(clusterServiceVersion, installPlan.Spec.ClusterServiceVersionNames)) && !installPlan.Spec.Approved {
			installPlan.Spec.Approved = true
			if err := r.Update(context.TODO(), installPlan); err != nil {
				return err
			}
			log.Infof("%s InstallPlan in %s project Approved", installPlan.Name, namespace)
		}
	}
	return nil
}
