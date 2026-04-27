FROM golang:1.24.13 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/agent-observability .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /out/agent-observability /app/agent-observability
COPY --from=builder /workspace/docs/swagger /app/docs/swagger

EXPOSE 8080

ENTRYPOINT ["/app/agent-observability"]
