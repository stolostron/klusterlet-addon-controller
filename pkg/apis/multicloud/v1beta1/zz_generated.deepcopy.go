// +build !ignore_autogenerated

// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Code generated by operator-sdk. DO NOT EDIT.

package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationManager) DeepCopyInto(out *ApplicationManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationManager.
func (in *ApplicationManager) DeepCopy() *ApplicationManager {
	if in == nil {
		return nil
	}
	out := new(ApplicationManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApplicationManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationManagerList) DeepCopyInto(out *ApplicationManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ApplicationManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationManagerList.
func (in *ApplicationManagerList) DeepCopy() *ApplicationManagerList {
	if in == nil {
		return nil
	}
	out := new(ApplicationManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApplicationManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationManagerSpec) DeepCopyInto(out *ApplicationManagerSpec) {
	*out = *in
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationManagerSpec.
func (in *ApplicationManagerSpec) DeepCopy() *ApplicationManagerSpec {
	if in == nil {
		return nil
	}
	out := new(ApplicationManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationManagerStatus) DeepCopyInto(out *ApplicationManagerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationManagerStatus.
func (in *ApplicationManagerStatus) DeepCopy() *ApplicationManagerStatus {
	if in == nil {
		return nil
	}
	out := new(ApplicationManagerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CISController) DeepCopyInto(out *CISController) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CISController.
func (in *CISController) DeepCopy() *CISController {
	if in == nil {
		return nil
	}
	out := new(CISController)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CISController) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CISControllerList) DeepCopyInto(out *CISControllerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CISController, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CISControllerList.
func (in *CISControllerList) DeepCopy() *CISControllerList {
	if in == nil {
		return nil
	}
	out := new(CISControllerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CISControllerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CISControllerSpec) DeepCopyInto(out *CISControllerSpec) {
	*out = *in
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CISControllerSpec.
func (in *CISControllerSpec) DeepCopy() *CISControllerSpec {
	if in == nil {
		return nil
	}
	out := new(CISControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CISControllerStatus) DeepCopyInto(out *CISControllerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CISControllerStatus.
func (in *CISControllerStatus) DeepCopy() *CISControllerStatus {
	if in == nil {
		return nil
	}
	out := new(CISControllerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertPolicyController) DeepCopyInto(out *CertPolicyController) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertPolicyController.
func (in *CertPolicyController) DeepCopy() *CertPolicyController {
	if in == nil {
		return nil
	}
	out := new(CertPolicyController)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CertPolicyController) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertPolicyControllerList) DeepCopyInto(out *CertPolicyControllerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CertPolicyController, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertPolicyControllerList.
func (in *CertPolicyControllerList) DeepCopy() *CertPolicyControllerList {
	if in == nil {
		return nil
	}
	out := new(CertPolicyControllerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CertPolicyControllerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertPolicyControllerSpec) DeepCopyInto(out *CertPolicyControllerSpec) {
	*out = *in
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertPolicyControllerSpec.
func (in *CertPolicyControllerSpec) DeepCopy() *CertPolicyControllerSpec {
	if in == nil {
		return nil
	}
	out := new(CertPolicyControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertPolicyControllerStatus) DeepCopyInto(out *CertPolicyControllerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertPolicyControllerStatus.
func (in *CertPolicyControllerStatus) DeepCopy() *CertPolicyControllerStatus {
	if in == nil {
		return nil
	}
	out := new(CertPolicyControllerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectionManager) DeepCopyInto(out *ConnectionManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectionManager.
func (in *ConnectionManager) DeepCopy() *ConnectionManager {
	if in == nil {
		return nil
	}
	out := new(ConnectionManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConnectionManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectionManagerList) DeepCopyInto(out *ConnectionManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ConnectionManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectionManagerList.
func (in *ConnectionManagerList) DeepCopy() *ConnectionManagerList {
	if in == nil {
		return nil
	}
	out := new(ConnectionManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConnectionManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectionManagerSpec) DeepCopyInto(out *ConnectionManagerSpec) {
	*out = *in
	if in.BootStrapConfig != nil {
		in, out := &in.BootStrapConfig, &out.BootStrapConfig
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectionManagerSpec.
func (in *ConnectionManagerSpec) DeepCopy() *ConnectionManagerSpec {
	if in == nil {
		return nil
	}
	out := new(ConnectionManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectionManagerStatus) DeepCopyInto(out *ConnectionManagerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectionManagerStatus.
func (in *ConnectionManagerStatus) DeepCopy() *ConnectionManagerStatus {
	if in == nil {
		return nil
	}
	out := new(ConnectionManagerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Endpoint.
func (in *Endpoint) DeepCopy() *Endpoint {
	if in == nil {
		return nil
	}
	out := new(Endpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Endpoint) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointApplicationManagerSpec) DeepCopyInto(out *EndpointApplicationManagerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointApplicationManagerSpec.
func (in *EndpointApplicationManagerSpec) DeepCopy() *EndpointApplicationManagerSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointApplicationManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointCISControllerSpec) DeepCopyInto(out *EndpointCISControllerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointCISControllerSpec.
func (in *EndpointCISControllerSpec) DeepCopy() *EndpointCISControllerSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointCISControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointCertPolicyControllerSpec) DeepCopyInto(out *EndpointCertPolicyControllerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointCertPolicyControllerSpec.
func (in *EndpointCertPolicyControllerSpec) DeepCopy() *EndpointCertPolicyControllerSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointCertPolicyControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointConnectionManagerSpec) DeepCopyInto(out *EndpointConnectionManagerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointConnectionManagerSpec.
func (in *EndpointConnectionManagerSpec) DeepCopy() *EndpointConnectionManagerSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointConnectionManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointIAMPolicyControllerSpec) DeepCopyInto(out *EndpointIAMPolicyControllerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointIAMPolicyControllerSpec.
func (in *EndpointIAMPolicyControllerSpec) DeepCopy() *EndpointIAMPolicyControllerSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointIAMPolicyControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointList) DeepCopyInto(out *EndpointList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Endpoint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointList.
func (in *EndpointList) DeepCopy() *EndpointList {
	if in == nil {
		return nil
	}
	out := new(EndpointList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EndpointList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointPolicyControllerSpec) DeepCopyInto(out *EndpointPolicyControllerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointPolicyControllerSpec.
func (in *EndpointPolicyControllerSpec) DeepCopy() *EndpointPolicyControllerSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointPolicyControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointPrometheusIntegrationSpec) DeepCopyInto(out *EndpointPrometheusIntegrationSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointPrometheusIntegrationSpec.
func (in *EndpointPrometheusIntegrationSpec) DeepCopy() *EndpointPrometheusIntegrationSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointPrometheusIntegrationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointSearchCollectorSpec) DeepCopyInto(out *EndpointSearchCollectorSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointSearchCollectorSpec.
func (in *EndpointSearchCollectorSpec) DeepCopy() *EndpointSearchCollectorSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointSearchCollectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointServiceRegistrySpec) DeepCopyInto(out *EndpointServiceRegistrySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointServiceRegistrySpec.
func (in *EndpointServiceRegistrySpec) DeepCopy() *EndpointServiceRegistrySpec {
	if in == nil {
		return nil
	}
	out := new(EndpointServiceRegistrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointSpec) DeepCopyInto(out *EndpointSpec) {
	*out = *in
	if in.ClusterLabels != nil {
		in, out := &in.ClusterLabels, &out.ClusterLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.BootStrapConfig != nil {
		in, out := &in.BootStrapConfig, &out.BootStrapConfig
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.SearchCollectorConfig = in.SearchCollectorConfig
	out.PolicyController = in.PolicyController
	out.ServiceRegistryConfig = in.ServiceRegistryConfig
	out.ApplicationManagerConfig = in.ApplicationManagerConfig
	out.ConnectionManagerConfig = in.ConnectionManagerConfig
	out.CertPolicyControllerConfig = in.CertPolicyControllerConfig
	out.CISControllerConfig = in.CISControllerConfig
	out.IAMPolicyControllerConfig = in.IAMPolicyControllerConfig
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointSpec.
func (in *EndpointSpec) DeepCopy() *EndpointSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointStatus) DeepCopyInto(out *EndpointStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointStatus.
func (in *EndpointStatus) DeepCopy() *EndpointStatus {
	if in == nil {
		return nil
	}
	out := new(EndpointStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointWorkManagerSpec) DeepCopyInto(out *EndpointWorkManagerSpec) {
	*out = *in
	if in.ClusterLabels != nil {
		in, out := &in.ClusterLabels, &out.ClusterLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointWorkManagerSpec.
func (in *EndpointWorkManagerSpec) DeepCopy() *EndpointWorkManagerSpec {
	if in == nil {
		return nil
	}
	out := new(EndpointWorkManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GlobalValues) DeepCopyInto(out *GlobalValues) {
	*out = *in
	if in.ImageOverrides != nil {
		in, out := &in.ImageOverrides, &out.ImageOverrides
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GlobalValues.
func (in *GlobalValues) DeepCopy() *GlobalValues {
	if in == nil {
		return nil
	}
	out := new(GlobalValues)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IAMPolicyController) DeepCopyInto(out *IAMPolicyController) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IAMPolicyController.
func (in *IAMPolicyController) DeepCopy() *IAMPolicyController {
	if in == nil {
		return nil
	}
	out := new(IAMPolicyController)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IAMPolicyController) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IAMPolicyControllerList) DeepCopyInto(out *IAMPolicyControllerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IAMPolicyController, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IAMPolicyControllerList.
func (in *IAMPolicyControllerList) DeepCopy() *IAMPolicyControllerList {
	if in == nil {
		return nil
	}
	out := new(IAMPolicyControllerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IAMPolicyControllerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IAMPolicyControllerSpec) DeepCopyInto(out *IAMPolicyControllerSpec) {
	*out = *in
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IAMPolicyControllerSpec.
func (in *IAMPolicyControllerSpec) DeepCopy() *IAMPolicyControllerSpec {
	if in == nil {
		return nil
	}
	out := new(IAMPolicyControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IAMPolicyControllerStatus) DeepCopyInto(out *IAMPolicyControllerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IAMPolicyControllerStatus.
func (in *IAMPolicyControllerStatus) DeepCopy() *IAMPolicyControllerStatus {
	if in == nil {
		return nil
	}
	out := new(IAMPolicyControllerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyController) DeepCopyInto(out *PolicyController) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyController.
func (in *PolicyController) DeepCopy() *PolicyController {
	if in == nil {
		return nil
	}
	out := new(PolicyController)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolicyController) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyControllerList) DeepCopyInto(out *PolicyControllerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PolicyController, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyControllerList.
func (in *PolicyControllerList) DeepCopy() *PolicyControllerList {
	if in == nil {
		return nil
	}
	out := new(PolicyControllerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolicyControllerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyControllerSpec) DeepCopyInto(out *PolicyControllerSpec) {
	*out = *in
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyControllerSpec.
func (in *PolicyControllerSpec) DeepCopy() *PolicyControllerSpec {
	if in == nil {
		return nil
	}
	out := new(PolicyControllerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyControllerStatus) DeepCopyInto(out *PolicyControllerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyControllerStatus.
func (in *PolicyControllerStatus) DeepCopy() *PolicyControllerStatus {
	if in == nil {
		return nil
	}
	out := new(PolicyControllerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SearchCollector) DeepCopyInto(out *SearchCollector) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SearchCollector.
func (in *SearchCollector) DeepCopy() *SearchCollector {
	if in == nil {
		return nil
	}
	out := new(SearchCollector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SearchCollector) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SearchCollectorList) DeepCopyInto(out *SearchCollectorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SearchCollector, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SearchCollectorList.
func (in *SearchCollectorList) DeepCopy() *SearchCollectorList {
	if in == nil {
		return nil
	}
	out := new(SearchCollectorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SearchCollectorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SearchCollectorSpec) DeepCopyInto(out *SearchCollectorSpec) {
	*out = *in
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SearchCollectorSpec.
func (in *SearchCollectorSpec) DeepCopy() *SearchCollectorSpec {
	if in == nil {
		return nil
	}
	out := new(SearchCollectorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SearchCollectorStatus) DeepCopyInto(out *SearchCollectorStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SearchCollectorStatus.
func (in *SearchCollectorStatus) DeepCopy() *SearchCollectorStatus {
	if in == nil {
		return nil
	}
	out := new(SearchCollectorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceRegistry) DeepCopyInto(out *ServiceRegistry) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceRegistry.
func (in *ServiceRegistry) DeepCopy() *ServiceRegistry {
	if in == nil {
		return nil
	}
	out := new(ServiceRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceRegistry) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceRegistryList) DeepCopyInto(out *ServiceRegistryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceRegistry, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceRegistryList.
func (in *ServiceRegistryList) DeepCopy() *ServiceRegistryList {
	if in == nil {
		return nil
	}
	out := new(ServiceRegistryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceRegistryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceRegistrySpec) DeepCopyInto(out *ServiceRegistrySpec) {
	*out = *in
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceRegistrySpec.
func (in *ServiceRegistrySpec) DeepCopy() *ServiceRegistrySpec {
	if in == nil {
		return nil
	}
	out := new(ServiceRegistrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceRegistryStatus) DeepCopyInto(out *ServiceRegistryStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceRegistryStatus.
func (in *ServiceRegistryStatus) DeepCopy() *ServiceRegistryStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceRegistryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkManager) DeepCopyInto(out *WorkManager) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkManager.
func (in *WorkManager) DeepCopy() *WorkManager {
	if in == nil {
		return nil
	}
	out := new(WorkManager)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkManager) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkManagerIngress) DeepCopyInto(out *WorkManagerIngress) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkManagerIngress.
func (in *WorkManagerIngress) DeepCopy() *WorkManagerIngress {
	if in == nil {
		return nil
	}
	out := new(WorkManagerIngress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkManagerList) DeepCopyInto(out *WorkManagerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WorkManager, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkManagerList.
func (in *WorkManagerList) DeepCopy() *WorkManagerList {
	if in == nil {
		return nil
	}
	out := new(WorkManagerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkManagerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkManagerService) DeepCopyInto(out *WorkManagerService) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkManagerService.
func (in *WorkManagerService) DeepCopy() *WorkManagerService {
	if in == nil {
		return nil
	}
	out := new(WorkManagerService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkManagerSpec) DeepCopyInto(out *WorkManagerSpec) {
	*out = *in
	if in.ClusterLabels != nil {
		in, out := &in.ClusterLabels, &out.ClusterLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Service = in.Service
	out.Ingress = in.Ingress
	in.GlobalValues.DeepCopyInto(&out.GlobalValues)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkManagerSpec.
func (in *WorkManagerSpec) DeepCopy() *WorkManagerSpec {
	if in == nil {
		return nil
	}
	out := new(WorkManagerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkManagerStatus) DeepCopyInto(out *WorkManagerStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkManagerStatus.
func (in *WorkManagerStatus) DeepCopy() *WorkManagerStatus {
	if in == nil {
		return nil
	}
	out := new(WorkManagerStatus)
	in.DeepCopyInto(out)
	return out
}
