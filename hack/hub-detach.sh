#!/bin/bash
###############################################################################
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

if [ -z $1 ]; then
  echo "please set the cluster name"
  exit 1
fi
CLUSTERNAME=$1

function destroyOrDetach {
  oc annotate klusterletaddonconfig -n ${CLUSTERNAME} ${CLUSTERNAME} klusterletaddonconfig-pause=true --overwrite=true
  oc delete manifestwork -n ${CLUSTERNAME} ${CLUSTERNAME}-appmgr --wait=false
  oc delete manifestwork -n ${CLUSTERNAME} ${CLUSTERNAME}-certpolicyctrl --wait=false
  oc delete manifestwork -n ${CLUSTERNAME} ${CLUSTERNAME}-cispolicyctrl --wait=false
  oc delete manifestwork -n ${CLUSTERNAME} ${CLUSTERNAME}-iampolicyctrl --wait=false
  oc delete manifestwork -n ${CLUSTERNAME} ${CLUSTERNAME}-policyctrl --wait=false
  oc delete manifestwork -n ${CLUSTERNAME} ${CLUSTERNAME}-search --wait=false
  oc delete manifestwork -n ${CLUSTERNAME} ${CLUSTERNAME}-workmgr --wait=false
  sleep 60
  oc delete klusterletaddonconfig --wait=false -n ${CLUSTERNAME} ${CLUSTERNAME}
  oc annotate klusterletaddonconfig -n ${CLUSTERNAME} ${CLUSTERNAME} klusterletaddonconfig-pause=false --overwrite=true
  sleep 30
  
  oc get clusterdeployment -n  ${CLUSTERNAME} ${CLUSTERNAME}
  if [ $? -eq 0 ]; then
    echo "Detected clusterdeployment. Destroying"
    oc delete clusterdeployment -n  ${CLUSTERNAME} ${CLUSTERNAME} --wait=false
    for i in `seq 1 180`; do
      echo "waiting clusterdeployment to be deleted"
      sleep 20
      oc get clusterdeployment -n  ${CLUSTERNAME} ${CLUSTERNAME} || break ; 
    done
  fi
  
  oc delete managedcluster ${CLUSTERNAME} --wait=false
  sleep 60
  oc patch managedcluster ${CLUSTERNAME} -p '{"metadata":{"finalizers":[]}}' --type=merge
  exit
}

if [ ${CLUSTERNAME} = 'all' ]; then
  pids=()
  for c in `oc get managedcluster -otemplate --template='{{range .items}}{{printf "%s\n" .metadata.name}}{{end}}'` ; do 
    CLUSTERNAME=${c}
    echo destroyOrDetach ${CLUSTERNAME}
    destroyOrDetach &
    pids+=($!)
  done
  for pid in ${pids[*]}; do
    wait $pid
  done
else
  destroyOrDetach
fi
