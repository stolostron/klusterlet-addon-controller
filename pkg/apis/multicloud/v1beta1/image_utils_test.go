// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package v1beta1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/open-cluster-management/endpoint-operator/version"
	"github.com/stretchr/testify/assert"
)

var manifestPath = filepath.Join("..", "..", "..", "..", "image-manifests", version.Version+".json")

func TestGetImage(t *testing.T) {
	os.Setenv("USE_SHA_MANIFEST", "false")
	err := LoadManifest(manifestPath)
	if err != nil {
		t.Error(err)
	}
	imageTagPostfixKey := "IMAGE_TAG_POSTFIX"
	type args struct {
		endpoint        *Endpoint
		component       string
		imageTagPostfix string
	}

	tests := []struct {
		name    string
		args    args
		want    GlobalValues
		wantErr bool
	}{
		{
			name: "Use Default Component Tag",
			args: args{
				endpoint: &Endpoint{
					Spec: EndpointSpec{
						ImageRegistry: "sample-registry/uniquePath",
					},
				},
				component:       "component-operator",
				imageTagPostfix: "",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator:1.0.0",
				},
			},
			wantErr: false,
		},
		{
			name: "Not Exists Component",
			args: args{
				endpoint:        &Endpoint{},
				component:       "notExistsComponent",
				imageTagPostfix: "",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
		{
			name: "With Postfix Set",
			args: args{
				endpoint: &Endpoint{
					Spec: EndpointSpec{
						ImageRegistry: "sample-registry-2/uniquePath",
					},
				},
				component:       "connection-manager",
				imageTagPostfix: "-aUnique-Post-Fix",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"multicloud_manager": "sample-registry-2/uniquePath/multicloud-manager:0.0.1-aUnique-Post-Fix",
				},
			},
			wantErr: false,
		},
		{
			name: "Use Component Image Tag",
			args: args{
				endpoint: &Endpoint{
					Spec: EndpointSpec{
						ImageRegistry: "sample-registry/uniquePath",
					},
				},
				component:       "connection-manager",
				imageTagPostfix: "-aUnique-Post-Fix",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"multicloud_manager": "sample-registry/uniquePath/multicloud-manager:0.0.1-aUnique-Post-Fix",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running tests %s", tt.name)
			err := os.Setenv(imageTagPostfixKey, tt.args.imageTagPostfix)
			if err != nil {
				t.Errorf("Cannot set env %s", imageTagPostfixKey)
			}
			imgKey, imgRepository, err := tt.args.endpoint.GetImage(tt.args.component)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly.")
			} else if !tt.wantErr {
				assert.Equal(t, tt.want.ImageOverrides[imgKey], imgRepository, "repository should match")
			}
		})
	}
}

func TestGetImageWithManifest(t *testing.T) {
	os.Setenv("USE_SHA_MANIFEST", "true")
	os.Setenv("IMAGE_TAG_POSTFIX", "")
	err := LoadManifest(manifestPath)
	if err != nil {
		t.Error(err)
	}
	type args struct {
		endpoint  *Endpoint
		component string
	}

	tests := []struct {
		name    string
		args    args
		want    GlobalValues
		wantErr bool
	}{
		{
			name: "Use Component Sha",
			args: args{
				endpoint: &Endpoint{
					Spec: EndpointSpec{
						ImageRegistry: "sample-registry/uniquePath",
					},
				},
				component: "component-operator",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256",
				},
			},
			wantErr: false,
		},
		{
			name: "Not Exists Component",
			args: args{
				endpoint:  &Endpoint{},
				component: "notExistsComponent",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running tests %s", tt.name)
			imgKey, imgRepository, err := tt.args.endpoint.GetImage(tt.args.component)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly. Error:%s", err)
			} else if !tt.wantErr {
				assert.Equal(t, tt.want.ImageOverrides[imgKey], imgRepository, "repository should match")
			}
		})
	}
}
