
is_bitnami_mariadb_image() {
    local image="$1"
    [[ "${image}" == *"/bitnami/mariadb:"* || "${image}" == "bitnami/mariadb:"* ]]
}

# Update RDS type to internal in config.yaml after MariaDB installation
update_rds_type_to_internal() {
    local config_file="${CONFIG_YAML_PATH}"
    
    if [[ ! -f "${config_file}" ]]; then
        log_warn "Config file not found: ${config_file}"
        return 1
    fi
    
    log_info "Updating RDS type to 'internal' in config.yaml..."
    
    # Update rds section's source_type from external to internal
    # Use a more flexible sed pattern that doesn't require end-of-line anchor
    sed -i '/^  rds:/,/^  [a-z]/s/source_type: external/source_type: internal/' "${config_file}"
    
    # Verify the change
    if grep -A 10 "^  rds:" "${config_file}" | grep -q "source_type: internal"; then
        log_info "✓ RDS type successfully updated to 'internal'"
        return 0
    else
        log_warn "⚠ Failed to update RDS type to 'internal', trying alternative method..."
        # Fallback: use a simpler global replacement
        sed -i 's/source_type: external/source_type: internal/g' "${config_file}"
        if grep -q "source_type: internal" "${config_file}"; then
            log_info "✓ RDS type successfully updated to 'internal' (using fallback)"
            return 0
        else
            log_error "✗ Failed to update RDS type to 'internal'"
            return 1
        fi
    fi
}

# When values file used loopback RDS hosts (e.g. mac/kind samples), point depServices.rds at
# the in-cluster MariaDB Service so Helm hooks and pods resolve MySQL correctly.
sync_config_rds_for_in_cluster_mariadb() {
    local config_file="${CONFIG_YAML_PATH:-}"
    local ns="${MARIADB_NAMESPACE:-resource}"
    local svc_host="mariadb.${ns}.svc.cluster.local"

    [[ -f "${config_file}" ]] || return 0
    if ! grep -A40 '^  rds:' "${config_file}" | grep -E '^[[:space:]]*host(Read)?:' | grep -qE "127\.0\.0\.1|'127\.0\.0\.1'|localhost|'localhost'"; then
        return 0
    fi

    log_info "Updating ${config_file}: depServices.rds -> internal + ${svc_host} (replacing loopback placeholders)"
    update_rds_type_to_internal || true

    local tmp
    tmp="$(mktemp)"
    awk -v h="${svc_host}" '
        /^  rds:/ { inr=1 }
        inr && /^  [a-z]/ && $0 !~ /^  rds:/ { inr=0 }
        inr && /^    host:/ { print "    host: '\''" h "'\''"; next }
        inr && /^    hostRead:/ { print "    hostRead: '\''" h "'\''"; next }
        { print }
    ' "${config_file}" > "${tmp}"
    mv "${tmp}" "${config_file}"
}

