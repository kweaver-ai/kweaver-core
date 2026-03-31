#!/bin/bash
set -xv

# 暴漏mariadb端口
kubectl delete svc/test -n resource
kubectl -n resource expose svc/mariadb-mariadb-master --port=3306 --target-port=3306 --type=NodePort --name test
kubectl -n resource patch svc/test --type='json' -p='[{"op": "replace", "path": "/spec/ports/0/nodePort", "value":30036}]'
firewall-cmd --add-port=30036/tcp --permanent

# 节点打标签
host=$(grep "host =" /root/AT/agent-AT/config/env.ini | awk -F'=' '{print $2}' | tr -d ' ')
host_name=$(kubectl get nodes -o wide | grep $host | awk '{print $1}')
kubectl label nodes ${host_name} kubernetes.io/testTag=AT --overwrite

# 执行AT
helm3 upgrade --install --create-namespace -n anyshare agent-at /root/agent-AT/helm