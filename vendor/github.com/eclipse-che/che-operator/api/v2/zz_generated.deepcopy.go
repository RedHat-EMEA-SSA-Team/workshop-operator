// +build !ignore_autogenerated

//
// Copyright (c) 2019-2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

// Code generated by controller-gen. DO NOT EDIT.

package v2

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Auth) DeepCopyInto(out *Auth) {
	*out = *in
	if in.OAuthAccessTokenInactivityTimeoutSeconds != nil {
		in, out := &in.OAuthAccessTokenInactivityTimeoutSeconds, &out.OAuthAccessTokenInactivityTimeoutSeconds
		*out = new(int32)
		**out = **in
	}
	if in.OAuthAccessTokenMaxAgeSeconds != nil {
		in, out := &in.OAuthAccessTokenMaxAgeSeconds, &out.OAuthAccessTokenMaxAgeSeconds
		*out = new(int32)
		**out = **in
	}
	in.Gateway.DeepCopyInto(&out.Gateway)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Auth.
func (in *Auth) DeepCopy() *Auth {
	if in == nil {
		return nil
	}
	out := new(Auth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BitBucketService) DeepCopyInto(out *BitBucketService) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BitBucketService.
func (in *BitBucketService) DeepCopy() *BitBucketService {
	if in == nil {
		return nil
	}
	out := new(BitBucketService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheCluster) DeepCopyInto(out *CheCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheCluster.
func (in *CheCluster) DeepCopy() *CheCluster {
	if in == nil {
		return nil
	}
	out := new(CheCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterComponents) DeepCopyInto(out *CheClusterComponents) {
	*out = *in
	out.DevWorkspace = in.DevWorkspace
	in.CheServer.DeepCopyInto(&out.CheServer)
	in.PluginRegistry.DeepCopyInto(&out.PluginRegistry)
	in.DevfileRegistry.DeepCopyInto(&out.DevfileRegistry)
	in.Database.DeepCopyInto(&out.Database)
	in.Dashboard.DeepCopyInto(&out.Dashboard)
	out.ImagePuller = in.ImagePuller
	out.Metrics = in.Metrics
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterComponents.
func (in *CheClusterComponents) DeepCopy() *CheClusterComponents {
	if in == nil {
		return nil
	}
	out := new(CheClusterComponents)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterContainerRegistry) DeepCopyInto(out *CheClusterContainerRegistry) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterContainerRegistry.
func (in *CheClusterContainerRegistry) DeepCopy() *CheClusterContainerRegistry {
	if in == nil {
		return nil
	}
	out := new(CheClusterContainerRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterDevEnvironments) DeepCopyInto(out *CheClusterDevEnvironments) {
	*out = *in
	in.Storage.DeepCopyInto(&out.Storage)
	if in.DefaultPlugins != nil {
		in, out := &in.DefaultPlugins, &out.DefaultPlugins
		*out = make([]WorkspaceDefaultPlugins, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]v1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.DefaultNamespace.DeepCopyInto(&out.DefaultNamespace)
	if in.TrustedCerts != nil {
		in, out := &in.TrustedCerts, &out.TrustedCerts
		*out = new(TrustedCerts)
		**out = **in
	}
	if in.DefaultComponents != nil {
		in, out := &in.DefaultComponents, &out.DefaultComponents
		*out = make([]v1alpha2.Component, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecondsOfInactivityBeforeIdling != nil {
		in, out := &in.SecondsOfInactivityBeforeIdling, &out.SecondsOfInactivityBeforeIdling
		*out = new(int32)
		**out = **in
	}
	if in.SecondsOfRunBeforeIdling != nil {
		in, out := &in.SecondsOfRunBeforeIdling, &out.SecondsOfRunBeforeIdling
		*out = new(int32)
		**out = **in
	}
	if in.DisableContainerBuildCapabilities != nil {
		in, out := &in.DisableContainerBuildCapabilities, &out.DisableContainerBuildCapabilities
		*out = new(bool)
		**out = **in
	}
	if in.ContainerBuildConfiguration != nil {
		in, out := &in.ContainerBuildConfiguration, &out.ContainerBuildConfiguration
		*out = new(ContainerBuildConfiguration)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterDevEnvironments.
func (in *CheClusterDevEnvironments) DeepCopy() *CheClusterDevEnvironments {
	if in == nil {
		return nil
	}
	out := new(CheClusterDevEnvironments)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterGitServices) DeepCopyInto(out *CheClusterGitServices) {
	*out = *in
	if in.GitHub != nil {
		in, out := &in.GitHub, &out.GitHub
		*out = make([]GitHubService, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.GitLab != nil {
		in, out := &in.GitLab, &out.GitLab
		*out = make([]GitLabService, len(*in))
		copy(*out, *in)
	}
	if in.BitBucket != nil {
		in, out := &in.BitBucket, &out.BitBucket
		*out = make([]BitBucketService, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterGitServices.
func (in *CheClusterGitServices) DeepCopy() *CheClusterGitServices {
	if in == nil {
		return nil
	}
	out := new(CheClusterGitServices)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterList) DeepCopyInto(out *CheClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CheCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterList.
func (in *CheClusterList) DeepCopy() *CheClusterList {
	if in == nil {
		return nil
	}
	out := new(CheClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CheClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpec) DeepCopyInto(out *CheClusterSpec) {
	*out = *in
	in.DevEnvironments.DeepCopyInto(&out.DevEnvironments)
	in.Components.DeepCopyInto(&out.Components)
	in.GitServices.DeepCopyInto(&out.GitServices)
	in.Networking.DeepCopyInto(&out.Networking)
	out.ContainerRegistry = in.ContainerRegistry
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpec.
func (in *CheClusterSpec) DeepCopy() *CheClusterSpec {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterSpecNetworking) DeepCopyInto(out *CheClusterSpecNetworking) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Auth.DeepCopyInto(&out.Auth)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterSpecNetworking.
func (in *CheClusterSpecNetworking) DeepCopy() *CheClusterSpecNetworking {
	if in == nil {
		return nil
	}
	out := new(CheClusterSpecNetworking)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheClusterStatus) DeepCopyInto(out *CheClusterStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheClusterStatus.
func (in *CheClusterStatus) DeepCopy() *CheClusterStatus {
	if in == nil {
		return nil
	}
	out := new(CheClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CheServer) DeepCopyInto(out *CheServer) {
	*out = *in
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(Deployment)
		(*in).DeepCopyInto(*out)
	}
	if in.Debug != nil {
		in, out := &in.Debug, &out.Debug
		*out = new(bool)
		**out = **in
	}
	if in.ClusterRoles != nil {
		in, out := &in.ClusterRoles, &out.ClusterRoles
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Proxy != nil {
		in, out := &in.Proxy, &out.Proxy
		*out = new(Proxy)
		(*in).DeepCopyInto(*out)
	}
	if in.ExtraProperties != nil {
		in, out := &in.ExtraProperties, &out.ExtraProperties
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CheServer.
func (in *CheServer) DeepCopy() *CheServer {
	if in == nil {
		return nil
	}
	out := new(CheServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Container) DeepCopyInto(out *Container) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Container.
func (in *Container) DeepCopy() *Container {
	if in == nil {
		return nil
	}
	out := new(Container)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerBuildConfiguration) DeepCopyInto(out *ContainerBuildConfiguration) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerBuildConfiguration.
func (in *ContainerBuildConfiguration) DeepCopy() *ContainerBuildConfiguration {
	if in == nil {
		return nil
	}
	out := new(ContainerBuildConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Dashboard) DeepCopyInto(out *Dashboard) {
	*out = *in
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(Deployment)
		(*in).DeepCopyInto(*out)
	}
	if in.HeaderMessage != nil {
		in, out := &in.HeaderMessage, &out.HeaderMessage
		*out = new(DashboardHeaderMessage)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Dashboard.
func (in *Dashboard) DeepCopy() *Dashboard {
	if in == nil {
		return nil
	}
	out := new(Dashboard)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DashboardHeaderMessage) DeepCopyInto(out *DashboardHeaderMessage) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DashboardHeaderMessage.
func (in *DashboardHeaderMessage) DeepCopy() *DashboardHeaderMessage {
	if in == nil {
		return nil
	}
	out := new(DashboardHeaderMessage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Database) DeepCopyInto(out *Database) {
	*out = *in
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(Deployment)
		(*in).DeepCopyInto(*out)
	}
	if in.Pvc != nil {
		in, out := &in.Pvc, &out.Pvc
		*out = new(PVC)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Database.
func (in *Database) DeepCopy() *Database {
	if in == nil {
		return nil
	}
	out := new(Database)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DefaultNamespace) DeepCopyInto(out *DefaultNamespace) {
	*out = *in
	if in.AutoProvision != nil {
		in, out := &in.AutoProvision, &out.AutoProvision
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DefaultNamespace.
func (in *DefaultNamespace) DeepCopy() *DefaultNamespace {
	if in == nil {
		return nil
	}
	out := new(DefaultNamespace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Deployment) DeepCopyInto(out *Deployment) {
	*out = *in
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecurityContext != nil {
		in, out := &in.SecurityContext, &out.SecurityContext
		*out = new(PodSecurityContext)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Deployment.
func (in *Deployment) DeepCopy() *Deployment {
	if in == nil {
		return nil
	}
	out := new(Deployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevWorkspace) DeepCopyInto(out *DevWorkspace) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevWorkspace.
func (in *DevWorkspace) DeepCopy() *DevWorkspace {
	if in == nil {
		return nil
	}
	out := new(DevWorkspace)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevfileRegistry) DeepCopyInto(out *DevfileRegistry) {
	*out = *in
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(Deployment)
		(*in).DeepCopyInto(*out)
	}
	if in.ExternalDevfileRegistries != nil {
		in, out := &in.ExternalDevfileRegistries, &out.ExternalDevfileRegistries
		*out = make([]ExternalDevfileRegistry, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevfileRegistry.
func (in *DevfileRegistry) DeepCopy() *DevfileRegistry {
	if in == nil {
		return nil
	}
	out := new(DevfileRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalDevfileRegistry) DeepCopyInto(out *ExternalDevfileRegistry) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalDevfileRegistry.
func (in *ExternalDevfileRegistry) DeepCopy() *ExternalDevfileRegistry {
	if in == nil {
		return nil
	}
	out := new(ExternalDevfileRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalPluginRegistry) DeepCopyInto(out *ExternalPluginRegistry) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalPluginRegistry.
func (in *ExternalPluginRegistry) DeepCopy() *ExternalPluginRegistry {
	if in == nil {
		return nil
	}
	out := new(ExternalPluginRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Gateway) DeepCopyInto(out *Gateway) {
	*out = *in
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(Deployment)
		(*in).DeepCopyInto(*out)
	}
	if in.ConfigLabels != nil {
		in, out := &in.ConfigLabels, &out.ConfigLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Gateway.
func (in *Gateway) DeepCopy() *Gateway {
	if in == nil {
		return nil
	}
	out := new(Gateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitHubService) DeepCopyInto(out *GitHubService) {
	*out = *in
	if in.DisableSubdomainIsolation != nil {
		in, out := &in.DisableSubdomainIsolation, &out.DisableSubdomainIsolation
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitHubService.
func (in *GitHubService) DeepCopy() *GitHubService {
	if in == nil {
		return nil
	}
	out := new(GitHubService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitLabService) DeepCopyInto(out *GitLabService) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitLabService.
func (in *GitLabService) DeepCopy() *GitLabService {
	if in == nil {
		return nil
	}
	out := new(GitLabService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImagePuller) DeepCopyInto(out *ImagePuller) {
	*out = *in
	out.Spec = in.Spec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImagePuller.
func (in *ImagePuller) DeepCopy() *ImagePuller {
	if in == nil {
		return nil
	}
	out := new(ImagePuller)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PVC) DeepCopyInto(out *PVC) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PVC.
func (in *PVC) DeepCopy() *PVC {
	if in == nil {
		return nil
	}
	out := new(PVC)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PluginRegistry) DeepCopyInto(out *PluginRegistry) {
	*out = *in
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(Deployment)
		(*in).DeepCopyInto(*out)
	}
	if in.ExternalPluginRegistries != nil {
		in, out := &in.ExternalPluginRegistries, &out.ExternalPluginRegistries
		*out = make([]ExternalPluginRegistry, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PluginRegistry.
func (in *PluginRegistry) DeepCopy() *PluginRegistry {
	if in == nil {
		return nil
	}
	out := new(PluginRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodSecurityContext) DeepCopyInto(out *PodSecurityContext) {
	*out = *in
	if in.RunAsUser != nil {
		in, out := &in.RunAsUser, &out.RunAsUser
		*out = new(int64)
		**out = **in
	}
	if in.FsGroup != nil {
		in, out := &in.FsGroup, &out.FsGroup
		*out = new(int64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodSecurityContext.
func (in *PodSecurityContext) DeepCopy() *PodSecurityContext {
	if in == nil {
		return nil
	}
	out := new(PodSecurityContext)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Proxy) DeepCopyInto(out *Proxy) {
	*out = *in
	if in.NonProxyHosts != nil {
		in, out := &in.NonProxyHosts, &out.NonProxyHosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Proxy.
func (in *Proxy) DeepCopy() *Proxy {
	if in == nil {
		return nil
	}
	out := new(Proxy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceList) DeepCopyInto(out *ResourceList) {
	*out = *in
	out.Memory = in.Memory.DeepCopy()
	out.Cpu = in.Cpu.DeepCopy()
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceList.
func (in *ResourceList) DeepCopy() *ResourceList {
	if in == nil {
		return nil
	}
	out := new(ResourceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceRequirements) DeepCopyInto(out *ResourceRequirements) {
	*out = *in
	if in.Requests != nil {
		in, out := &in.Requests, &out.Requests
		*out = new(ResourceList)
		(*in).DeepCopyInto(*out)
	}
	if in.Limits != nil {
		in, out := &in.Limits, &out.Limits
		*out = new(ResourceList)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceRequirements.
func (in *ResourceRequirements) DeepCopy() *ResourceRequirements {
	if in == nil {
		return nil
	}
	out := new(ResourceRequirements)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServerMetrics) DeepCopyInto(out *ServerMetrics) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServerMetrics.
func (in *ServerMetrics) DeepCopy() *ServerMetrics {
	if in == nil {
		return nil
	}
	out := new(ServerMetrics)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrustedCerts) DeepCopyInto(out *TrustedCerts) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrustedCerts.
func (in *TrustedCerts) DeepCopy() *TrustedCerts {
	if in == nil {
		return nil
	}
	out := new(TrustedCerts)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkspaceDefaultPlugins) DeepCopyInto(out *WorkspaceDefaultPlugins) {
	*out = *in
	if in.Plugins != nil {
		in, out := &in.Plugins, &out.Plugins
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkspaceDefaultPlugins.
func (in *WorkspaceDefaultPlugins) DeepCopy() *WorkspaceDefaultPlugins {
	if in == nil {
		return nil
	}
	out := new(WorkspaceDefaultPlugins)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkspaceStorage) DeepCopyInto(out *WorkspaceStorage) {
	*out = *in
	if in.PerUserStrategyPvcConfig != nil {
		in, out := &in.PerUserStrategyPvcConfig, &out.PerUserStrategyPvcConfig
		*out = new(PVC)
		**out = **in
	}
	if in.PerWorkspaceStrategyPvcConfig != nil {
		in, out := &in.PerWorkspaceStrategyPvcConfig, &out.PerWorkspaceStrategyPvcConfig
		*out = new(PVC)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkspaceStorage.
func (in *WorkspaceStorage) DeepCopy() *WorkspaceStorage {
	if in == nil {
		return nil
	}
	out := new(WorkspaceStorage)
	in.DeepCopyInto(out)
	return out
}