# Grant permissions to MariaDB user (always run for internal MariaDB)
grant_mariadb_user_permissions() {
    local ns="${MARIADB_NAMESPACE}"

    log_info "Granting permissions to MariaDB user '${MARIADB_USER}'..."

    # Find the correct pod name
    local pod_name
    pod_name=$(kubectl -n "${ns}" get pods -l app=mariadb -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

    if [[ -z "${pod_name}" ]]; then
        log_warn "MariaDB Pod not found, skipping permission grant"
        return 1
    fi

    # Wait for MariaDB to be fully ready
    log_info "Waiting for MariaDB to be ready (Pod: ${pod_name})..."
    local max_attempts=15
    local attempt=0
    while [[ $attempt -lt $max_attempts ]]; do
        if kubectl -n "${ns}" exec "${pod_name}" -- mariadb-admin ping -h localhost -u root -p"${MARIADB_ROOT_PASSWORD}" &>/dev/null; then
            break
        fi
        ((attempt++))
        sleep 2
    done

    # Grant all privileges to the user
    local sql_commands="GRANT ALL PRIVILEGES ON *.* TO '${MARIADB_USER}'@'%' WITH GRANT OPTION;"
    sql_commands+=" FLUSH PRIVILEGES;"

    if echo "${sql_commands}" | kubectl -n "${ns}" exec -i "${pod_name}" -- mariadb -u root -p"${MARIADB_ROOT_PASSWORD}" 2>&1; then
        log_info "✓ Successfully granted all privileges to '${MARIADB_USER}'@'%'"
    else
        log_error "✗ Failed to grant privileges to '${MARIADB_USER}'@'%'"
        return 1
    fi
}

# Create additional databases and grant permissions to adp user
setup_mariadb_databases() {
    local ns="${MARIADB_NAMESPACE}"
    local mariadb_host="mariadb.${ns}.svc.cluster.local"
    local mariadb_port="3306"

    # Always grant user permissions first
    grant_mariadb_user_permissions

    # Check if ISF or Core manifest has pre-stage data-migrator (0.6.0+)
    # If so, skip manual database creation - the data-migrator chart will handle it
    local isf_manifest core_manifest
    isf_manifest="${ISF_VERSION_MANIFEST_FILE:-$(resolve_embedded_release_manifest "isf" "${HELM_CHART_VERSION:-}")}"
    core_manifest="${CORE_VERSION_MANIFEST_FILE:-$(resolve_embedded_release_manifest "kweaver-core" "${HELM_CHART_VERSION:-}")}"

    if [[ -f "${isf_manifest}" ]] && should_skip_db_init_for_manifest "${isf_manifest}"; then
        log_info "ISF manifest ${isf_manifest} has pre-stage data-migrator (0.6.0+), skipping manual database creation"
        return 0
    fi

    if [[ -f "${core_manifest}" ]] && should_skip_db_init_for_manifest "${core_manifest}"; then
        log_info "Core manifest ${core_manifest} has pre-stage data-migrator (0.6.0+), skipping manual database creation"
        return 0
    fi

    log_info "Setting up additional databases and permissions..."

    # Find the correct pod name (could be mariadb-0 or mariadb-proton-mariadb-0 depending on chart)
    local pod_name
    pod_name=$(kubectl -n "${ns}" get pods -l app=mariadb -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

    if [[ -z "${pod_name}" ]]; then
        log_warn "MariaDB Pod not found, skipping database setup"
        return 1
    fi

    # Wait for MariaDB to be fully ready
    log_info "Waiting for MariaDB to be ready (Pod: ${pod_name})..."
    local max_attempts=15
    local attempt=0
    while [[ $attempt -lt $max_attempts ]]; do
        if kubectl -n "${ns}" exec "${pod_name}" -- mariadb-admin ping -h localhost -u root -p"${MARIADB_ROOT_PASSWORD}" &>/dev/null; then
            break
        fi
        ((attempt++))
        sleep 2
    done

    # List of databases to create
    local databases=(
        "user_management" "anyshare" "policy_mgnt" "privacy" "authentication"
        "eofs" "deploy" "sharemgnt_db" "ets" "ossmanager" "license"
        "nodemgnt" "sites" "anydata" "third_app_mgnt" "hydra_v2" "thirdparty_message"
    )

    # Execute SQL commands to create databases
    log_info "Creating databases..."
    local sql_commands="CREATE DATABASE IF NOT EXISTS \`${MARIADB_DATABASE}\`;"
    for db in "${databases[@]}"; do
        sql_commands+=" CREATE DATABASE IF NOT EXISTS \`${db}\`;"
    done

    if echo "${sql_commands}" | kubectl -n "${ns}" exec -i "${pod_name}" -- mariadb -u root -p"${MARIADB_ROOT_PASSWORD}" 2>&1; then
        log_info "✓ Successfully created databases"
    else
        log_warn "⚠ Some databases may already exist or failed to create"
    fi

    log_info "MariaDB database setup completed"
}

# Install single-node MariaDB 11 using proton-mariadb Helm chart
install_mariadb_helm() {
    log_info "Installing MariaDB (single-node) via proton-mariadb Helm chart..."

    local ns="${MARIADB_NAMESPACE}"
    local fresh_install="true"

    # Check if MariaDB is already installed
    if helm status mariadb -n "${ns}" >/dev/null 2>&1; then
        fresh_install="false"
        log_info "MariaDB is already installed (Helm release exists). Skipping installation."
        # Even if MariaDB is already installed, regenerate config.yaml if password is missing
        local existing_pass
        existing_pass=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" 2>/dev/null | grep "password:" | head -1 | awk '{print $2}' | tr -d "'\"")
        if [[ -z "${existing_pass}" ]] && [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
            log_info "MariaDB password is missing in config.yaml, regenerating..."
            generate_config_yaml
            update_rds_type_to_internal
        fi
        return 0
    fi

    # MariaDB password handling
    local existing_pass=$(get_existing_password "mariadb.password")
    if [[ -n "${existing_pass}" ]]; then
        MARIADB_PASSWORD="${existing_pass}"
        log_info "Using existing MariaDB password from config.yaml"
    else
        MARIADB_PASSWORD=$(generate_random_password 10)
        log_info "Generated random 10-character MariaDB password"
    fi

    local existing_root_pass=$(get_existing_password "mariadb.root_password")
    if [[ -n "${existing_root_pass}" ]]; then
        MARIADB_ROOT_PASSWORD="${existing_root_pass}"
        log_info "Using existing MariaDB root password from config.yaml"
    else
        MARIADB_ROOT_PASSWORD=$(generate_random_password 10)
        log_info "Generated random 10-character MariaDB root password"
    fi

    if [[ -z "${MARIADB_IMAGE}" ]]; then
        MARIADB_IMAGE="$(image_from_registry "${MARIADB_IMAGE_REPOSITORY}" "${MARIADB_IMAGE_TAG}" "${MARIADB_IMAGE_FALLBACK}")"
    fi

    # Parse image registry/repository/tag from MARIADB_IMAGE
    local image_without_tag="${MARIADB_IMAGE%:*}"
    local image_tag="${MARIADB_IMAGE##*:}"

    local chart_ref="proton-mariadb"
    local use_local_chart="false"
    if [[ -f "${MARIADB_CHART_TGZ}" ]]; then
        chart_ref="${MARIADB_CHART_TGZ}"
        use_local_chart="true"
        log_info "Using local MariaDB chart: ${chart_ref}"
    else
        log_info "Using remote MariaDB chart: ${chart_ref} (version ${MARIADB_CHART_VERSION})"
    fi

    kubectl create namespace "${ns}" 2>/dev/null || true

    local persistence_enabled="${MARIADB_PERSISTENCE_ENABLED}"
    if [[ "${persistence_enabled}" == "true" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            log_warn "No StorageClass found. MariaDB PVC will stay Pending."
            if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                install_localpv
            fi
            if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                log_warn "Still no StorageClass found. Disabling MariaDB persistence to avoid Pending PVC."
                persistence_enabled="false"
            fi
        fi
    fi

    local -a helm_args
    helm_args=(
        upgrade --install mariadb "${chart_ref}"
        --namespace "${ns}"
        --set image.repository="${image_without_tag}"
        --set image.tag="${image_tag}"
        --set mariadb.auth.rootPassword="${MARIADB_ROOT_PASSWORD}"
        --set mariadb.auth.database="${MARIADB_DATABASE}"
        --set mariadb.auth.username="${MARIADB_USER}"
        --set mariadb.auth.password="${MARIADB_PASSWORD}"
        --set mariadb.config.maxConnections="${MARIADB_MAX_CONNECTIONS}"
        --set mariadb.persistence.enabled="${persistence_enabled}"
    )

    if [[ "${persistence_enabled}" == "true" ]]; then
        helm_args+=(--set mariadb.persistence.size="${MARIADB_STORAGE_SIZE}")
        if [[ -n "${MARIADB_STORAGE_CLASS}" ]]; then
            helm_args+=(--set mariadb.persistence.storageClassName="${MARIADB_STORAGE_CLASS}")
        fi
    fi

    helm_args+=(--wait --timeout=600s)

    log_info "Installing MariaDB with values:"
    log_info "  Chart: ${chart_ref}"
    log_info "  Namespace: ${ns}"
    log_info "  Image: ${image_without_tag}:${image_tag}"
    log_info "  Max Connections: ${MARIADB_MAX_CONNECTIONS}"
    log_info "  Storage: ${persistence_enabled}"

    helm "${helm_args[@]}"

    # Wait for MariaDB Pod to be ready
    log_info "Waiting for MariaDB Pod to be ready..."
    kubectl wait --for=condition=ready pod -l "app.kubernetes.io/instance=mariadb" -n "${ns}" --timeout=300s 2>/dev/null || {
        log_warn "MariaDB Pod may not be ready yet"
    }

    log_info "MariaDB installed successfully"
    log_info "MariaDB connection info:"
    log_info "  Host: mariadb.${ns}.svc.cluster.local"
    log_info "  Port: 3306"
    log_info "  Root Password: ${MARIADB_ROOT_PASSWORD}"
    log_info "  Database: ${MARIADB_DATABASE}"
    log_info "  Username: ${MARIADB_USER}"
    log_info "  Password: ${MARIADB_PASSWORD}"

    # Create additional databases and grant permissions
    setup_mariadb_databases

    sync_config_rds_for_in_cluster_mariadb

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Updating conf/config.yaml after MariaDB fresh install..."
        generate_config_yaml
        # Update RDS type to internal since MariaDB is installed internally
        update_rds_type_to_internal
    fi
}

# Install single-node MariaDB 11 via manifest (official mariadb image)
install_mariadb_official() {
    log_info "Installing MariaDB (single-node) via manifest..."

    local ns="${MARIADB_NAMESPACE}"

    # MariaDB password handling
    local existing_pass=$(get_existing_password "mariadb.password")
    if [[ -n "${existing_pass}" ]]; then
        MARIADB_PASSWORD="${existing_pass}"
        log_info "Using existing MariaDB password from config.yaml"
    else
        MARIADB_PASSWORD=$(generate_random_password 10)
        log_info "Generated random 10-character MariaDB password"
    fi

    local existing_root_pass=$(get_existing_password "mariadb.root_password")
    if [[ -n "${existing_root_pass}" ]]; then
        MARIADB_ROOT_PASSWORD="${existing_root_pass}"
        log_info "Using existing MariaDB root password from config.yaml"
    else
        MARIADB_ROOT_PASSWORD=$(generate_random_password 10)
        log_info "Generated random 10-character MariaDB root password"
    fi

    if kubectl -n "${ns}" get statefulset mariadb >/dev/null 2>&1; then
        log_info "MariaDB is already installed (StatefulSet exists). Skipping installation."
        # Even if MariaDB is already installed, regenerate config.yaml if password is missing
        local existing_pass
        existing_pass=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" 2>/dev/null | grep "password:" | head -1 | awk '{print $2}' | tr -d "'\"")
        if [[ -z "${existing_pass}" ]] && [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
            log_info "MariaDB password is missing in config.yaml, regenerating..."
            generate_config_yaml
            update_rds_type_to_internal
        fi
        return 0
    fi

    if [[ -z "${MARIADB_IMAGE}" ]]; then
        MARIADB_IMAGE="$(image_from_registry "${MARIADB_IMAGE_REPOSITORY}" "${MARIADB_IMAGE_TAG}" "${MARIADB_IMAGE_FALLBACK}")"
    fi

    # Create namespace if not exists
    kubectl create namespace "${ns}" 2>/dev/null || true

    local persistence_enabled="${MARIADB_PERSISTENCE_ENABLED}"
    if [[ "${persistence_enabled}" == "true" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            log_warn "No StorageClass found. MariaDB PVC will stay Pending."
            if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                install_localpv
            fi

            if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                log_warn "Still no StorageClass found. Disabling MariaDB persistence to avoid Pending PVC."
                persistence_enabled="false"
            fi
        fi
    fi

    # Best-effort cleanup if a previous Bitnami release exists.
    helm uninstall mariadb -n "${ns}" 2>/dev/null || true
    kubectl -n "${ns}" delete statefulset mariadb 2>/dev/null || true
    kubectl -n "${ns}" delete svc mariadb mariadb-headless 2>/dev/null || true
    kubectl -n "${ns}" delete secret mariadb-auth 2>/dev/null || true

    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: mariadb-auth
  namespace: ${ns}
type: Opaque
stringData:
  mariadb-root-password: "${MARIADB_ROOT_PASSWORD}"
  mariadb-database: "${MARIADB_DATABASE}"
  mariadb-user: "${MARIADB_USER}"
  mariadb-password: "${MARIADB_PASSWORD}"
---
apiVersion: v1
kind: Service
metadata:
  name: mariadb
  namespace: ${ns}
  labels:
    app: mariadb
spec:
  type: ClusterIP
  selector:
    app: mariadb
  ports:
    - name: mysql
      port: 3306
      targetPort: 3306
---
apiVersion: v1
kind: Service
metadata:
  name: mariadb-headless
  namespace: ${ns}
  labels:
    app: mariadb
spec:
  clusterIP: None
  selector:
    app: mariadb
  ports:
    - name: mysql
      port: 3306
      targetPort: 3306
EOF

    if [[ "${persistence_enabled}" == "true" ]]; then
        local storage_class_yaml=""
        if [[ -n "${MARIADB_STORAGE_CLASS}" ]]; then
            storage_class_yaml="        storageClassName: \"${MARIADB_STORAGE_CLASS}\""
        fi

        cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mariadb
  namespace: ${ns}
spec:
  serviceName: mariadb-headless
  replicas: 1
  selector:
    matchLabels:
      app: mariadb
  template:
    metadata:
      labels:
        app: mariadb
    spec:
      containers:
        - name: mariadb
          image: ${MARIADB_IMAGE}
          imagePullPolicy: IfNotPresent
          ports:
            - name: mysql
              containerPort: 3306
          args:
            - --max-connections=${MARIADB_MAX_CONNECTIONS}
          env:
            - name: MARIADB_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-root-password
            - name: MARIADB_DATABASE
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-database
            - name: MARIADB_USER
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-user
            - name: MARIADB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-password
          volumeMounts:
            - name: data
              mountPath: /var/lib/mysql
          readinessProbe:
            tcpSocket:
              port: 3306
            initialDelaySeconds: 10
            periodSeconds: 5
          livenessProbe:
            tcpSocket:
              port: 3306
            initialDelaySeconds: 30
            periodSeconds: 10
          resources:
            requests:
              cpu: 250m
              memory: 256Mi
            limits:
              cpu: 375m
              memory: 384Mi
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: ${MARIADB_STORAGE_SIZE}
${storage_class_yaml}
EOF
    else
        cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mariadb
  namespace: ${ns}
spec:
  serviceName: mariadb-headless
  replicas: 1
  selector:
    matchLabels:
      app: mariadb
  template:
    metadata:
      labels:
        app: mariadb
    spec:
      containers:
        - name: mariadb
          image: ${MARIADB_IMAGE}
          imagePullPolicy: IfNotPresent
          ports:
            - name: mysql
              containerPort: 3306
          args:
            - --max-connections=${MARIADB_MAX_CONNECTIONS}
          env:
            - name: MARIADB_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-root-password
            - name: MARIADB_DATABASE
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-database
            - name: MARIADB_USER
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-user
            - name: MARIADB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: mariadb-auth
                  key: mariadb-password
          volumeMounts:
            - name: data
              mountPath: /var/lib/mysql
          readinessProbe:
            tcpSocket:
              port: 3306
            initialDelaySeconds: 10
            periodSeconds: 5
          livenessProbe:
            tcpSocket:
              port: 3306
            initialDelaySeconds: 30
            periodSeconds: 10
          resources:
            requests:
              cpu: 250m
              memory: 256Mi
            limits:
              cpu: 375m
              memory: 384Mi
      volumes:
        - name: data
          emptyDir: {}
EOF
    fi

    kubectl -n "${ns}" rollout status statefulset/mariadb --timeout=600s 2>/dev/null || true
    kubectl -n "${ns}" wait --for=condition=Ready pod/mariadb-0 --timeout=600s 2>/dev/null || true
    
    log_info "MariaDB installed successfully"
    log_info "MariaDB connection info:"
    log_info "  Host: mariadb.${ns}.svc.cluster.local"
    log_info "  Port: 3306"
    log_info "  Root Password: ${MARIADB_ROOT_PASSWORD}"
    log_info "  Database: ${MARIADB_DATABASE}"
    log_info "  Username: ${MARIADB_USER}"
    log_info "  Password: ${MARIADB_PASSWORD}"

    # Create additional databases and grant permissions
    setup_mariadb_databases

    if [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Updating conf/config.yaml after MariaDB fresh install..."
        generate_config_yaml
        # Update RDS type to internal since MariaDB is installed internally
        update_rds_type_to_internal
    fi
}

# Install single-node MariaDB 11 using Bitnami chart (requires Bitnami image layout)
install_mariadb_bitnami() {
    log_info "Installing MariaDB (single-node) via Bitnami Helm chart..."

    local ns="${MARIADB_NAMESPACE}"
    local fresh_install="true"

    # Check if MariaDB is already installed
    if is_helm_installed "mariadb" "${ns}"; then
        fresh_install="false"
        log_info "MariaDB is already installed (Helm release exists). Skipping installation."
        # Even if MariaDB is already installed, regenerate config.yaml if password is missing
        local existing_pass
        existing_pass=$(grep -A 20 "^  rds:" "${CONFIG_YAML_PATH}" 2>/dev/null | grep "password:" | head -1 | awk '{print $2}' | tr -d "'\"")
        if [[ -z "${existing_pass}" ]] && [[ "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
            log_info "MariaDB password is missing in config.yaml, regenerating..."
            generate_config_yaml
            update_rds_type_to_internal
        fi
        return 0
    fi

    # MariaDB password handling
    local existing_pass=$(get_existing_password "mariadb.password")
    if [[ -n "${existing_pass}" ]]; then
        MARIADB_PASSWORD="${existing_pass}"
        log_info "Using existing MariaDB password from config.yaml"
    else
        MARIADB_PASSWORD=$(generate_random_password 10)
        log_info "Generated random 10-character MariaDB password"
    fi

    local existing_root_pass=$(get_existing_password "mariadb.root_password")
    if [[ -n "${existing_root_pass}" ]]; then
        MARIADB_ROOT_PASSWORD="${existing_root_pass}"
        log_info "Using existing MariaDB root password from config.yaml"
    else
        MARIADB_ROOT_PASSWORD=$(generate_random_password 10)
        log_info "Generated random 10-character MariaDB root password"
    fi

    if [[ -z "${MARIADB_IMAGE}" ]]; then
        MARIADB_IMAGE="$(image_from_registry "${MARIADB_IMAGE_REPOSITORY}" "${MARIADB_IMAGE_TAG}" "${MARIADB_IMAGE_FALLBACK}")"
    fi

    # Parse image registry/repository/tag from MARIADB_IMAGE
    local image_without_tag="${MARIADB_IMAGE%:*}"
    local image_tag="${MARIADB_IMAGE##*:}"
    local image_registry="${image_without_tag%%/*}"
    local image_repo="${image_without_tag#*/}"

    local chart_ref="bitnami/mariadb"
    local use_local_chart="false"
    if [[ -f "${MARIADB_CHART_TGZ}" ]]; then
        chart_ref="${MARIADB_CHART_TGZ}"
        use_local_chart="true"
        log_info "Using local MariaDB chart: ${chart_ref}"
    else
        log_info "Using remote MariaDB chart: ${chart_ref} (version ${MARIADB_CHART_VERSION})"
    fi

    kubectl create namespace "${ns}" 2>/dev/null || true

    local persistence_enabled="${MARIADB_PERSISTENCE_ENABLED}"
    if [[ "${persistence_enabled}" == "true" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            log_warn "No StorageClass found. MariaDB PVC will stay Pending."
            if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                install_localpv
            fi
            if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
                log_warn "Still no StorageClass found. Disabling MariaDB persistence to avoid Pending PVC."
                persistence_enabled="false"
            fi
        fi
    fi

    if [[ "${use_local_chart}" != "true" ]]; then
        helm repo add --force-update bitnami "${HELM_REPO_BITNAMI}"
        helm repo update
    fi

    local -a helm_args
    helm_args=(
        upgrade --install mariadb "${chart_ref}"
        --namespace "${ns}"
        --set image.registry="${image_registry}"
        --set image.repository="${image_repo}"
        --set image.tag="${image_tag}"
        --set architecture=standalone
        --set auth.rootPassword="${MARIADB_ROOT_PASSWORD}"
        --set auth.database="${MARIADB_DATABASE}"
        --set auth.username="${MARIADB_USER}"
        --set auth.password="${MARIADB_PASSWORD}"
        --set primary.extraFlags="--max-connections=${MARIADB_MAX_CONNECTIONS}"
        --wait --timeout=600s
    )

    if [[ "${use_local_chart}" != "true" ]]; then
        helm_args+=(--version "${MARIADB_CHART_VERSION}")
    fi

    if [[ "${persistence_enabled}" == "true" ]]; then
        helm_args+=(
            --set primary.persistence.enabled=true
            --set primary.persistence.size="${MARIADB_STORAGE_SIZE}"
        )
        if [[ -n "${MARIADB_STORAGE_CLASS}" ]]; then
            helm_args+=(--set primary.persistence.storageClass="${MARIADB_STORAGE_CLASS}")
        fi
    else
        helm_args+=(--set primary.persistence.enabled=false)
    fi

    helm "${helm_args[@]}"

    # Wait for MariaDB Pod to be ready
    log_info "Waiting for MariaDB Pod to be ready..."
    kubectl wait --for=condition=ready pod -l "app.kubernetes.io/instance=mariadb" -n "${ns}" --timeout=300s 2>/dev/null || {
        log_warn "MariaDB Pod may not be ready yet"
    }

    log_info "MariaDB installed successfully"
    log_info "MariaDB connection info:"
    log_info "  Host: mariadb.${ns}.svc.cluster.local"
    log_info "  Port: 3306"
    log_info "  Root Password: ${MARIADB_ROOT_PASSWORD}"
    log_info "  Database: ${MARIADB_DATABASE}"
    log_info "  Username: ${MARIADB_USER}"
    log_info "  Password: ${MARIADB_PASSWORD}"

    # Create additional databases and grant permissions
    setup_mariadb_databases

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Updating conf/config.yaml after MariaDB fresh install..."
        generate_config_yaml
        # Update RDS type to internal since MariaDB is installed internally
        update_rds_type_to_internal
    fi
}

install_mariadb() {
    if [[ -z "${MARIADB_IMAGE}" ]]; then
        MARIADB_IMAGE="$(image_from_registry "${MARIADB_IMAGE_REPOSITORY}" "${MARIADB_IMAGE_TAG}" "${MARIADB_IMAGE_FALLBACK}")"
    fi

    # Use proton-mariadb Helm chart by default
    install_mariadb_helm
}

uninstall_mariadb() {
    local ns="${MARIADB_NAMESPACE}"
    log_info "Uninstalling MariaDB from namespace ${ns}..."

    # Use Helm to uninstall the release (automatically removes all managed resources)
    if helm status mariadb -n "${ns}" >/dev/null 2>&1; then
        log_info "Removing Helm release: mariadb"
        helm uninstall mariadb -n "${ns}"
        log_warn "Deleting MariaDB PVCs (data loss!)"
        kubectl delete pvc -n "${ns}" -l app.kubernetes.io/instance=mariadb 2>/dev/null || true
        kubectl delete pvc -n "${ns}" -l app.kubernetes.io/name=mariadb 2>/dev/null || true
        kubectl delete pvc -n "${ns}" data-mariadb-0 2>/dev/null || true
    else
        log_info "Helm release mariadb not found, skipping Helm uninstall"
    fi

    log_info "MariaDB uninstall done"
}
