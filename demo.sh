#!/bin/sh

minikube profile kubecon
minikube start
kubectl apply -f kubernetes/service-deploy.yaml
kubectl apply -f kubernetes/service-svc.yaml
kubectl apply -f kubernetes/proxy-deploy.yaml
kubectl apply -f kubernetes/proxy-svc.yaml

echo
echo "Service URL: $(minikube service --url kubecon-service)"
echo "Proxy URL  : $(minikube service --url kubecon-proxy)"
