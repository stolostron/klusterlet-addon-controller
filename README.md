## Intro 
Klusterlet operator is a Go operator build with the operator-sdk that's use to manage the Create Update Delete of the component CR in the Klusterlet Component Operator https://github.ibm.com/IBMPrivateCloud/klusterlet-component-operator

## Running Klusterlet Operator locally for development
1. Run the klusterlet component operator
https://github.ibm.com/IBMPrivateCloud/klusterlet-component-operator/README.md


2. Install klusterlet CRD
```
make install-crd
```

3. Run Klusterlet Operator 
```
make operator:run
```

## Installing Klusterlet using Klusterlet Operator 
To create a klusterlet deployment with the klusterlet operator u need to create the klusterlet CR

Example of Klusterlet CR `/deploy/crds/klusterlet_v1alpha1_klusterletservice_cr.yaml`