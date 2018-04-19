#!/bin/bash

LABEL_SELECTOR='app=kubecon-service'
BASE_PORT=9000

i=0
for pod in `kubectl get pods -l${LABEL_SELECTOR} -ojsonpath='{.items[*].metadata.name}'`; do
    kubectl port-forward $pod `expr $BASE_PORT + $i`:8080 >/dev/null 2>/dev/null &
    i=`expr $i + 1`
done
