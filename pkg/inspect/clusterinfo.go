/*
 * IBM Confidential
 * OCO Source Materials
 * 5737-E67
 * (C) Copyright IBM Corporation 2018 All Rights Reserved
 * The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
 */

package inspect

import (
	"context"

	mcmv1alpha1 "github.ibm.com/IBMPrivateCloud/hcm-api/pkg/apis/mcm/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeployedOnHub checks that: Is this cluster a Hub Cluster?
func DeployedOnHub(c client.Client) bool {
	clusterStatusList := &mcmv1alpha1.ClusterStatusList{}
	err := c.List(context.TODO(), &client.ListOptions{}, clusterStatusList)
	return err != nil
}
