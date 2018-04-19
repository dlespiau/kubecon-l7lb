#!/bin/sh

LABEL_SELECTOR='app=kubecon-service'
BASE_PORT=9000

i=0
for pod in `kubectl get pods -l${LABEL_SELECTOR} -ojsonpath='{.items[*].metadata.name}'`; do
    curl -s http://localhost:`expr $BASE_PORT + $i`/info
    i=`expr $i + 1`
done
