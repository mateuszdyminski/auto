# auto
Kubernetes Auto-Scaling presentation and code samples

### Demo Auto-Scaling 

Start minikube:
```
minikube start
```

Install Elasticsearch
```
cd kube 
kubectl apply -f es/es-namespace.yaml
kubectl apply -f es
```

Install NATS 
```
cd kube
kubectl apply -f nats/nats-operator.yaml
kubectl apply -f nats/deployment.yaml
```

Install Metrics Server 
```
cd kube
kubectl apply -f metrics-server
```

### TBD
