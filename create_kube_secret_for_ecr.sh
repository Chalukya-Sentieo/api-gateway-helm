#!/bin/sh
kubectl create secret docker-registry regcred \
  --docker-server=602037364990.dkr.ecr.us-east-1.amazonaws.com \
  --docker-username=AWS \
  --docker-password=$(aws ecr get-login-password) \
  --namespace=kong
