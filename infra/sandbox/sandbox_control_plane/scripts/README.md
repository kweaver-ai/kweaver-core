# Sandbox Platform Setup Scripts

This directory contains utility scripts for setting up and configuring the Sandbox Platform.

## Scripts

### install_s3fs.sh (Optional Development Tool)

Installs s3fs-fuse on the host system for S3 workspace mounting.

**Important**: This is an **optional development tool** and is **NOT required for production**.

**Purpose**: Mount S3 buckets (including MinIO) to Docker containers using FUSE, allowing user code to access uploaded files directly via the filesystem.

**Usage**:
```bash
sudo ./scripts/install_s3fs.sh
```

**Supported OS**:
- Ubuntu/Debian Linux
- macOS (via Homebrew)

**What it does**:
1. Detects the operating system
2. Installs s3fs-fuse package
3. Verifies installation
4. Checks /dev/fuse device availability

**After installation**:
- The host system can mount S3 buckets using s3fs
- Docker containers can be configured to mount S3 buckets via entrypoint scripts
- The executor image should include s3fs (or have it mounted from the host)

## How S3 Mounting Works

**Important**: The `s3fs` command runs **inside the container**, not on the host.

When `docker_scheduler.py` creates a container with S3 workspace, it:
1. Adds `/dev/fuse` device to the container
2. Adds `SYS_ADMIN` capability (required for FUSE)
3. Creates an entrypoint script that runs `s3fs` **inside the container**
4. The entrypoint script mounts the S3 bucket before starting the executor

Therefore, **s3fs must be installed inside the executor Docker image**.

## Executor Docker Image Requirements

For S3 workspace mounting to work, the executor Docker image needs:

1. **s3fs binary**: Must be installed in the image (see Dockerfile below)
2. **FUSE support**: /dev/fuse device accessible in container (added by scheduler)
3. **Shell**: /bin/sh for the entrypoint script

### Option 1: Include s3fs in executor image (Recommended)

Update the executor Dockerfile:

```dockerfile
FROM python:3.11-slim

# Install s3fs for S3 workspace mounting
RUN apt-get update && apt-get install -y \
    s3fs \
    && rm -rf /var/lib/apt/lists/*

# Install executor application
COPY executor/ /usr/local/bin/sandbox-executor/

# Create workspace directory
RUN mkdir /workspace && \
    chown -R 1000:1000 /workspace

USER 1000:1000
WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/sandbox-executor"]
```

Build and push:
```bash
docker build -t sandbox-executor:latest -f executor/Dockerfile .
docker push sandbox-executor:latest
```

### Option 2: Mount s3fs from host

If s3fs is installed on the host, mount it into the container:

In docker-compose.yml (in project root):
```yaml
services:
  control-plane:
    volumes:
      - /usr/bin/s3fs:/usr/bin/s3fs:ro
```

The Docker scheduler will automatically add this mount when creating containers with S3 workspace.

## Verification

After installing s3fs, verify the setup:

```bash
# 1. Check s3fs is available
which s3fs
s3fs --version

# 2. Check FUSE device
ls -l /dev/fuse

# 3. Test MinIO connection (if using docker-compose)
docker-compose up -d minio
curl -I http://localhost:9000/minio/health/live

# 4. Create a test session and verify S3 mount
# (This requires the control plane to be running)
```

## Troubleshooting

### s3fs mount fails in container

1. **Check /dev/fuse is accessible**:
   ```bash
   docker exec <container-id> ls -l /dev/fuse
   ```

2. **Check SYS_ADMIN capability**:
   ```bash
   docker inspect <container-id> | grep CapAdd
   # Should include "SYS_ADMIN"
   ```

3. **Check s3fs logs**:
   ```bash
   docker logs <container-id>
   ```

4. **Check S3 credentials**:
   ```bash
   docker exec <container-id> cat /etc/passwd-s3fs
   ```

### Permission denied errors

- Ensure container runs as a user that can access FUSE
- The executor uses UID:GID 1000:1000
- FUSE operations typically work for non-root users

### MinIO connection errors

- Check MinIO is running: `docker-compose ps minio`
- Check endpoint URL in environment variables
- For MinIO, ensure `use_path_request_style` option is used

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Host System                             │
│  ┌────────────┐  ┌────────────┐  ┌──────────────────────┐ │
│  │ Docker     │  │ s3fs       │  │ MinIO (Optional)      │ │
│  │ Daemon    │  │ Binary     │  │ Port 9000/9001       │ │
│  └─────┬──────┘  └──────┬─────┘  └──────────────────────┘ │
└────────┼────────────────┼────────────────────────────────┘
         │                │
         │ /dev/fuse      │ s3:// mount
         │                │
┌────────▼────────────────▼────────────────────────────────┐
│                Docker Container                           │
│  ┌────────────────────────────────────────────────────┐  │
│  │ /dev/fuse (device)                              │  │
│  │ /workspace (s3fs mount point)                    │  │
│  │ └── sessions/{session_id}/                        │  │
│  │     ├── uploads/                                  │  │
│  │     └── artifacts/                                 │  │
│  └────────────────────────────────────────────────────┘  │
│                                                              │
│  User code can read/write files in /workspace as if local │
└────────────────────────────────────────────────────────────┘
```

## Security Considerations

### SYS_ADMIN Capability

S3 mounting requires `SYS_ADMIN` capability, which:
- Allows FUSE filesystem operations
- Is a privileged capability (though less than full privileged mode)
- Is balanced by `no-new-privileges` security option
- Is only added when S3 workspace is configured

### S3 Credentials

S3 credentials are passed to containers via:
1. Environment variables (in container config)
2. Written to `/etc/passwd-s3fs` (in entrypoint script)

For production, consider:
- Using IAM roles for Kubernetes (IRSA) instead of credentials
- Using Docker secrets or Swarm configs for credential storage
- Short-lived credentials with automatic rotation

### Resource Limits

S3FS uses memory for caching:
- Tmpfs at `/tmp` (100M) for s3fs cache
- Consider adjusting size based on workload
- Monitor container memory usage

## References

- [s3fs-fuse GitHub](https://github.com/s3fs-fuse/s3fs-fuse)
- [MinIO Documentation](https://min.io/docs/minio/linux/index.html)
- [FUSE (Filesystem in Userspace)](https://github.com/libfuse/libfuse)
