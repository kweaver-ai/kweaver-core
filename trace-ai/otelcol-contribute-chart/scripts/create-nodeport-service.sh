#!/bin/sh

set -eu

RELEASE_NAME="${1:-otelcol-contrib}"
NAMESPACE="${2:-observability}"

SERVICE_NAME="${SERVICE_NAME:-${RELEASE_NAME}-nodeport}"
APP_NAME="${APP_NAME:-otelcol-contrib}"
OTLP_GRPC_PORT="${OTLP_GRPC_PORT:-4317}"
OTLP_HTTP_PORT="${OTLP_HTTP_PORT:-4318}"
OTLP_GRPC_NODE_PORT="${OTLP_GRPC_NODE_PORT:-30417}"
OTLP_HTTP_NODE_PORT="${OTLP_HTTP_NODE_PORT:-30418}"

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: ${SERVICE_NAME}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/name: ${APP_NAME}
    app.kubernetes.io/instance: ${RELEASE_NAME}
spec:
  type: NodePort
  selector:
    app.kubernetes.io/name: ${APP_NAME}
    app.kubernetes.io/instance: ${RELEASE_NAME}
  ports:
    - name: otlp-grpc
      protocol: TCP
      port: ${OTLP_GRPC_PORT}
      targetPort: otlp-grpc
      nodePort: ${OTLP_GRPC_NODE_PORT}
    - name: otlp-http
      protocol: TCP
      port: ${OTLP_HTTP_PORT}
      targetPort: otlp-http
      nodePort: ${OTLP_HTTP_NODE_PORT}
EOF

echo "NodePort service applied:"
echo "  namespace: ${NAMESPACE}"
echo "  service:   ${SERVICE_NAME}"
echo "  grpc:      <node-ip>:${OTLP_GRPC_NODE_PORT}"
echo "  http:      <node-ip>:${OTLP_HTTP_NODE_PORT}"
echo
echo "telemetrygen examples:"
echo "  telemetrygen traces --otlp-insecure --otlp-endpoint <node-ip>:${OTLP_GRPC_NODE_PORT}"
echo "  telemetrygen metrics --otlp-insecure --otlp-endpoint <node-ip>:${OTLP_GRPC_NODE_PORT}"
