# Build stage
# Mirror configuration for special network environments
ARG NODE_IMAGE=node:20-alpine
ARG PRO_IMAGE=nginx:latest


FROM ${NODE_IMAGE} AS builder

# Re-declare ARG after FROM (ARGs don't persist across FROM)
ARG USE_MIRROR=false
ARG APK_MIRROR=mirrors.ustc.edu.cn

WORKDIR /app

# Configure APK mirror if needed
RUN if [ "$USE_MIRROR" = "true" ]; then \
  sed -i 's/dl-cdn.alpinelinux.org/'$APK_MIRROR'/g' /etc/apk/repositories; \
  fi

RUN npm config set registry https://registry.npmmirror.com/ && \
    npm config set fetch-retries 5 && \
    npm config set fetch-retry-mintimeout 20000 && \
    npm config set fetch-retry-maxtimeout 120000 && \
    npm config set fetch-timeout 180000

# Copy dependency manifests so Docker installs from the lockfile.
COPY package.json package-lock.json ./

# Install exactly the versions captured in package-lock.json.
RUN npm ci || (echo "Retrying npm ci..." && npm ci)

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM ${PRO_IMAGE} AS production

WORKDIR /app

# Copy built files
COPY --from=builder /app/dist /usr/share/nginx/html
COPY --from=builder /app/nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
