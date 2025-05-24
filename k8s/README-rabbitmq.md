# RabbitMQ Setup with Kubernetes Operator

## Installation Steps

1. **Install the RabbitMQ Cluster Operator:**
   ```bash
   kubectl apply -f "https://github.com/rabbitmq/cluster-operator/releases/latest/download/cluster-operator.yml"
   ```

2. **Deploy RabbitMQ Cluster:**
   ```bash
   kubectl apply -f k8s/rabbitmq-operator.yaml
   ```

3. **Verify Installation:**
   ```bash
   # Check operator is running
   kubectl get pods -l app.kubernetes.io/name=rabbitmq-cluster-operator
   
   # Check RabbitMQ cluster
   kubectl get rabbitmqclusters
   kubectl get pods -l app.kubernetes.io/name=rabbitmq
   ```

## Service Names

The operator creates the following services:
- **AMQP Service:** `rabbitmq` (port 5672)
- **Management UI:** `rabbitmq-management` (port 15672)

## Configuration Update Needed

Update your application configuration to use:
- **RABHOST:** `rabbitmq` (not `hello-world`)
- **RABPORT:** `5672`
- **RABUSER:** `guest`
- **RABPASS:** `guest`

## Access Management UI

```bash
# Port forward to access management UI
kubectl port-forward service/rabbitmq-management 15672:15672

# Then visit: http://localhost:15672 (guest/guest)
```