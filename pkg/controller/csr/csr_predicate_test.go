// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package csr

import (
	"testing"

	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func Test_newCSRPredicate(t *testing.T) {
	testClusterName := "cluster1"

	tests := []struct {
		name     string
		csr      *certificatesv1.CertificateSigningRequest
		expected bool
	}{
		{
			name: "csr without labels",
			csr:  &certificatesv1.CertificateSigningRequest{},
		},
		{
			name: "csr without cluster name label",
			csr: &certificatesv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						managedClusterAddonNameLabel: "application-manager",
					},
				},
			},
		},
		{
			name: "csr without addon name label",
			csr: &certificatesv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterNameLabel: testClusterName,
					},
				},
			},
		},
		{
			name: "invalid signer",
			csr: &certificatesv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterNameLabel:             testClusterName,
						managedClusterAddonNameLabel: "application-manager",
					},
				},
				Spec: certificatesv1.CertificateSigningRequestSpec{
					SignerName: "example.com/signer1",
				},
			},
		},
		{
			name: "invalid requester",
			csr: &certificatesv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterNameLabel:             testClusterName,
						managedClusterAddonNameLabel: "application-manager",
					},
				},
				Spec: certificatesv1.CertificateSigningRequestSpec{
					SignerName: certificatesv1.KubeAPIServerClientSignerName,
					Username:   "anonymous",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := event.CreateEvent{
				Meta:   tt.csr.GetObjectMeta(),
				Object: tt.csr,
			}
			actual := newCSRPredicate().Create(e)
			if actual != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}
