package controllers

import (
	"context"
	"fmt"

//	aov1 "github.com/openshift/api/operator/v1"

	workshopv1 "github.com/RedHat-EMEA-SSA-Team/workshop-operator/api/v1"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/kubernetes"
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    
	"github.com/RedHat-EMEA-SSA-Team/workshop-operator/common/util"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciling Pipeline
func (r *WorkshopReconciler) reconcilePipelines(ctx context.Context, workshop *workshopv1.Workshop) (reconcile.Result, error) {
	enabledPipeline := workshop.Spec.Infrastructure.Pipeline.Enabled

	if enabledPipeline {
		if result, err := r.addPipelines(ctx, workshop); util.IsRequeued(result, err) {
			return result, err
		}
	}

	//Success
	return reconcile.Result{}, nil
}

func (r *WorkshopReconciler) addPipelines(ctx context.Context, workshop *workshopv1.Workshop) (reconcile.Result, error) {

	name := "openshift-pipelines-operator-rh"
	channel := workshop.Spec.Infrastructure.Pipeline.OperatorHub.Channel
	clusterServiceVersion := workshop.Spec.Infrastructure.Pipeline.OperatorHub.ClusterServiceVersion

	pipelineSubscription := kubernetes.NewRedHatSubscription(workshop, r.Scheme, name, "openshift-operators",
		name, channel, clusterServiceVersion)
	if err := r.Create(context.TODO(), pipelineSubscription); err != nil && !errors.IsAlreadyExists(err) {
		return reconcile.Result{}, err
	} else if err == nil {
		log.Infof("Created %s Subscription", pipelineSubscription.Name)
	}

	// Approve the installation
	if err := r.ApproveInstallPlan(clusterServiceVersion, name, "openshift-operators"); err != nil {
		log.Infof("Waiting for Subscription to create InstallPlan for %s", name)
		return reconcile.Result{Requeue: true}, nil
	}


	// 1. Define the GVK for the Console Operator resource
    consoleGVK := schema.GroupVersionKind{
        Group:   "operator.openshift.io",
        Version: "v1",
        Kind:    "Console",
    }

    // 2. Initialize an Unstructured object
    console := &unstructured.Unstructured{}
    console.SetGroupVersionKind(consoleGVK)

    // 3. Fetch the "cluster" instance
    err := r.Get(ctx, client.ObjectKey{Name: "cluster"}, console)
    if err != nil {
        // If the resource doesn't exist (not an OpenShift cluster?), skip
        return reconcile.Result{}, client.IgnoreNotFound(err)
    }

    // 4. Extract the existing plugins list safely
    plugins, found, err := unstructured.NestedStringSlice(console.Object, "spec", "plugins")
    if err != nil {
        return reconcile.Result{}, fmt.Errorf("failed to parse console plugins: %w", err)
    }
    if !found {
        plugins = []string{}
		log.Infof("No console plugins found")
    }

    // 5. Check if our plugin needs to be added
    targetPlugin := "pipelines-console-plugin"
    needsUpdate := true
    for _, p := range plugins {
        if p == targetPlugin {
            needsUpdate = false
            break
        }
    }

    if needsUpdate {
        // Create a patch object from the current state
        patchBase := client.MergeFrom(console.DeepCopy())
        
        // Update the slice and put it back into the unstructured object
        plugins = append(plugins, targetPlugin)
        if err := unstructured.SetNestedStringSlice(console.Object, plugins, "spec", "plugins"); err != nil {
            return reconcile.Result{}, fmt.Errorf("failed to set console plugins: %w", err)
        }

        // 6. Perform the Patch
        if err := r.Patch(ctx, console, patchBase); err != nil {
            return reconcile.Result{}, fmt.Errorf("failed to patch console plugins: %w", err)
        }

		log.Infof("Added console plugin: %s", targetPlugin)

    }


	//Success
	return reconcile.Result{}, nil
}
