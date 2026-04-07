/**
 * API 类型定义
 * 基于 Control Plane OpenAPI 规范
 */

// ============================================
// 通用类型
// ============================================

/** 运行时类型 */
export type RuntimeType = 'python3.11' | 'nodejs20' | 'java17' | 'go1.21';

/** 会话状态 */
export type SessionStatus = 'PENDING' | 'CREATING' | 'STARTING' | 'RUNNING' | 'COMPLETED' | 'TERMINATED' | 'FAILED' | 'TIMEOUT';

/** 依赖安装状态 */
export type DependencyInstallStatus = 'pending' | 'installing' | 'completed' | 'failed';

/** 执行状态 */
export type ExecutionStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'TIMEOUT' | 'CRASHED';

/** 编程语言 */
export type CodeLanguage = 'python' | 'javascript' | 'shell';

// ============================================
// Template 相关类型
// ============================================

/** 模板响应 */
export interface TemplateResponse {
  id: string;
  name: string;
  image_url: string;
  runtime_type: RuntimeType;
  default_cpu_cores: number;
  default_memory_mb: number;
  default_disk_mb: number;
  default_timeout_sec: number;
  default_env_vars?: Record<string, string>;
  is_active: boolean;
  created_at?: string;
  updated_at?: string;
}

/** 创建模板请求 */
export interface CreateTemplateRequest {
  id: string;
  name: string;
  image_url: string;
  runtime_type: RuntimeType;
  default_cpu_cores?: number;
  default_memory_mb?: number;
  default_disk_mb?: number;
  default_timeout?: number;
  default_env_vars?: Record<string, string>;
}

/** 更新模板请求 */
export interface UpdateTemplateRequest {
  name?: string;
  image_url?: string;
  default_cpu_cores?: number;
  default_memory_mb?: number;
  default_disk_mb?: number;
  default_timeout?: number;
  default_env_vars?: Record<string, string>;
}

// ============================================
// Session 相关类型
// ============================================

/** 资源限制响应 */
export interface ResourceLimitResponse {
  cpu: string;
  memory: string;
  disk: string;
  max_processes?: number;
}

/** 会话响应 */
export interface SessionResponse {
  id: string;
  template_id: string;
  status: SessionStatus;
  resource_limit?: ResourceLimitResponse;
  workspace_path?: string;
  runtime_type: RuntimeType;
  runtime_node?: string;
  container_id?: string;
  pod_name?: string;
  env_vars: Record<string, string>;
  timeout: number;
  python_package_index_url?: string;
  requested_dependencies?: DependencySpec[];
  installed_dependencies?: InstalledDependencyResponse[];
  dependency_install_status?: DependencyInstallStatus;
  dependency_install_error?: string | null;
  dependency_install_started_at?: string | null;
  dependency_install_completed_at?: string | null;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  last_activity_at?: string;
}

/** 依赖包规范 */
export interface DependencySpec {
  name: string;
  version?: string;
}

/** 已安装依赖 */
export interface InstalledDependencyResponse {
  name: string;
  version: string;
  install_location: string;
  install_time: string;
  is_from_template?: boolean;
}

/** 创建会话请求 */
export interface CreateSessionRequest {
  template_id: string;
  timeout?: number;
  cpu?: string;
  memory?: string;
  disk?: string;
  env_vars?: Record<string, string>;
  event?: Record<string, unknown>;
  python_package_index_url?: string;
  dependencies?: DependencySpec[];
  install_timeout?: number;
  fail_on_dependency_error?: boolean;
  allow_version_conflicts?: boolean;
}

/** 安装会话依赖请求 */
export interface InstallSessionDependenciesRequest {
  python_package_index_url?: string;
  dependencies: DependencySpec[];
}

/** 会话列表响应 */
export interface SessionListResponse {
  items: SessionResponse[];
  total: number;
  limit: number;
  offset: number;
}

/** 会话列表查询参数 */
export interface ListSessionsParams {
  status?: SessionStatus | null;
  template_id?: string | null;
  limit?: number;
  offset?: number;
}

// ============================================
// Execution 相关类型
// ============================================

/** 文件制品响应 */
export interface ArtifactResponse {
  path: string;
  size: number;
  mime_type: string;
  type: string;
  created_at: string;
  checksum?: string;
}

/** 执行指标 */
export interface ExecutionMetrics {
  duration_ms: number;
  cpu_time_ms?: number;
  peak_memory_mb?: number;
  io_read_bytes?: number;
  io_write_bytes?: number;
}

/** 执行响应 */
export interface ExecutionResponse {
  id: string;
  session_id: string;
  status: ExecutionStatus;
  code?: string;
  language?: CodeLanguage;
  timeout?: number;
  exit_code?: number;
  error_message?: string;
  execution_time?: number;
  stdout?: string;
  stderr?: string;
  artifacts: ArtifactResponse[];
  retry_count: number;
  created_at?: string;
  started_at?: string;
  completed_at?: string;
  return_value?: Record<string, unknown>;
  metrics?: ExecutionMetrics;
}

/** 执行代码请求 */
export interface ExecuteCodeRequest {
  code: string;
  language: CodeLanguage;
  timeout?: number;
  event?: Record<string, unknown>;
}

/** 执行代码响应 */
export interface ExecuteCodeResponse {
  execution_id: string;
  session_id: string;
  status: ExecutionStatus;
  created_at?: string;
}

/** 执行列表响应 */
export interface ExecutionListResponse {
  items: ExecutionResponse[];
  total: number;
  limit: number;
  offset: number;
}

// ============================================
// Health 相关类型
// ============================================

/** 健康检查响应 */
export interface HealthResponse {
  status: string;
  version: string;
  uptime: number;
}

/** 详细健康检查响应 */
export interface DetailedHealthResponse extends Record<string, unknown> {
  status: string;
  version?: string;
  uptime?: number;
  dependencies?: Record<string, string>;
}

// ============================================
// File 相关类型
// ============================================

/** 文件上传响应 */
export interface FileUploadResponse {
  session_id: string;
  file_path: string;
  size: number;
}

// ============================================
// 通用 API 响应类型
// ============================================

/** API 错误响应 */
export interface ErrorResponse {
  detail: Array<{
    loc: Array<string | number>;
    msg: string;
    type: string;
  }>;
}

/** 分页参数 */
export interface PaginationParams {
  limit?: number;
  offset?: number;
}
