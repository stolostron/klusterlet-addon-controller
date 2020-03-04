//Package clusterimport ...
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
	"fmt"
	"os"
	"testing"

	"github.com/open-cluster-management/endpoint-operator/pkg/image"
	"github.com/stretchr/testify/assert"
)

func TestGetImage(t *testing.T) {
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
				Repository: fmt.Sprintf("sample-registry/uniquePath/%s", defaultComponentImageMap["component-operator"]),
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
				imageTagPostfix: "aUnique-Post-Fix",
			},
			want: image.Image{
				Repository: fmt.Sprintf("sample-registry-2/uniquePath/%s", defaultComponentImageMap["connection-manager"]),
				Tag:        fmt.Sprintf("%s-aUnique-Post-Fix", defaultComponentTagMap["connection-manager"]),
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
				imageTagPostfix: "",
			},
			want: image.Image{
				Repository: fmt.Sprintf("sample-registry/uniquePath/%s", defaultComponentImageMap["connection-manager"]),
				Tag:        "some-special-version-tag",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(imageTagPostfixKey, tt.args.imageTagPostfix)
			if err != nil {
				t.Errorf("Cannot set env %s", imageTagPostfixKey)
			}
			img, err := tt.args.endpoint.GetImage(tt.args.component)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly.")
			} else if !tt.wantErr {
				assert.Equal(t, img.Repository, tt.want.Repository, "image name should match")
				assert.Equal(t, img.Tag, tt.want.Tag, "image tag should match")
			}
		})
	}
}
