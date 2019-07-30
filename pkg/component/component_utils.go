//Package component Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package component

import (
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"

	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("component_utils")

// InstallComponentCRDs look for the component crd yaml in /opt/components and create/update the CRDs
func InstallComponentCRDs(cs *crdclientset.Clientset) error {
	log.Info("Installing component CRDs")

	componentCRDDirPath := "/opt/component-operator/deploy/crds"
	files, err := ioutil.ReadDir(componentCRDDirPath)
	if err != nil {
		log.Error(err, "Fail to Read Component CRD Directory", "dirname", componentCRDDirPath)
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), "crd.yaml") {
			crdFilePath := componentCRDDirPath + "/" + file.Name()
			log.V(1).Info("Found Component CRD Yaml", "file", crdFilePath)
			crdYaml, err := ioutil.ReadFile(crdFilePath)
			if err != nil {
				log.Error(err, "Fail to ReadFile", "filename", crdFilePath)
				return err
			}
			crd := &crdv1beta1.CustomResourceDefinition{}
			err = yaml.Unmarshal(crdYaml, crd)
			if err != nil {
				log.Error(err, "Fail to Unmarshal", "content", crdYaml)
				return err
			}
			createOrUpdateCRD(crd, cs)
		}
	}
	return nil
}

func createOrUpdateCRD(crd *crdv1beta1.CustomResourceDefinition, cs *crdclientset.Clientset) error {
	log.Info("Create or Update component CRD", "CRD.Name", crd.Name)

	log.V(1).Info("Looking for CRD", "CRD.Name", crd.Name)
	foundCRD, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(1).Info("Creating CRD", "CRD.Name", crd.Name)
			_, err = cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
			if err != nil {
				log.Error(err, "Fail to CREATE CRD", "CRD.Name", crd.Name)
				return err
			}
		} else {
			log.Error(err, "Unexpected error GET CRD", "CRD.Name", crd.Name)
			return err
		}
	} else {
		// NOTE: the UPDATE will always run since API server add additional stuff to the Spec but that's ok
		// 	However this does present a problem for when rolling back the version of klusterlet operator...
		//  If the newer version have a newer API than if we rollback to older version and it try to call Update
		//  the Update will fail
		log.V(1).Info("Updating CRD", "CRD.Name", crd.Name)
		foundCRD.Spec = crd.Spec
		_, err = cs.ApiextensionsV1beta1().CustomResourceDefinitions().Update(foundCRD)
		if err != nil {
			log.Error(err, "Fail to UPDATE CRD", "CRD.Name", foundCRD.Name)
			return err
		}
	}
	return nil
}
