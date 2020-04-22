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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/open-cluster-management/endpoint-operator/pkg/image"
	"github.com/open-cluster-management/endpoint-operator/version"
	"github.com/stretchr/testify/assert"
)

func TestGetImage(t *testing.T) {
	os.Setenv("USE_SHA_MANIFEST", "false")
	err := loadTestManifest()
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
		want    image.Image
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
			want: image.Image{
				Repository: "sample-registry/uniquePath",
				Name:       defaultComponentImageMap["component-operator"],
				Tag:        defaultComponentTagMap["component-operator"],
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
			want:    image.Image{},
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
			want: image.Image{
				Repository: "sample-registry-2/uniquePath",
				Name:       defaultComponentImageMap["connection-manager"],
				Tag:        defaultComponentTagMap["connection-manager"],
				TagPostfix: "aUnique-Post-Fix",
			},
			wantErr: false,
		},
		{
			name: "Use Component Image Tag",
			args: args{
				endpoint: &Endpoint{
					Spec: EndpointSpec{
						ImageRegistry: "sample-registry/uniquePath",
						ComponentsImagesTag: map[string]string{
							"connection-manager": "some-special-version-tag",
						},
					},
				},
				component:       "connection-manager",
				imageTagPostfix: "-aUnique-Post-Fix",
			},
			want: image.Image{
				Repository: "sample-registry/uniquePath",
				Name:       defaultComponentImageMap["connection-manager"],
				Tag:        "some-special-version-tag",
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
			var imageShaDigests map[string]string
			img, _, err := tt.args.endpoint.GetImage(tt.args.component, imageShaDigests)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly.")
			} else if !tt.wantErr {
				assert.Equal(t, tt.want.Repository, img.Repository, "repository should match")
				assert.Equal(t, tt.want.Name, img.Name, "image name should match")
				assert.Equal(t, tt.want.Tag, img.Tag, "image tag should match")
			}
		})
	}
}

func TestGetImageWithManifest(t *testing.T) {
	os.Setenv("USE_SHA_MANIFEST", "true")
	os.Setenv("IMAGE_TAG_POSTFIX", "")
	err := loadTestManifest()
	if err != nil {
		t.Error(err)
	}
	type args struct {
		endpoint        *Endpoint
		component       string
		imageTagPostfix string
	}

	tests := []struct {
		name    string
		args    args
		want    image.Image
		sha     string
		shaKey  string
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
				component:       "component-operator",
				imageTagPostfix: "hello",
			},
			want: image.Image{
				Repository: "sample-registry/uniquePath",
				Name:       defaultComponentImageMap["component-operator"],
				Tag:        defaultComponentTagMap["component-operator"],
				TagPostfix: "",
			},
			sha:     "sha256:b3edec494a5c9f5a9bf65699d0592ca2e50c205132f5337e8df07a7808d03887",
			shaKey:  "endpoint_component_operator",
			wantErr: false,
		},
		{
			name: "Not Exists Component",
			args: args{
				endpoint:        &Endpoint{},
				component:       "notExistsComponent",
				imageTagPostfix: "",
			},
			want:    image.Image{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running tests %s", tt.name)
			imageShaDigests := make(map[string]string)
			img, sha, err := tt.args.endpoint.GetImage(tt.args.component, imageShaDigests)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly. Error:%s", err)
			} else if !tt.wantErr {
				assert.Equal(t, tt.want.Repository, img.Repository, "repository should match")
				assert.Equal(t, tt.want.Name, img.Name, "image name should match")
				assert.Equal(t, tt.want.Tag, img.Tag, "image tag should match")
				assert.Equal(t, tt.want.TagPostfix, img.TagPostfix, "image tag should match")
				sha, ok := sha[tt.shaKey]
				assert.True(t, ok)
				assert.Equal(t, tt.sha, sha, "image sha should match")
			}
		})
	}
}

func loadTestManifest() error {
	Manifest.Images = make([]imageManifest, 0)
	filePath := filepath.Join("..", "..", "..", "..", "image-manifests", version.Version+".json")
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, &Manifest.Images)
	if err != nil {
		return err
	}
	return nil
}
