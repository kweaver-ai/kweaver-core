
generate_config_yaml() {
    log_info "Generating config.yaml..."
    local out="${CONFIG_YAML_PATH}"
    mkdir -p "$(dirname "${out}")"

    load_image_registry_from_config

    local cfg_namespace="kweaver"
    local cfg_lang="en_US.UTF-8"
    local cfg_tz="Asia/Shanghai"
    if [[ -f "${out}" ]]; then
        local v
        v="$(awk '$1=="namespace:"{print $2; exit}' "${out}" 2>/dev/null | sed -e 's/^["'\'']//; s/["'\'']$//' || true)"
        if [[ -n "${v}" ]]; then cfg_namespace="${v}"; fi
        v="$(awk '$1=="env:"{in=1; next} in && $1=="language:"{print $2; exit} in && $0~/^[^ ]/{in=0}' "${out}" 2>/dev/null | sed -e 's/^["'\'']//; s/["'\'']$//' || true)"
        if [[ -n "${v}" ]]; then cfg_lang="${v}"; fi
        v="$(awk '$1=="env:"{in=1; next} in && $1=="timezone:"{print $2; exit} in && $0~/^[^ ]/{in=0}' "${out}" 2>/dev/null | sed -e 's/^["'\'']//; s/["'\'']$//' || true)"
        if [[ -n "${v}" ]]; then cfg_tz="${v}"; fi
    fi

    local node_ip
    # Try to get the first non-loopback IP address
    node_ip="$(hostname -I 2>/dev/null | tr ' ' '\n' | grep -v '^127\.' | head -1 | tr -d '\n' || true)"
    # If no valid IP found, try alternative methods
    if [[ -z "${node_ip}" ]] || [[ "${node_ip}" == "127.0.0.1" ]]; then
        # Try to get IP from ip command (more reliable)
        node_ip="$(ip addr show 2>/dev/null | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | grep -v '^127\.' | head -1 || true)"
    fi
    # Final fallback
    if [[ -z "${node_ip}" ]]; then
        node_ip="10.x.x.x"
    fi

    # Storage (local-path)
    local storage_class_name="${STORAGE_STORAGE_CLASS_NAME}"
    if [[ -z "${storage_class_name}" ]]; then
        if kubectl get storageclass local-path >/dev/null 2>&1; then
            storage_class_name="local-path"
        fi
    fi
    local storage_section=""
    if [[ -n "${storage_class_name}" ]]; then
        storage_section=$(cat <<STORAGE_EOF
storage:
  storageClassName: $(yaml_quote "${storage_class_name}")
STORAGE_EOF
)
    fi

    # MariaDB
    local mariadb_ns="${MARIADB_NAMESPACE}"
    local mariadb_host="mariadb-proton-mariadb.${mariadb_ns}.svc.cluster.local"
    # Use values from environment or config.yaml if set
    local mariadb_user="${MARIADB_USER:-adp}"
    local mariadb_password="${MARIADB_PASSWORD}"
    local mariadb_root_password="${MARIADB_ROOT_PASSWORD}"
    local mariadb_database="${MARIADB_DATABASE:-adp}"
    local mariadb_configured=false

    # Try to find MariaDB secret by label first (more reliable than hardcoded name)
    # The proton-mariadb chart creates a secret named mariadb-proton-mariadb-auth
    local mariadb_secret
    mariadb_secret="$(kubectl -n "${mariadb_ns}" get secret -l app.kubernetes.io/instance=mariadb,app=mariadb -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
    if [[ -n "${mariadb_secret}" ]]; then
        mariadb_configured=true
        local from_secret
        from_secret="$(get_secret_b64_key "${mariadb_ns}" "${mariadb_secret}" mariadb-user 2>/dev/null || echo "")"
        if [[ -n "${from_secret}" ]]; then
            mariadb_user="${from_secret}"
        fi
        from_secret="$(get_secret_b64_key "${mariadb_ns}" "${mariadb_secret}" mariadb-password 2>/dev/null || echo "")"
        if [[ -n "${from_secret}" ]]; then
            mariadb_password="${from_secret}"
        fi
        from_secret="$(get_secret_b64_key "${mariadb_ns}" "${mariadb_secret}" mariadb-root-password 2>/dev/null || echo "")"
        if [[ -n "${from_secret}" ]]; then
            mariadb_root_password="${from_secret}"
        fi
        from_secret="$(get_secret_b64_key "${mariadb_ns}" "${mariadb_secret}" mariadb-database 2>/dev/null || echo "")"
        if [[ -n "${from_secret}" ]]; then
            mariadb_database="${from_secret}"
        fi
    fi
    # Generate admin_key: base64 encoded "user:password"
    local mariadb_admin_key
    mariadb_admin_key="$(printf '%s:%s' "${mariadb_user}" "${mariadb_password}" | base64 -w 0 2>/dev/null || printf '%s:%s' "${mariadb_user}" "${mariadb_password}" | base64 | tr -d '\n')"

    # Redis
    local redis_ns="${REDIS_NAMESPACE}"
    # Try to detect Redis release name (could be "redis" or "proton-redis" depending on chart)
    local redis_release_name="redis"
    # Check for proton-redis release first
    if helm list -n "${redis_ns}" -q 2>/dev/null | grep -q "^proton-redis"; then
        redis_release_name="proton-redis"
    elif helm list -n "${redis_ns}" -q 2>/dev/null | grep -q "^redis"; then
        redis_release_name="redis"
    fi
    
    # Try to get actual StatefulSet name (could be redis, redis-proton-redis, or proton-redis-proton-redis)
    # For proton-redis chart: StatefulSet name is {release-name}-proton-redis (e.g., redis-proton-redis)
    local redis_sts_name=""
    if kubectl -n "${redis_ns}" get statefulset redis-proton-redis >/dev/null 2>&1; then
        redis_sts_name="redis-proton-redis"
    elif kubectl -n "${redis_ns}" get statefulset proton-redis-proton-redis >/dev/null 2>&1; then
        redis_sts_name="proton-redis-proton-redis"
    elif kubectl -n "${redis_ns}" get statefulset redis >/dev/null 2>&1; then
        redis_sts_name="redis"
    fi

    # Only treat Redis as configured when a release/resource truly exists in cluster.
    local redis_configured=false
    if [[ -n "${redis_sts_name}" ]] || helm status "${redis_release_name}" -n "${redis_ns}" >/dev/null 2>&1; then
        redis_configured=true
    fi
    
    # Default username is "root" for local chart, "default" for Bitnami chart
    local redis_user="root"
    # Try to get username from StatefulSet or Helm values
    if [[ -n "${redis_sts_name}" ]]; then
        # Try to get from StatefulSet env (for local chart)
        local user_from_sts
        user_from_sts="$(kubectl -n "${redis_ns}" get statefulset "${redis_sts_name}" -o jsonpath='{.spec.template.spec.containers[?(@.name=="redis")].env[?(@.name=="ROOT_USER")].value}' 2>/dev/null || echo "")"
        if [[ -n "${user_from_sts}" ]]; then
            redis_user="${user_from_sts}"
        else
            # Check Helm values for local chart
            local helm_values_json
            helm_values_json="$(helm get values "${redis_release_name}" -n "${redis_ns}" -o json 2>/dev/null || true)"
            if [[ -n "${helm_values_json}" ]]; then
                local user_from_helm
                user_from_helm="$(echo "${helm_values_json}" | grep -oE '"redis":\{[^}]*"rootUsername":"[^"]*"' | grep -oE '"rootUsername":"[^"]*"' | cut -d'"' -f4 || echo "")"
                if [[ -n "${user_from_helm}" ]]; then
                    redis_user="${user_from_helm}"
                fi
            fi
        fi
    fi
    
    # Use default value from script defaults (redis-password) if not set
    local redis_password="${REDIS_PASSWORD:-}"
    # Try to get password from secret (check multiple possible secret names)
    # For proton-redis chart: secret name is {release-name}-proton-redis-secret (e.g., redis-proton-redis-secret)
    local redis_secret_names=(
        "${redis_release_name}-proton-redis-secret"  # proton-redis chart naming
        "proton-redis-proton-redis-secret"           # alternative naming
        "${redis_release_name}-secret"                # generic naming
        "redis-auth"                                  # fallback
    )
    local redis_secret_password=""
    for secret_name in "${redis_secret_names[@]}"; do
        redis_secret_password="$(get_secret_b64_key "${redis_ns}" "${secret_name}" password 2>/dev/null || echo "")"
        if [[ -n "${redis_secret_password}" ]]; then
            # Check if password is base64 encoded (Bitnami chart stores base64-encoded password)
            # Try to decode it; if it's already plain text, decoding will fail or produce garbage
            local decoded_password
            decoded_password="$(printf '%s' "${redis_secret_password}" | base64 -d 2>/dev/null || echo "")"
            if [[ -n "${decoded_password}" ]] && [[ "${decoded_password}" != "${redis_secret_password}" ]]; then
                # Successfully decoded, use the decoded version
                redis_password="${decoded_password}"
            else
                # Not base64 or already plain text
                redis_password="${redis_secret_password}"
            fi
            break
        fi
        # Also try nonEncrpt-password key (used by local chart)
        redis_secret_password="$(get_secret_b64_key "${redis_ns}" "${secret_name}" nonEncrpt-password 2>/dev/null || echo "")"
        if [[ -n "${redis_secret_password}" ]]; then
            # Check if password is base64 encoded
            local decoded_password
            decoded_password="$(printf '%s' "${redis_secret_password}" | base64 -d 2>/dev/null || echo "")"
            if [[ -n "${decoded_password}" ]] && [[ "${decoded_password}" != "${redis_secret_password}" ]]; then
                # Successfully decoded, use the decoded version
                redis_password="${decoded_password}"
            else
                # Not base64 or already plain text
                redis_password="${redis_secret_password}"
            fi
            break
        fi
    done
    
    # Detect Redis deployment mode (standalone or sentinel)
    local redis_connect_type="standalone"
    local redis_host="redis.${redis_ns}.svc.cluster.local"
    local redis_sentinel_host=""
    local redis_sentinel_port="26379"
    local redis_master_group_name="${REDIS_MASTER_GROUP_NAME:-mymaster}"
    
    # Check if Redis is deployed in sentinel mode
    # Method 1: Check if sentinel service exists (for local chart)
    # For proton-redis chart: service name is {release-name}-proton-redis-sentinel
    # If release name is "redis", service is "redis-proton-redis-sentinel"
    # If release name is "proton-redis", service is "proton-redis-proton-redis-sentinel"
    local sentinel_svc_names=(
        "${redis_release_name}-proton-redis-sentinel"  # proton-redis chart naming (release=redis)
        "proton-redis-proton-redis-sentinel"           # proton-redis chart naming (release=proton-redis)
        "${redis_release_name}-sentinel"               # generic naming
        "redis-proton-redis-sentinel"                  # fallback
        "redis-sentinel"                                # fallback
    )
    local sentinel_svc_found=false
    for svc_name in "${sentinel_svc_names[@]}"; do
        if kubectl -n "${redis_ns}" get svc "${svc_name}" >/dev/null 2>&1; then
            redis_connect_type="sentinel"
            # Construct FQDN and remove trailing dot if present
            redis_sentinel_host="${svc_name}.${redis_ns}.svc.cluster.local"
            redis_sentinel_host="${redis_sentinel_host%.}"  # Remove trailing dot
            sentinel_svc_found=true
            log_info "Redis sentinel mode detected (via service: ${svc_name})"
            break
        fi
    done
    
    # Method 2: Check StatefulSet for sentinel container (for local chart)
    if [[ "${sentinel_svc_found}" == "false" ]] && [[ -n "${redis_sts_name}" ]]; then
        local sts_containers
        sts_containers="$(kubectl -n "${redis_ns}" get statefulset "${redis_sts_name}" -o jsonpath='{.spec.template.spec.containers[*].name}' 2>/dev/null || echo "")"
        if echo "${sts_containers}" | grep -q sentinel; then
            redis_connect_type="sentinel"
            # Try to find sentinel service name
            for svc_name in "${sentinel_svc_names[@]}"; do
                if kubectl -n "${redis_ns}" get svc "${svc_name}" >/dev/null 2>&1; then
                    redis_sentinel_host="${svc_name}.${redis_ns}.svc.cluster.local"
                    redis_sentinel_host="${redis_sentinel_host%.}"  # Remove trailing dot
                    break
                fi
            done
            # If no service found, use default naming for proton-redis chart
            # For proton-redis chart: {release-name}-proton-redis-sentinel
            if [[ -z "${redis_sentinel_host}" ]]; then
                # Calculate StatefulSet name: {release-name}-proton-redis
                local calculated_sts_name="${redis_release_name}-proton-redis"
                redis_sentinel_host="${calculated_sts_name}-sentinel.${redis_ns}.svc.cluster.local"
                redis_sentinel_host="${redis_sentinel_host%.}"  # Remove trailing dot
            fi
            log_info "Redis sentinel mode detected (via StatefulSet containers)"
        fi
    fi
    
    # Do NOT infer sentinel mode from REDIS_ARCHITECTURE default variable here,
    # otherwise a fresh cluster without Redis would be mis-detected as sentinel.
    
    # For sentinel mode, try to get master group name from StatefulSet or use default
    if [[ "${redis_connect_type}" == "sentinel" ]] && [[ -n "${redis_sts_name}" ]]; then
        # Try to get from StatefulSet env or config
        local master_group_from_sts
        master_group_from_sts="$(kubectl -n "${redis_ns}" get statefulset "${redis_sts_name}" -o jsonpath='{.spec.template.spec.containers[?(@.name=="sentinel")].env[?(@.name=="MASTER_GROUP")].value}' 2>/dev/null || echo "")"
        if [[ -n "${master_group_from_sts}" ]]; then
            redis_master_group_name="${master_group_from_sts}"
        else
            # Try to get from Helm values
            local helm_values_json
            helm_values_json="$(helm get values "${redis_release_name}" -n "${redis_ns}" -o json 2>/dev/null || true)"
            if [[ -n "${helm_values_json}" ]]; then
                local master_group_from_helm
                master_group_from_helm="$(echo "${helm_values_json}" | grep -oE '"redis":\{[^}]*"masterGroupName":"[^"]*"' | grep -oE '"masterGroupName":"[^"]*"' | cut -d'"' -f4 || true)"
                if [[ -n "${master_group_from_helm}" ]]; then
                    redis_master_group_name="${master_group_from_helm}"
                fi
            fi
        fi
    fi

    # OpenSearch
    local os_ns="${OPENSEARCH_NAMESPACE}"
    local os_host="${OPENSEARCH_CLUSTER_NAME}-${OPENSEARCH_NODE_GROUP}.${os_ns}.svc.cluster.local"
    local os_user="admin"
    local os_password="${OPENSEARCH_INITIAL_ADMIN_PASSWORD}"
    local os_protocol="${OPENSEARCH_PROTOCOL}"
    local opensearch_configured=false
    if [[ -z "${os_protocol}" ]]; then
        os_protocol="http"
    fi
    if helm status "${OPENSEARCH_RELEASE_NAME}" -n "${os_ns}" >/dev/null 2>&1 || \
       kubectl -n "${os_ns}" get svc "${OPENSEARCH_CLUSTER_NAME}-${OPENSEARCH_NODE_GROUP}" >/dev/null 2>&1; then
        opensearch_configured=true
    fi

    # MongoDB - only generate config if MongoDB is installed
    local mongodb_ns="${MONGODB_NAMESPACE}"
    # Prefer a stable service name "mongodb" (release name) for clients: mongodb.<ns>.svc.cluster.local
    local mongodb_host="${MONGODB_RELEASE_NAME}.${mongodb_ns}.svc.cluster.local"
    local mongodb_port="28000"
    local mongodb_user="${MONGODB_SECRET_USERNAME}"
    local mongodb_password="${MONGODB_SECRET_PASSWORD}"
    local mongodb_configured=false
    
    # Check if MongoDB secret exists (indicates MongoDB is installed)
    if kubectl -n "${mongodb_ns}" get secret "${MONGODB_SECRET_NAME}" >/dev/null 2>&1; then
        mongodb_configured=true
        if [[ -z "${mongodb_password}" ]]; then
            # Try to get password from secret
            mongodb_password=$(kubectl -n "${mongodb_ns}" get secret "${MONGODB_SECRET_NAME}" -o jsonpath='{.data.password}' 2>/dev/null | base64 -d 2>/dev/null || echo "")
        fi
    fi
    
    # MongoDB connection parameters (config.yaml schema expected by proton-cli)
    # Set replicaSet based on whether replica set is enabled
    local mongodb_replica_set=""
    if [[ "${mongodb_configured}" == "true" ]]; then
        # Check if replica set is enabled by checking StatefulSet args or using default config
        # First, try to detect from StatefulSet
        local sts_replset
        sts_replset=$(kubectl -n "${mongodb_ns}" get statefulset "${MONGODB_RELEASE_NAME}-mongodb" -o jsonpath='{.spec.template.spec.containers[0].args[*]}' 2>/dev/null | grep -o "replSet [^ ]*" | awk '{print $2}' || echo "")
        if [[ -n "${sts_replset}" ]]; then
            mongodb_replica_set="${sts_replset}"
        elif [[ "${MONGODB_REPLSET_ENABLED:-true}" == "true" ]]; then
            # Use default from script variable (default is true for single-node replica set)
            mongodb_replica_set="${MONGODB_REPLSET_NAME:-rs0}"
        fi
        # If keyfile exists in secret, it's likely replica set mode
        local has_keyfile
        has_keyfile=$(kubectl -n "${mongodb_ns}" get secret "${MONGODB_SECRET_NAME}" -o jsonpath='{.data.mongodb\.keyfile}' 2>/dev/null || echo "")
        if [[ -n "${has_keyfile}" ]] && [[ -z "${mongodb_replica_set}" ]]; then
            mongodb_replica_set="${MONGODB_REPLSET_NAME:-rs0}"
        fi
    fi
    # Always use admin as authSource
    local mongodb_auth_source="admin"

    # Kafka
    local kafka_ns="${KAFKA_NAMESPACE}"
    local kafka_mechanism="${KAFKA_SASL_MECHANISM}"
    local kafka_user="${KAFKA_CLIENT_USER}"
    local kafka_password="${KAFKA_CLIENT_PASSWORD}"
    local kafka_configured=false
    if [[ "${KAFKA_AUTH_ENABLED}" == "true" ]]; then
        local client_pw
        client_pw="$(get_secret_b64_key "${kafka_ns}" "${KAFKA_SASL_SECRET_NAME}" client-passwords)"
        if [[ -n "${client_pw}" ]]; then
            kafka_password="${client_pw%%,*}"
        fi
    fi
    local kafka_svc
    kafka_svc="$(first_service_with_port "${kafka_ns}" "app.kubernetes.io/instance=${KAFKA_RELEASE_NAME}" 9092)"
    if [[ -n "${kafka_svc}" ]]; then
        kafka_configured=true
    elif kubectl -n "${kafka_ns}" get svc "${KAFKA_RELEASE_NAME}" >/dev/null 2>&1; then
        kafka_svc="${KAFKA_RELEASE_NAME}"
        kafka_configured=true
    fi
    local kafka_host=""
    if [[ "${kafka_configured}" == "true" ]]; then
        kafka_host="${kafka_svc}.${kafka_ns}.svc.cluster.local"
    fi

    # Zookeeper - only generate config if Zookeeper is installed
    local zookeeper_ns="${ZOOKEEPER_NAMESPACE}"
    # Zookeeper uses headless service: {release-name}-headless.{namespace}.svc.cluster.local
    # Default release name is "zookeeper", so service name is "zookeeper-headless"
    # Use full FQDN for reliability across namespaces and clusters
    local zookeeper_host="${ZOOKEEPER_RELEASE_NAME}-headless.${zookeeper_ns}.svc.cluster.local"
    local zookeeper_port="${ZOOKEEPER_SERVICE_PORT:-2181}"
    local zookeeper_configured=false
    
    # Check if Zookeeper StatefulSet or Service exists (indicates Zookeeper is installed)
    # Try multiple detection methods for robustness
    local zookeeper_detected=false
    if kubectl -n "${zookeeper_ns}" get statefulset "${ZOOKEEPER_RELEASE_NAME}" >/dev/null 2>&1; then
        zookeeper_detected=true
    elif kubectl -n "${zookeeper_ns}" get svc "${ZOOKEEPER_RELEASE_NAME}-headless" >/dev/null 2>&1; then
        zookeeper_detected=true
    elif kubectl -n "${zookeeper_ns}" get statefulset -l "app=${ZOOKEEPER_RELEASE_NAME}" >/dev/null 2>&1; then
        zookeeper_detected=true
    elif kubectl -n "${zookeeper_ns}" get statefulset -l "app=zookeeper" >/dev/null 2>&1; then
        zookeeper_detected=true
    fi
    
    if [[ "${zookeeper_detected}" == "true" ]]; then
        zookeeper_configured=true
    fi

    # Build MongoDB config section if MongoDB is installed
    local mongodb_section=""
    if [[ "${mongodb_configured}" == "true" ]]; then
        mongodb_section=$(cat <<MONGODB_EOF
  mongodb:
    source_type: internal
    host: $(yaml_quote "${mongodb_host}")
    port: ${mongodb_port}
    user: $(yaml_quote "${mongodb_user}")
    password: $(yaml_quote "${mongodb_password}")
    replicaSet: $(yaml_quote "${mongodb_replica_set}")
    options:
      authSource: $(yaml_quote "${mongodb_auth_source}")
MONGODB_EOF
)
    fi

    # Build Zookeeper config section if Zookeeper is installed
    local zookeeper_section=""
    if [[ "${zookeeper_configured}" == "true" ]]; then
        # Use full FQDN for reliability across namespaces and clusters
        zookeeper_section=$(cat <<ZOOKEEPER_EOF
  zookeeper:
    host: $(yaml_quote "${zookeeper_host}")
    port: ${zookeeper_port}
ZOOKEEPER_EOF
)
    fi

    # Ingress-Nginx - detect actual IngressClass name
    local ingress_class_name=""
    local ingress_class_configured=false
    local ingress_nginx_release="ingress-nginx"
    local ingress_nginx_namespace="ingress-nginx"
    
    # First, try to get IngressClass from actual Kubernetes resource (most reliable)
    ingress_class_name="$(kubectl get ingressclass -o jsonpath='{.items[?(@.spec.controller=="k8s.io/ingress-nginx")].metadata.name}' 2>/dev/null | awk '{print $1}' || true)"
    
    # If not found, try to get from Helm release values
    if [[ -z "${ingress_class_name}" ]] && helm status "${ingress_nginx_release}" -n "${ingress_nginx_namespace}" >/dev/null 2>&1; then
        # Try to get from Helm values
        local helm_values_json
        helm_values_json="$(helm get values "${ingress_nginx_release}" -n "${ingress_nginx_namespace}" -o json 2>/dev/null || true)"
        if [[ -n "${helm_values_json}" ]]; then
            # Extract controller.ingressClassResource.name or controller.ingressClass
            ingress_class_name="$(echo "${helm_values_json}" | grep -oE '"controller\.(ingressClassResource\.name|ingressClass)":"[^"]*"' | head -1 | cut -d'"' -f4 || true)"
        fi
    fi
    
    # If still not found, try to get from deployment args
    if [[ -z "${ingress_class_name}" ]]; then
        local deploy_args
        deploy_args="$(kubectl -n "${ingress_nginx_namespace}" get deploy ingress-nginx-controller -o jsonpath='{.spec.template.spec.containers[0].args[*]}' 2>/dev/null || true)"
        if [[ -n "${deploy_args}" ]]; then
            # Extract --ingress-class value
            ingress_class_name="$(echo "${deploy_args}" | grep -oE '--ingress-class[= ]([^ ]+)' | awk '{print $2}' | tr -d '=' || true)"
        fi
    fi
    
    # Final fallback: use default from script variable
    if [[ -z "${ingress_class_name}" ]]; then
        ingress_class_name="${INGRESS_NGINX_CLASS:-class-443}"
    fi
    
    # Check if ingress-nginx is actually installed (Helm release or deployment exists)
    if helm status "${ingress_nginx_release}" -n "${ingress_nginx_namespace}" >/dev/null 2>&1 || \
       kubectl -n "${ingress_nginx_namespace}" get deploy ingress-nginx-controller >/dev/null 2>&1; then
        ingress_class_configured=true
    fi

    # Build ingress class config section (always use "class-443" as key, but with actual ingressClass value)
    local ingress_class_section=""
    if [[ "${ingress_class_configured}" == "true" ]]; then
        ingress_class_section=$(cat <<INGRESS_CLASS_EOF
  class-443:
    ingressClass: $(yaml_quote "${ingress_class_name}")
INGRESS_CLASS_EOF
)
    fi

    # Build Redis config section based on deployment mode
    local redis_section=""
    if [[ "${redis_configured}" == "true" ]]; then
        if [[ "${redis_connect_type}" == "sentinel" ]]; then
            redis_section=$(cat <<REDIS_SENTINEL_EOF
  redis:
    connectInfo:
      masterGroupName: $(yaml_quote "${redis_master_group_name}")
      password: $(yaml_quote "${redis_password}")
      sentinelHost: $(yaml_quote "${redis_sentinel_host}")
      sentinelPassword: $(yaml_quote "${redis_password}")
      sentinelPort: ${redis_sentinel_port}
      sentinelUsername: $(yaml_quote "${redis_user}")
      username: $(yaml_quote "${redis_user}")
    connectType: $(yaml_quote "${redis_connect_type}")
    sourceType: internal
REDIS_SENTINEL_EOF
)
        else
            redis_section=$(cat <<REDIS_STANDALONE_EOF
  redis:
    connectInfo:
      host: $(yaml_quote "${redis_host}")
      port: 6379
      username: $(yaml_quote "${redis_user}")
      password: $(yaml_quote "${redis_password}")
    connectType: $(yaml_quote "${redis_connect_type}")
    sourceType: internal
REDIS_STANDALONE_EOF
)
        fi
    fi

    local mq_section=""
    if [[ "${kafka_configured}" == "true" ]]; then
        mq_section=$(cat <<MQ_EOF
  mq:
    auth:
      mechanism: $(yaml_quote "${kafka_mechanism}")
      username: $(yaml_quote "${kafka_user}")
      password: $(yaml_quote "${kafka_password}")
    mqHost: $(yaml_quote "${kafka_host}")
    mqLookupdHost: ""
    mqLookupdPort: 0
    mqPort: 9092
    mqType: kafka
MQ_EOF
)
    fi

    local opensearch_section=""
    if [[ "${opensearch_configured}" == "true" ]]; then
        opensearch_section=$(cat <<OS_EOF
  opensearch:
    distribution: opensearch
    host: $(yaml_quote "${os_host}")
    user: $(yaml_quote "${os_user}")
    password: $(yaml_quote "${os_password}")
    port: 9200
    protocol: ${os_protocol}
    version: ""
OS_EOF
)
    fi

    local rds_section=""
    if [[ "${mariadb_configured}" == "true" ]]; then
        rds_section=$(cat <<RDS_EOF
  rds:
    admin_key: $(yaml_quote "${mariadb_admin_key}")
    host: $(yaml_quote "${mariadb_host}")
    hostRead: $(yaml_quote "${mariadb_host}")
    port: 3306
    portRead: 3306
    source_type: internal
    type: MariaDB
    user: $(yaml_quote "${mariadb_user}")
    password: $(yaml_quote "${mariadb_password}")
    root_password: $(yaml_quote "${mariadb_root_password}")
    database: $(yaml_quote "${mariadb_database}")
RDS_EOF
)
    fi

    local dep_services_section=""
    if [[ -n "${mq_section}${opensearch_section}${mongodb_section}${zookeeper_section}${rds_section}${redis_section}" ]]; then
        dep_services_section=$(cat <<DEP_EOF
depServices:
${mq_section}
${opensearch_section}
${mongodb_section}
${zookeeper_section}
${rds_section}
${redis_section}
${ingress_class_section}
DEP_EOF
)
    fi

    cat > "${out}" <<EOF
namespace: ${cfg_namespace}
env:
  language: ${cfg_lang}
  timezone: ${cfg_tz}
image:
  registry: ${IMAGE_REGISTRY}
${storage_section}
accessAddress:
  host: ${node_ip}
  port: 443
  scheme: https
  path: /
${dep_services_section}
EOF

    log_info "Wrote config file: ${out}"
    local included_services=()
    [[ "${mongodb_configured}" == "true" ]] && included_services+=("MongoDB")
    [[ "${zookeeper_configured}" == "true" ]] && included_services+=("Zookeeper")
    [[ "${ingress_class_configured}" == "true" ]] && included_services+=("Ingress-Nginx")
    [[ "${redis_configured}" == "true" ]] && included_services+=("Redis")
    [[ "${kafka_configured}" == "true" ]] && included_services+=("Kafka")
    [[ "${opensearch_configured}" == "true" ]] && included_services+=("OpenSearch")
    [[ "${mariadb_configured}" == "true" ]] && included_services+=("MariaDB")
    if [[ ${#included_services[@]} -gt 0 ]]; then
        log_info "Included services in config.yaml: ${included_services[*]}"
    else
        log_info "No dependency services detected; depServices section not written"
    fi
}
