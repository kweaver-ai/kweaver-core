# Changelog

All notable changes to this project will be documented in this file.

## [0.2.1] - 2026-04-07

### Improvements

- Standardized the `agent-observability` and `otelcol-contrib` Helm chart image values to use `image.registry`, `image.repository`, and `image.tag`, so offline packaging can extract images from the default values.
- Updated the `agent-observability` release workflow to write the resolved release version into the chart default `image.tag` before packaging.

### Upgrade Notes

- If you override `image.repository` with a full image reference that includes the registry, split it into `image.registry` and `image.repository` to match the new chart values structure.

## [0.2.0] - 2026-03-31

### Improvements

- Updated the release workflows under `.github/workflows/` to use the relocated `trace-ai/` paths for version resolution, Go module discovery, Docker build context, and Helm chart packaging.

### Documentation

- Added a reference command in `README.md` for creating and verifying the multi-architecture manifest of `opentelemetry-collector-contrib` in SWR.

## [0.1.1] - 2026-03-27

### Improvements

- Increased the default result size limit to `1000` for the conversation-based trace search API so a single request can return more matching traces.

## [0.1.0] - 2026-03-25

Initial project release.

### Added

- Added the `agent-observability` service for querying agent traces from OpenSearch.
- Added trace query APIs for raw DSL search and conversation-based lookup, with generated Swagger documentation.
- Added Docker, Helm, and GitHub Actions workflows for building and releasing `agent-observability`.
- Added the `otelcol-contribute-chart` Helm chart for deploying OpenTelemetry Collector Contrib on Kubernetes.
- Added OTLP ingestion and OpenSearch export defaults in the collector chart to support trace and log pipelines.
- Added repository-level English and Chinese README documents describing the Tracing AI architecture, capabilities, and quick start flow.

### Documentation

- Added product and implementation documents for the Agent tracing system under `agent-observability/docs/`, including PRD, design, API schema, and Swagger assets.
