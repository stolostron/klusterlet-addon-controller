// Copyright (c) 2020 Red Hat, Inc.

package webhook

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
)

var manifestPath = filepath.Join("..", "..", "image-manifests")

var klusterletaddonconfigsSchema = metav1.GroupVersionResource{
	Group:    "agent.open-cluster-management.io",
	Version:  "v1",
	Resource: "klusterletaddonconfigs",
}

func TestKlusterletAddonConfigValidate(t *testing.T) {
	err := agentv1.LoadManifests(manifestPath)
	if err != nil {
		t.Error(err)
	}
	cases := []struct {
		name             string
		request          admission.Request
		expectedResponse admission.Response
	}{
		{
			name: "validate deleting operation",
			request: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Resource:  klusterletaddonconfigsSchema,
					Operation: admissionv1beta1.Delete,
				},
			},
			expectedResponse: admission.Allowed(""),
		},
		{
			name: "validate creating klusterletaddonconfig",
			request: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Resource:  klusterletaddonconfigsSchema,
					Operation: admissionv1beta1.Create,
					Object:    newKlusterletAddonConfigObj(),
				},
			},
			expectedResponse: admission.Allowed(""),
		},
		{
			name: "validate creating klusterletaddonconfig with invalid version",
			request: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Resource:  klusterletaddonconfigsSchema,
					Operation: admissionv1beta1.Create,
					Object:    newKlusterletAddonConfigObjWithVersion("2.0.0.12"),
				},
			},
			expectedResponse: admission.Denied("Version \"2.0.0.12\" is invalid semantic version"),
		},
		{
			name: "validate creating klusterletaddonconfig with unavailable version",
			request: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Resource:  klusterletaddonconfigsSchema,
					Operation: admissionv1beta1.Create,
					Object:    newKlusterletAddonConfigObjWithVersion("2.2.0"),
				},
			},
			expectedResponse: admission.Denied("Version 2.2.0 is not available. Available Versions are: [2.0.0 2.1.0]"),
		},
		{
			name: "validate updating version",
			request: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Resource:  klusterletaddonconfigsSchema,
					Operation: admissionv1beta1.Update,
					OldObject: newKlusterletAddonConfigObjWithVersion("2.0.0"),
					Object:    newKlusterletAddonConfigObjWithVersion("2.1.0"),
				},
			},
			expectedResponse: admission.Allowed(""),
		},
		{
			name: "validate updating not available version",
			request: admission.Request{
				AdmissionRequest: admissionv1beta1.AdmissionRequest{
					Resource:  klusterletaddonconfigsSchema,
					Operation: admissionv1beta1.Update,
					OldObject: newKlusterletAddonConfigObjWithVersion("2.0.0"),
					Object:    newKlusterletAddonConfigObjWithVersion("1.0.0"),
				},
			},
			expectedResponse: admission.Denied("Version 1.0.0 is not available. Available Versions are: [2.0.0 2.1.0]"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kubeClient := kubefake.NewSimpleClientset()
			kubeClient.PrependReactor(
				"create",
				"subjectaccessreviews",
				func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &authorizationv1.SubjectAccessReview{
						Status: authorizationv1.SubjectAccessReviewStatus{
							Allowed: true,
						},
					}, nil
				},
			)

			admissionHook := &klusterletAddonConfigValidator{}
			var ctx context.Context
			actualResponse := admissionHook.Handle(ctx, c.request)

			if !reflect.DeepEqual(actualResponse, c.expectedResponse) {
				t.Errorf("expected %#v but got: %#v", c.expectedResponse, actualResponse)
			}
		})
	}
}

func newKlusterletAddonConfigObj() runtime.RawExtension {
	klusterletAddonConf := &agentv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-klusterletaddonconfig",
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			Version: "2.0.0",
		},
	}
	klusterletAddonConfObj, _ := json.Marshal(klusterletAddonConf)
	return runtime.RawExtension{
		Raw: klusterletAddonConfObj,
	}
}

func newKlusterletAddonConfigObjWithVersion(invalidVersion string) runtime.RawExtension {
	klusterletAddonConf := &agentv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-klusterletaddonconfig",
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			Version: invalidVersion,
		},
	}
	klusterletAddonConfObj, _ := json.Marshal(klusterletAddonConf)
	return runtime.RawExtension{
		Raw: klusterletAddonConfObj,
	}
}
