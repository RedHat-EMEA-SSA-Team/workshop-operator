package controllers

import (
	"context"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *WorkshopReconciler) finalizeWorkshop(reqLogger logr.Logger, workshop *workshopv1.Workshop) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	reqLogger.Info("Successfully finalized workshop")
	return nil
}

func (r *WorkshopReconciler) addFinalizer(reqLogger logr.Logger, workshop *workshopv1.Workshop) error {
	reqLogger.Info("Adding Finalizer for the Workshop")
	controllerutil.AddFinalizer(workshop, workshopFinalizer)

	// Update CR
	err := r.Update(context.TODO(), workshop)
	if err != nil {
		reqLogger.Error(err, "Failed to update Workshop with finalizer")
		return err
	}
	return nil
}
