
# Create MongoDB databases (osssys, automation)
# Initialize MongoDB replica set
setup_mongodb_replicaset() {
    local ns="${MONGODB_NAMESPACE}"
    local pod_name="${MONGODB_RELEASE_NAME}-mongodb-0"
    local mongodb_user="${MONGODB_SECRET_USERNAME}"
    local mongodb_password="${MONGODB_SECRET_PASSWORD}"
    local replicas="${MONGODB_REPLICAS:-1}"
    local replset_name="${MONGODB_REPLSET_NAME:-rs0}"
    
    log_info "Initializing MongoDB replica set: ${replset_name} with ${replicas} member(s)..."
    
    # Temporarily disable set -e to prevent script from exiting on detection failures
    set +e
    
    if [[ -z "${mongodb_password}" ]]; then
        mongodb_password=$(kubectl -n "${ns}" get secret "${MONGODB_SECRET_NAME}" -o jsonpath='{.data.password}' 2>/dev/null | base64 -d 2>/dev/null)
    fi
    
    if [[ -z "${mongodb_password}" ]]; then
        log_warn "MongoDB password not found, skipping replica set initialization"
        set -e
        return 0
    fi
    
    # Wait for pod(s) to be ready using kubectl wait
    log_info "Waiting for MongoDB pod(s) to be ready..."
    if kubectl -n "${ns}" wait --for=condition=Ready pod -l "app=${MONGODB_RELEASE_NAME}-mongodb" --timeout=120s 2>/dev/null; then
        log_info "All ${replicas} MongoDB pod(s) are ready"
    else
        log_warn "MongoDB pod(s) may not be ready, but continuing with replica set initialization..."
    fi
    
    # Detect mongo tool
    log_info "Detecting MongoDB shell tool..."
    local mongo_tool=""
    local tool_attempt=0
    while [[ $tool_attempt -lt 10 ]]; do
        if kubectl -n "${ns}" exec "${pod_name}" -c mongodb -- mongosh --version >/dev/null 2>&1; then
            mongo_tool="mongosh"
            break
        elif kubectl -n "${ns}" exec "${pod_name}" -c mongodb -- mongo --version >/dev/null 2>&1; then
            mongo_tool="mongo"
            break
        fi
        tool_attempt=$((tool_attempt + 1))
        sleep 2
    done
    
    if [[ -z "${mongo_tool}" ]]; then
        log_warn "Neither mongo nor mongosh found, skipping replica set initialization"
        set -e
        return 0
    fi
    log_info "Detected MongoDB tool: ${mongo_tool}"
    
    # Check if replica set is already initialized by trying to get status with authentication
    log_info "Checking if replica set is already initialized..."
    local rs_status
    rs_status=$(kubectl -n "${ns}" exec "${pod_name}" -c mongodb -- ${mongo_tool} \
        --quiet \
        --port 28000 \
        -u "${mongodb_user}" \
        -p "${mongodb_password}" \
        --authenticationDatabase admin \
        --eval "try { var s = rs.status(); if (s.ok === 1) { print('INITIALIZED'); } else { print('NOT_INITIALIZED'); } } catch(e) { print('NOT_INITIALIZED'); }" 2>/dev/null | tail -1)
    
    if [[ "${rs_status}" == "INITIALIZED" ]]; then
        log_info "Replica set ${replset_name} is already initialized"
        set -e
        return 0
    fi
    
    # Build members array
    log_info "Building replica set members configuration..."
    local service_name="${MONGODB_RELEASE_NAME}-mongodb"
    local members_js=""
    local i=0
    while [[ $i -lt $replicas ]]; do
        local member_host="${service_name}-${i}.${service_name}.${ns}.svc.cluster.local:28000"
        # For single-node replica set, use priority 1 (default)
        # For multi-node, first node gets priority 2 (primary preference)
        if [[ $i -gt 0 ]]; then
            members_js="${members_js}, "
        fi
        if [[ $i -eq 0 ]] && [[ $replicas -gt 1 ]]; then
            members_js="${members_js}{ _id: ${i}, host: \"${member_host}\", priority: 2 }"
        else
            members_js="${members_js}{ _id: ${i}, host: \"${member_host}\" }"
        fi
        i=$((i + 1))
    done
    
    log_info "Initializing replica set with members: ${members_js}"
    
    # First, create the initial admin user (this works with localhost exception before replica set init)
    log_info "Creating initial admin user..."
    kubectl -n "${ns}" exec -i "${pod_name}" -c mongodb -- ${mongo_tool} \
        --quiet \
        --port 28000 <<EOF
try {
    var adminDB = db.getSiblingDB('admin');
    adminDB.createUser({
        user: '${mongodb_user}',
        pwd: '${mongodb_password}',
        roles: [{role: 'root', db: 'admin'}]
    });
    print("✓ Initial admin user created successfully");
} catch(e) {
    if (e.message && e.message.indexOf("already exists") !== -1) {
        print("Admin user already exists, continuing...");
    } else {
        print("Note: Could not create admin user: " + e.message);
    }
}
EOF
    
    # Now initialize replica set with authentication
    log_info "Executing rs.initiate() and creating databases..."
    
    # Format the JS array for databases
    local db_list_js
    db_list_js=$(printf "'%s'," "osssys" "automation" | sed 's/,$//')
    
    kubectl -n "${ns}" exec -i "${pod_name}" -c mongodb -- ${mongo_tool} \
        --quiet \
        --port 28000 \
        -u "${mongodb_user}" \
        -p "${mongodb_password}" \
        --authenticationDatabase admin <<EOF
try {
    var cfg = {
        _id: "${replset_name}",
        members: [${members_js}]
    };
    var result = rs.initiate(cfg);
    print("Replica set initialization command executed");
    print("Result: " + JSON.stringify(result));
    
    // Wait a moment for replica set to initialize
    sleep(5000);
    
    // Create databases and grant permissions to admin user
    try {
        var dbs = [${db_list_js}];
        for (var i = 0; i < dbs.length; i++) {
            var dbName = dbs[i];
            var dbObj = db.getSiblingDB(dbName);
            dbObj.healthcheck.insert({init: true, timestamp: new Date()});
            print("✓ Ensured database exists: " + dbName);
        }
        
        // Grant permissions to admin user for osssys and automation databases
        var adminDB = db.getSiblingDB('admin');
        var userRoles = [];
        dbs.forEach(function(dbName) {
            userRoles.push({role: 'dbOwner', db: dbName});
            userRoles.push({role: 'readWrite', db: dbName});
        });
        
        adminDB.grantRolesToUser('${mongodb_user}', userRoles);
        print("✓ Permissions granted to admin user for databases: " + dbs.join(', '));
    } catch(e) {
        print("Note: Could not create databases/grant permissions: " + e.message);
    }
} catch(e) {
    print("Error initializing replica set: " + e);
    if (e.message && (e.message.indexOf("already initialized") !== -1 || e.message.indexOf("already been initiated") !== -1)) {
        print("Replica set already initialized, continuing...");
    } else {
        print("Unexpected error, re-throwing...");
        throw e;
    }
}
EOF
    
    local init_rc=$?
    set -e
    
    if [[ $init_rc -eq 0 ]]; then
        log_info "Replica set initialization command executed successfully"
        log_info "Waiting for replica set to elect PRIMARY (this may take 30-60 seconds)..."
        
        # Wait and verify replica set has a PRIMARY
        local verify_attempts=0
        local max_verify_attempts=60
        local has_primary=false
        
        while [[ $verify_attempts -lt $max_verify_attempts ]]; do
            sleep 2
            set +e
            local status_check
            # Check if there's a PRIMARY in the replica set
            status_check=$(kubectl -n "${ns}" exec "${pod_name}" -c mongodb -- ${mongo_tool} \
                --quiet \
                --port 28000 \
                -u "${mongodb_user}" \
                -p "${mongodb_password}" \
                --authenticationDatabase admin \
                --eval "try { var s = rs.status(); var hasPrimary = false; if (s && s.members) { for (var i = 0; i < s.members.length; i++) { if (s.members[i].state === 1) { hasPrimary = true; break; } } } print(hasPrimary ? 'PRIMARY_FOUND' : 'NO_PRIMARY'); } catch(e) { print('ERROR: ' + e.message); }" 2>/dev/null | tail -1)
            set -e
            
            if [[ "${status_check}" == *"PRIMARY_FOUND"* ]]; then
                has_primary=true
                break
            fi
            
            verify_attempts=$((verify_attempts + 1))
            if [[ $((verify_attempts % 10)) -eq 0 ]]; then
                log_info "Waiting for PRIMARY election... (attempt ${verify_attempts}/${max_verify_attempts})"
            fi
        done
        
        if [[ "${has_primary}" == "true" ]]; then
            log_info "✓ Replica set ${replset_name} has elected PRIMARY successfully"
        else
            log_warn "Replica set initialization command executed, but PRIMARY election timed out."
            log_warn "The replica set may still be initializing. Continuing anyway..."
        fi
    else
        log_warn "Replica set initialization command returned non-zero exit code, but continuing..."
        log_warn "Please check replica set status manually"
    fi
}

