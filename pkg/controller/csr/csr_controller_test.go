// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package csr

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileCSR_Reconcile(t *testing.T) {
	testClusterName := "cluster1"
	testAddonName := "application-manager"
	testCSRName := "csr1"
	testRequester := fmt.Sprintf("system:open-cluster-management:%s:agent1", testClusterName)

	testscheme := scheme.Scheme
	_ = managedclusterv1.AddToScheme(testscheme)
	_ = addonv1alpha1.AddToScheme(testscheme)

	managedClusterAddOn := &addonv1alpha1.ManagedClusterAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testClusterName,
			Name:      testAddonName,
		},
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: testCSRName,
		},
	}

	tests := []struct {
		name     string
		initObjs []runtime.Object
		csr      *certificatesv1.CertificateSigningRequest
		want     reconcile.Result
		approved bool
	}{
		{
			name: "denied csr",
			csr: newCSR(testCSRName, testAddonName, testClusterName, testRequester, []certificatesv1.CertificateSigningRequestCondition{
				{Type: certificatesv1.CertificateDenied},
			}),
			want: reconcile.Result{Requeue: false},
		},
		{
			name: "no addon",
			initObjs: []runtime.Object{
				&managedclusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: testClusterName,
					},
					Status: managedclusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{
							{
								Type:   managedclusterv1.ManagedClusterConditionHubAccepted,
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			csr:  newCSR(testCSRName, testAddonName, testClusterName, testRequester, nil),
			want: reconcile.Result{Requeue: false},
		},
		{
			name: "created by invalid requestor",
			initObjs: []runtime.Object{
				managedClusterAddOn,
				&managedclusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: testClusterName,
					},
					Status: managedclusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{
							{
								Type:   managedclusterv1.ManagedClusterConditionHubAccepted,
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			csr:  newCSR(testCSRName, testAddonName, testClusterName, "invalid-requester", nil),
			want: reconcile.Result{},
		},
		{
			name: "cluster is not joined yet",
			initObjs: []runtime.Object{
				managedClusterAddOn,
				&managedclusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: testClusterName,
					},
				},
			},
			csr:  newCSR(testCSRName, testAddonName, testClusterName, testRequester, nil),
			want: reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second},
		},
		{
			name: "approve csr",
			initObjs: []runtime.Object{
				managedClusterAddOn,
				&managedclusterv1.ManagedCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name: testClusterName,
					},
					Status: managedclusterv1.ManagedClusterStatus{
						Conditions: []metav1.Condition{
							{
								Type:   managedclusterv1.ManagedClusterConditionHubAccepted,
								Status: metav1.ConditionTrue,
							},
						},
					},
				},
			},
			csr:      newCSR(testCSRName, testAddonName, testClusterName, testRequester, nil),
			want:     reconcile.Result{},
			approved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeKubeClient := kubefake.NewSimpleClientset(tt.csr)
			reconcileCSR := &ReconcileCSR{
				client:    fake.NewFakeClientWithScheme(testscheme, append(tt.initObjs, tt.csr)...),
				scheme:    testscheme,
				csrClient: fakeKubeClient.CertificatesV1().CertificateSigningRequests(),
			}

			actual, err := reconcileCSR.Reconcile(request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("expected %v but got %v", tt.want, actual)
			}

			if !tt.approved {
				return
			}

			csr, err := fakeKubeClient.CertificatesV1().CertificateSigningRequests().Get(context.Background(), testCSRName, metav1.GetOptions{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for _, condition := range csr.Status.Conditions {
				if condition.Type == certificatesv1.CertificateApproved {
					return
				}
			}

			t.Errorf("csr should have been approved")
		})
	}
}

func Test_isValidAddonCSR(t *testing.T) {
	testClusterName := "cluster1"
	testAddonName := "application-manager"
	testOrganization := fmt.Sprintf("system:open-cluster-management:cluster:%s:addon:%s", testClusterName, testAddonName)

	tests := []struct {
		name     string
		csr      *certificatesv1.CertificateSigningRequest
		expected bool
	}{
		{
			name: "invalid request",
			csr: &certificatesv1.CertificateSigningRequest{
				Spec: certificatesv1.CertificateSigningRequestSpec{
					Request: []byte("invalid csr"),
				},
			},
		},
		{
			name: "invalid organization",
			csr: &certificatesv1.CertificateSigningRequest{
				Spec: certificatesv1.CertificateSigningRequestSpec{
					Request: newCSRRequestData("user1", []string{"org1"}),
				},
			},
		},
		{
			name: "invalid common name",
			csr: &certificatesv1.CertificateSigningRequest{
				Spec: certificatesv1.CertificateSigningRequestSpec{
					Request: newCSRRequestData("user1", []string{
						testOrganization,
					}),
				},
			},
		},
		{
			name: "valid csr",
			csr: &certificatesv1.CertificateSigningRequest{
				Spec: certificatesv1.CertificateSigningRequestSpec{
					Request: newCSRRequestData(
						fmt.Sprintf("%s:agent:agent1", testOrganization), []string{
							testOrganization,
						}),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidAddonCSR(tt.csr, testAddonName, testClusterName)
			if actual != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, actual)
			}
		})
	}
}

func newCSR(csrName, addonName, clusterName, requester string, conditions []certificatesv1.CertificateSigningRequestCondition) *certificatesv1.CertificateSigningRequest {

	organization := fmt.Sprintf("system:open-cluster-management:cluster:%s:addon:%s", clusterName, addonName)
	commonName := fmt.Sprintf("%s:agent:agent1", organization)

	return &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: csrName,
			Labels: map[string]string{
				clusterNameLabel:             clusterName,
				managedClusterAddonNameLabel: addonName,
			},
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request:  newCSRRequestData(commonName, []string{organization}),
			Username: requester,
		},
		Status: certificatesv1.CertificateSigningRequestStatus{
			Conditions: conditions,
		},
	}
}

func newCSRRequestData(commonName string, organization []string) []byte {
	insecureRand := rand.New(rand.NewSource(0))
	pk, err := ecdsa.GenerateKey(elliptic.P256(), insecureRand)
	if err != nil {
		panic(err)
	}
	csrb, err := x509.CreateCertificateRequest(insecureRand, &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: organization,
		},
		DNSNames:       []string{},
		EmailAddresses: []string{},
		IPAddresses:    []net.IP{},
	}, pk)
	if err != nil {
		panic(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrb})
}
