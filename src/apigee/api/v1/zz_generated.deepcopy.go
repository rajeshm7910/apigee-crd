// +build !ignore_autogenerated

/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiProduct) DeepCopyInto(out *ApiProduct) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiProduct.
func (in *ApiProduct) DeepCopy() *ApiProduct {
	if in == nil {
		return nil
	}
	out := new(ApiProduct)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApiProduct) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiProductList) DeepCopyInto(out *ApiProductList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ApiProduct, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiProductList.
func (in *ApiProductList) DeepCopy() *ApiProductList {
	if in == nil {
		return nil
	}
	out := new(ApiProductList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApiProductList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiProductSpec) DeepCopyInto(out *ApiProductSpec) {
	*out = *in
	if in.ApiResources != nil {
		in, out := &in.ApiResources, &out.ApiResources
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Attributes != nil {
		in, out := &in.Attributes, &out.Attributes
		*out = make([]AttributesSpec, len(*in))
		copy(*out, *in)
	}
	if in.Environments != nil {
		in, out := &in.Environments, &out.Environments
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Proxies != nil {
		in, out := &in.Proxies, &out.Proxies
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Scopes != nil {
		in, out := &in.Scopes, &out.Scopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiProductSpec.
func (in *ApiProductSpec) DeepCopy() *ApiProductSpec {
	if in == nil {
		return nil
	}
	out := new(ApiProductSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiProductStatus) DeepCopyInto(out *ApiProductStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiProductStatus.
func (in *ApiProductStatus) DeepCopy() *ApiProductStatus {
	if in == nil {
		return nil
	}
	out := new(ApiProductStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AttributesSpec) DeepCopyInto(out *AttributesSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AttributesSpec.
func (in *AttributesSpec) DeepCopy() *AttributesSpec {
	if in == nil {
		return nil
	}
	out := new(AttributesSpec)
	in.DeepCopyInto(out)
	return out
}