setup_mongodb_databases() {
    local ns="${MONGODB_NAMESPACE}"
    local pod_name="${MONGODB_RELEASE_NAME}-mongodb-0"
    local mongodb_user="${MONGODB_SECRET_USERNAME}"
    local mongodb_password="${MONGODB_SECRET_PASSWORD}"
    
    # Temporarily disable set -e to prevent script from exiting on detection failures
    set +e
    
    if [[ -z "${mongodb_password}" ]]; then
        mongodb_password=$(kubectl -n "${ns}" get secret "${MONGODB_SECRET_NAME}" -o jsonpath='{.data.password}' 2>/dev/null | base64 -d 2>/dev/null)
    fi
    
    if [[ -z "${mongodb_password}" ]]; then
        log_warn "MongoDB password not found, skipping database creation"
        set -e
        return 0
    fi
    
    log_info "Setting up MongoDB databases..."
    
    # Detection loop for mongo or mongosh
    log_info "Waiting for MongoDB to be ready and detecting shell tool..."
    local max_attempts=30
    local attempt=0
    local mongo_tool=""
    
    while [[ $attempt -lt $max_attempts ]]; do
        # We use '|| true' to be absolutely sure set -e doesn't catch these
        if kubectl -n "${ns}" exec "${pod_name}" -c mongodb -- mongosh --version >/dev/null 2>&1; then
            mongo_tool="mongosh"
            break
        elif kubectl -n "${ns}" exec "${pod_name}" -c mongodb -- mongo --version >/dev/null 2>&1; then
            mongo_tool="mongo"
            break
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    if [[ -z "${mongo_tool}" ]]; then
        log_warn "Neither mongo nor mongosh found in container ${pod_name}, skipping DB setup"
        set -e
        return 0
    fi
    log_info "Detected MongoDB tool: ${mongo_tool}"
    
    # List of databases to create
    local databases=(
        "osssys"
        "automation"
    )
    
    log_info "MongoDB databases and users already created during replica set initialization"
    log_info "MongoDB database setup completed"
}

