## Intro 
Klusterlet operator is a Go operator build with the operator-sdk that's use to manage the Create Update Delete of the component CR in the Klusterlet Component Operator https://github.ibm.com/IBMPrivateCloud/klusterlet-component-operator

## Running Klusterlet Operator locally for development
1. Install klusterlet CRD
```
make utils:crds:install
```

2. Create ImagePullSecret for Artifactory

3. Run Klusterlet Operator 
```
make operator:run
```

## Installing Klusterlet using Klusterlet Operator 
To create a klusterlet deployment with the klusterlet operator u need to create the klusterlet CR

Example of Klusterlet CR `/deploy/crds/klusterlet_v1alpha1_klusterletservice_cr.yaml`

## Build and publish a personal build to scratch artifactory
- `make init`
- `make operator:build`
- `make docker:tag`
- `make docker:push`