# Install MongoDB via Helm
install_mongodb() {
    log_info "Installing MongoDB via Helm..."

    kubectl create namespace "${MONGODB_NAMESPACE}" 2>/dev/null || true

    local fresh_install="true"
    if is_helm_installed "${MONGODB_RELEASE_NAME}" "${MONGODB_NAMESPACE}"; then
        fresh_install="false"
        log_info "MongoDB is already installed (Helm release exists). Skipping installation."
        return 0
    fi

    # Check for StorageClass if persistence is enabled
    local storage_class="${MONGODB_STORAGE_CLASS}"
    if [[ -z "${storage_class}" ]]; then
        if [[ -z "$(kubectl get storageclass --no-headers 2>/dev/null)" ]]; then
            log_warn "No StorageClass found. MongoDB PVC will stay Pending."
            if [[ "${AUTO_INSTALL_LOCALPV}" == "true" ]]; then
                install_localpv
            fi
            # Try to use local-path if available
            if kubectl get storageclass local-path >/dev/null 2>&1; then
                storage_class="local-path"
                log_info "Using local-path StorageClass for MongoDB"
            fi
        else
            # Use first available StorageClass
            storage_class="$(kubectl get storageclass -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)"
            log_info "Using StorageClass: ${storage_class}"
        fi
    fi

    # Create or check MongoDB secret
    if ! kubectl -n "${MONGODB_NAMESPACE}" get secret "${MONGODB_SECRET_NAME}" >/dev/null 2>&1; then
        log_info "Creating MongoDB secret..."
        local mongodb_password="${MONGODB_SECRET_PASSWORD}"
        if [[ -z "${mongodb_password}" ]]; then
            mongodb_password="$(generate_random_password 10)"
            log_info "Generated MongoDB password"
        fi
        MONGODB_SECRET_PASSWORD="${mongodb_password}"
        local secret_args=(
            --from-literal=username="${MONGODB_SECRET_USERNAME}"
            --from-literal=password="${mongodb_password}"
        )
        local mongodb_keyfile
        # MongoDB keyfile must be <= 1024 bytes (not base64 encoded size)
        # Generate 768 bytes of random data, base64 encode it (will be ~1024 chars)
        # Then truncate to exactly 1024 bytes to be safe
        mongodb_keyfile="$(openssl rand -base64 768 | tr -d '\n' | head -c 1024)"
        secret_args+=(--from-literal=mongodb.keyfile="${mongodb_keyfile}")
        kubectl create secret generic "${MONGODB_SECRET_NAME}" \
            "${secret_args[@]}" \
            -n "${MONGODB_NAMESPACE}" 2>/dev/null || {
            log_error "Failed to create MongoDB secret"
            return 1
        }
        log_info "MongoDB secret created"
    else
        log_info "MongoDB secret already exists"
        # Ensure mongodb.keyfile exists in the Secret (older installs may not have it)
        local existing_keyfile_b64
        existing_keyfile_b64="$(kubectl -n "${MONGODB_NAMESPACE}" get secret "${MONGODB_SECRET_NAME}" -o jsonpath='{.data.mongodb\.keyfile}' 2>/dev/null || true)"
        if [[ -z "${existing_keyfile_b64}" ]]; then
            log_warn "MongoDB secret ${MONGODB_SECRET_NAME} has no mongodb.keyfile; patching secret for replica set auth..."
            local mongodb_keyfile
            # MongoDB keyfile must be <= 1024 bytes (not base64 encoded size)
            # Generate 768 bytes of random data, base64 encode it (will be ~1024 chars)
            # Then truncate to exactly 1024 bytes to be safe
            mongodb_keyfile="$(openssl rand -base64 768 | tr -d '\n' | head -c 1024)"
            local mongodb_keyfile_b64
            mongodb_keyfile_b64="$(printf '%s' "${mongodb_keyfile}" | base64 | tr -d '\n')"
            kubectl -n "${MONGODB_NAMESPACE}" patch secret "${MONGODB_SECRET_NAME}" --type merge \
                -p "{\"data\":{\"mongodb.keyfile\":\"${mongodb_keyfile_b64}\"}}" >/dev/null 2>&1 || {
                log_error "Failed to patch mongodb.keyfile into secret ${MONGODB_SECRET_NAME}"
                return 1
            }
            log_info "Patched mongodb.keyfile into secret ${MONGODB_SECRET_NAME}"
        fi
    fi

    # Prepare chart tgz
    local chart_path="${MONGODB_CHART_TGZ}"
    if [[ ! -f "${chart_path}" ]]; then
        log_error "MongoDB chart tgz not found at: ${chart_path}"
        log_error "Please ensure the MongoDB chart tgz is available at ${chart_path}"
        return 1
    fi

    # Prepare values
    local mongodb_image="${MONGODB_IMAGE}"
    if [[ -z "${mongodb_image}" ]]; then
        mongodb_image="${MONGODB_IMAGE_REPOSITORY}:${MONGODB_IMAGE_TAG}"
    fi

    # Parse image repository/tag
    local image_repo="${mongodb_image%:*}"
    local image_tag="${mongodb_image##*:}"
    if [[ "${image_repo}" == "${mongodb_image}" || -z "${image_tag}" ]]; then
        log_error "Invalid MONGODB_IMAGE (expected repo:tag): ${mongodb_image}"
        return 1
    fi

    # Build Helm values
    local helm_values=(
        "mongodb.image.repository=${image_repo}"
        "mongodb.image.tag=${image_tag}"
        "mongodb.replicas=${MONGODB_REPLICAS}"
        "mongodb.replSet.enabled=true"
        "mongodb.replSet.name=${MONGODB_REPLSET_NAME}"
        "mongodb.service.type=${MONGODB_SERVICE_TYPE}"
        "mongodb.service.port=${MONGODB_SERVICE_PORT}"
        "mongodb.conf.wiredTigerCacheSizeGB=${MONGODB_WIRED_TIGER_CACHE_SIZE_GB}"
        "mongodb.resources.requests.cpu=${MONGODB_RESOURCES_REQUESTS_CPU}"
        "mongodb.resources.requests.memory=${MONGODB_RESOURCES_REQUESTS_MEMORY}"
        "mongodb.resources.limits.cpu=${MONGODB_RESOURCES_LIMITS_CPU}"
        "mongodb.resources.limits.memory=${MONGODB_RESOURCES_LIMITS_MEMORY}"
        "storage.capacity=${MONGODB_STORAGE_SIZE}"
        "secret.name=${MONGODB_SECRET_NAME}"
        "secret.createSecret=false"
    )

    if [[ -n "${storage_class}" ]]; then
        helm_values+=("storage.storageClassName=${storage_class}")
    else
        helm_values+=("storage.storageClassName=")
    fi

    # Install via Helm
    log_info "Installing MongoDB Helm chart..."
    log_info "Chart path: ${chart_path}"
    log_info "Release name: ${MONGODB_RELEASE_NAME}"
    log_info "Namespace: ${MONGODB_NAMESPACE}"
    
    # Build helm command with all values
    local helm_cmd=(
        helm upgrade --install "${MONGODB_RELEASE_NAME}" "${chart_path}"
        --namespace "${MONGODB_NAMESPACE}"
    )
    
    # Add all set values
    for val in "${helm_values[@]}"; do
        helm_cmd+=(--set "${val}")
    done
    
    helm_cmd+=(--wait --timeout=600s)
    
    log_info "Running: ${helm_cmd[*]}"
    
    # Execute helm command and capture output
    local helm_output
    helm_output=$("${helm_cmd[@]}" 2>&1)
    local helm_exit_code=$?
    
    if [[ ${helm_exit_code} -ne 0 ]]; then
        log_error "Failed to install MongoDB"
        log_error "Helm command exit code: ${helm_exit_code}"
        log_error "Helm output:"
        echo "${helm_output}" | while IFS= read -r line; do
            log_error "  ${line}"
        done
        
        # Try to get more details from helm status if release exists
        if helm status "${MONGODB_RELEASE_NAME}" -n "${MONGODB_NAMESPACE}" >/dev/null 2>&1; then
            log_info "Release exists, checking status..."
            helm status "${MONGODB_RELEASE_NAME}" -n "${MONGODB_NAMESPACE}" 2>&1 | while IFS= read -r line; do
                log_info "  ${line}"
            done
        fi
        
        # Check for pod issues
        log_info "Checking for pod issues..."
        if kubectl -n "${MONGODB_NAMESPACE}" get pods -l "app=${MONGODB_RELEASE_NAME}-mongodb" >/dev/null 2>&1; then
            kubectl -n "${MONGODB_NAMESPACE}" get pods -l "app=${MONGODB_RELEASE_NAME}-mongodb" 2>&1 | while IFS= read -r line; do
                log_info "  ${line}"
            done
            
            # Get pod events
            local pod_name
            pod_name=$(kubectl -n "${MONGODB_NAMESPACE}" get pods -l "app=${MONGODB_RELEASE_NAME}-mongodb" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
            if [[ -n "${pod_name}" ]]; then
                log_info "Pod events for ${pod_name}:"
                kubectl -n "${MONGODB_NAMESPACE}" describe pod "${pod_name}" 2>&1 | grep -A 20 "Events:" | while IFS= read -r line; do
                    log_info "  ${line}"
                done || true
            fi
        fi
        
        return 1
    else
        log_info "Helm installation output:"
        echo "${helm_output}" | while IFS= read -r line; do
            log_info "  ${line}"
        done
    fi

    log_info "MongoDB installed successfully"
    
    # Wait for MongoDB to be ready before initializing replica set and creating databases
    log_info "Waiting for MongoDB Pod(s) to be ready..."
    kubectl wait --for=condition=ready pod -l "app=${MONGODB_RELEASE_NAME}-mongodb" -n "${MONGODB_NAMESPACE}" --timeout=300s 2>/dev/null || {
        log_warn "MongoDB Pod(s) may not be ready yet"
    }
    
    log_info "Initializing MongoDB replica set..."
    setup_mongodb_replicaset
    
    # Create databases
    setup_mongodb_databases
    
    log_info "MongoDB connection info:"
    log_info "  Service: ${MONGODB_RELEASE_NAME}-mongodb.${MONGODB_NAMESPACE}.svc.cluster.local"
    log_info "  Port: 28000"
    log_info "  Username: ${MONGODB_SECRET_USERNAME}"
    local secret_password
    secret_password=$(kubectl -n "${MONGODB_NAMESPACE}" get secret "${MONGODB_SECRET_NAME}" -o jsonpath='{.data.password}' 2>/dev/null | base64 -d 2>/dev/null || echo "")
    if [[ -n "${secret_password}" ]]; then
        log_info "  Password: ${secret_password}"
    else
        log_info "  Password: (check secret ${MONGODB_SECRET_NAME})"
    fi
    log_info "  AuthSource: admin"
    log_info "  ReplicaSet: ${MONGODB_REPLSET_NAME}"

    if [[ "${fresh_install}" == "true" && "${AUTO_GENERATE_CONFIG}" == "true" ]]; then
        log_info "Updating conf/config.yaml after MongoDB fresh install..."
        generate_config_yaml
    fi
}

# Uninstall MongoDB
uninstall_mongodb() {
    log_info "Uninstalling MongoDB from namespace ${MONGODB_NAMESPACE}..."

    helm_uninstall_safe "${MONGODB_RELEASE_NAME}" "${MONGODB_NAMESPACE}"

    # Delete PVCs by default (MongoDB PVCs are deleted on uninstall)
    log_info "Deleting MongoDB PVCs..."
    
    # Get all PVCs related to MongoDB before deletion
    local pvc_names
    pvc_names=$(kubectl get pvc -n "${MONGODB_NAMESPACE}" -l "app=${MONGODB_RELEASE_NAME}-mongodb" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null || true)
    
    # Delete PVCs by label
    kubectl delete pvc -n "${MONGODB_NAMESPACE}" -l "app=${MONGODB_RELEASE_NAME}-mongodb" 2>/dev/null || true
    
    # Wait a bit for PVs to be released
    sleep 2
    
    # Try to delete PVs that are in Released state
    log_info "Cleaning up Released PVs..."
    local released_pvs
    released_pvs=$(kubectl get pv -o jsonpath='{.items[?(@.status.phase=="Released")].metadata.name}' 2>/dev/null || true)
    if [[ -n "${released_pvs}" ]]; then
        for pv in ${released_pvs}; do
            # Check if this PV was bound to one of the deleted PVCs
            local pv_claim
            pv_claim=$(kubectl get pv "${pv}" -o jsonpath='{.spec.claimRef.name}' 2>/dev/null || true)
            if [[ -n "${pv_claim}" ]] && echo "${pvc_names}" | grep -q "${pv_claim}"; then
                log_info "Deleting Released PV: ${pv}"
                kubectl delete pv "${pv}" 2>/dev/null || true
            fi
        done
    fi

    log_info "MongoDB uninstall done"
}
