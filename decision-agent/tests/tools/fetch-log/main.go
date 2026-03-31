package main

import (
	_ "embed"
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed prompts/log_analysis.txt
var logAnalysisTemplate []byte

//go:embed ref/dolphin_frendly_program.txt
var dolphinDoc []byte

//go:embed ref/agent_knowledge_items.json
var knowledgeBase []byte

// LogDetail represents a single log file entry
type LogDetail struct {
	LogName    string `json:"log_name"`
	LogContent string `json:"log_content"`
}

// LogEntry represents a single log entry for a service
type LogEntry struct {
	SvcName      string      `json:"svc_name"`
	Pod          string      `json:"pod"`
	FetchTime    string      `json:"fetch_time"`
	FetchLogLines int         `json:"fecth_log_lines"`
	LogDetail    []LogDetail `json:"log_detail"`
}

// LogResult represents the collection of log entries
type LogResult []LogEntry

// AIConfig represents the AI configuration for log analysis
type AIConfig struct {
	IP    string `json:"ip"`
	Token string `json:"token"`
	Model string `json:"model"`
}

const defaultLogLines = 50
const maxLogLines = 300
const defaultServices = "agent-app,agent-executor"
const aiConfigFile = ".fetch_log_ai_config.json"
const aiConfigDir = ".fetch_log"

// Global variables for AI command line flags
var (
	aiToken *string
	aiModel *string
	aiIP    *string
)

func main() {
	// Check for help flag before flag parsing
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printUsage()
		os.Exit(0)
	}

	// Parse command line flags
	svcList := flag.String("svc_list", "", "指定要收集日志的服务名称列表，多个服务用逗号分隔 (例如: agent-factory,agent-memory)")
	listPods := flag.Bool("list", false, "列出集群中所有 Pod 的名称和命名空间，不收集日志")
	previewLogs := flag.Bool("preview", false, "在控制台中预览收集到的日志内容")
	logLines := flag.Int("logline", defaultLogLines, "获取日志的行数 (1-300，默认50)")
	containerName := flag.String("c", "", "指定容器名称（当 Pod 有多个容器时必须指定）")
	enableDebug := flag.Bool("enable_debug", false, "启用 agent-executor 调试模式 (修改 ConfigMap 并重启服务)")
	disableDebug := flag.Bool("disable_debug", false, "禁用 agent-executor 调试模式 (修改 ConfigMap 并重启服务)")
	enableAI := flag.Bool("ai", false, "启用 AI 分析功能，使用大模型分析收集到的日志")
	aiToken = flag.String("token", "", "更新 AI 配置中的 Token")
	aiModel = flag.String("model", "", "更新 AI 配置中的模型名称")
	aiIP = flag.String("ip", "", "更新 AI 配置中的 ADP 平台 IP 地址")
	flag.Parse()

	// Validate logLines parameter
	if *logLines < 1 {
		fmt.Printf("⚠️  Warning: logline must be at least 1, using default: %d\n", defaultLogLines)
		*logLines = defaultLogLines
	}
	if *logLines > maxLogLines {
		fmt.Printf("⚠️  Warning: logline cannot exceed %d, using max: %d\n", maxLogLines, maxLogLines)
		*logLines = maxLogLines
	}

	// Handle debug mode operations
	if *enableDebug && *disableDebug {
		fmt.Println("❌ Error: Cannot enable and disable debug mode at the same time")
		os.Exit(1)
	}

	if *enableDebug {
		handleDebugMode(true)
		return
	}

	if *disableDebug {
		handleDebugMode(false)
		return
	}

	// If --list flag is provided, list all pods and exit
	if *listPods {
		listAllPods()
		return
	}

	// Determine which services to fetch logs for
	var services []string
	if *svcList != "" {
		services = strings.Split(*svcList, ",")
		// Trim whitespace from each service name
		for i, svc := range services {
			services[i] = strings.TrimSpace(svc)
		}
	} else {
		services = strings.Split(defaultServices, ",")
	}

	fmt.Printf("📋 Starting log collection for services: %v\n", services)

	// Collect logs from all services
	var result LogResult
	for _, svc := range services {
		entries, err := fetchLogsForService(svc, *logLines, *containerName)
		if err != nil {
			fmt.Printf("❌ Error fetching logs for %s: %v\n", svc, err)
			continue
		}
		result = append(result, entries...)
	}

	// Generate output filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	outputFile := fmt.Sprintf("log_%s.json", timestamp)

	// Write results to JSON file
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("❌ Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Printf("❌ Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Log collection completed. Output saved to: %s\n", outputFile)
	fmt.Printf("📊 Total entries collected: %d\n", len(result))

	// If --preview flag is provided, display logs in console
	if *previewLogs {
		displayLogsPreview(result)
	}

	// If --ai flag is provided, perform AI analysis
	if *enableAI {
		performAIAnalysis(result)
	}
}

// displayLogsPreview displays collected logs in a friendly format
func displayLogsPreview(result LogResult) {
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("📋 LOG PREVIEW\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n\n")

	for i, entry := range result {
		fmt.Printf("📦 Entry %d: %s\n", i+1, entry.SvcName)
		fmt.Printf(strings.Repeat("-", 80) + "\n")
		fmt.Printf("  Pod:           %s\n", entry.Pod)
		fmt.Printf("  Fetch Time:    %s\n", entry.FetchTime)
		fmt.Printf("  Log Lines:     %d\n", entry.FetchLogLines)
		fmt.Printf("  Log Files:     %d\n", len(entry.LogDetail))
		fmt.Printf("\n")

		for j, detail := range entry.LogDetail {
			fmt.Printf("  📄 Log File %d: %s\n", j+1, detail.LogName)

			// Display first 100 lines of log content, except requests.log which shows all
			lines := strings.Split(detail.LogContent, "\n")
			maxLines := 100

			// requests.log always shows all lines (fixed at 10)
			if detail.LogName == "requests.log" {
				maxLines = len(lines)
			} else if len(lines) < maxLines {
				maxLines = len(lines)
			}

			fmt.Printf("  " + strings.Repeat("-", 76) + "\n")
			for k := 0; k < maxLines; k++ {
				fmt.Printf("  │ %s\n", lines[k])
			}

			if detail.LogName != "requests.log" && len(lines) > 100 {
				fmt.Printf("  │ ... (%d more lines)\n", len(lines)-100)
			}
			fmt.Printf("  " + strings.Repeat("-", 76) + "\n")
			fmt.Printf("  📊 Total lines: %d\n\n", len(lines))
		}

		fmt.Printf(strings.Repeat("=", 80) + "\n\n")
	}

	fmt.Printf("✅ Preview complete. Full logs saved to JSON file.\n\n")
}

// listAllPods lists all pods in the cluster
func listAllPods() {
	fmt.Println("📋 Listing all pods in the cluster...\n")

	cmd := exec.Command("kubectl", "get", "pods", "-A", "-o", "custom-columns=NAMESPACE:.metadata.namespace,NAME:.metadata.name", "--no-headers")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Error getting pods: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		os.Exit(1)
	}

	fmt.Println(string(output))
	os.Exit(0)
}

// fetchLogsForService fetches logs for all pods of a given service
func fetchLogsForService(svcName string, logLines int, containerName string) (LogResult, error) {
	var result LogResult

	fmt.Printf("\n🔍 Processing service: %s\n", svcName)

	// Get all pods for the service
	pods, err := getPodsForService(svcName)
	if err != nil {
		return nil, err
	}

	if len(pods) == 0 {
		fmt.Printf("⚠️  No pods found for service: %s\n", svcName)
		return result, nil
	}

	// If multiple pods found, prompt user to select
	var selectedPods []podInfo
	if len(pods) > 1 {
		selectedPod := selectPod(pods)
		if selectedPod == nil {
			return result, fmt.Errorf("user cancelled pod selection")
		}
		selectedPods = []podInfo{*selectedPod}
	} else {
		selectedPods = pods
		fmt.Printf("Found 1 pod for service %s: %s\n", svcName, pods[0].podName)
	}

	// Fetch logs for selected pod(s)
	for _, podInfo := range selectedPods {
		entry := LogEntry{
			SvcName:      svcName,
			Pod:          podInfo.podName,
			FetchTime:    time.Now().Format("2006-01-02 15:04:05"),
			FetchLogLines: logLines,
		}

		// Special handling for agent-executor
		if svcName == "agent-executor" {
			// Fetch console logs
			consoleLog, err := fetchConsoleLogs(podInfo.namespace, podInfo.podName, logLines, containerName)
			if err != nil {
				fmt.Printf("❌ Error fetching console logs for pod %s: %v\n", podInfo.podName, err)
				entry.LogDetail = []LogDetail{
					{LogName: "agent-executor", LogContent: fmt.Sprintf("Error: %v", err)},
				}
			} else {
				// Fetch file logs from agent-executor container
				fileLogs, err := fetchExecutorFileLogs(podInfo.namespace, podInfo.podName, logLines, containerName)
				if err != nil {
					fmt.Printf("❌ Error fetching file logs for pod %s: %v\n", podInfo.podName, err)
					// Even if file logs fail, we still have console logs
					entry.LogDetail = []LogDetail{
						{LogName: "agent-executor", LogContent: consoleLog},
					}
				} else {
					// Create separate log entries for each log file
					entry.LogDetail = []LogDetail{
						{LogName: "agent-executor.log", LogContent: fileLogs["agent-executor.log"]},
						{LogName: "dolphin.log", LogContent: fileLogs["dolphin.log"]},
						{LogName: "requests.log", LogContent: fileLogs["requests.log"]},
					}
				}
			}
		} else {
			// Standard console logs only
			logs, err := fetchConsoleLogs(podInfo.namespace, podInfo.podName, logLines, containerName)
			if err != nil {
				fmt.Printf("❌ Error fetching logs for pod %s: %v\n", podInfo.podName, err)
				entry.LogDetail = []LogDetail{
					{LogName: svcName, LogContent: fmt.Sprintf("Error: %v", err)},
				}
			} else {
				entry.LogDetail = []LogDetail{
					{LogName: svcName, LogContent: logs},
				}
			}
		}

		result = append(result, entry)
		fmt.Printf("✓ Collected logs from pod: %s\n", podInfo.podName)
	}

	return result, nil
}

// podInfo contains namespace and pod name
type podInfo struct {
	namespace string
	podName   string
}

// getPodsForService gets all pods for a given service using kubectl
func getPodsForService(svcName string) ([]podInfo, error) {
	cmd := exec.Command("kubectl", "get", "pods", "-A")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v, output: %s", err, string(output))
	}

	var pods []podInfo
	lines := strings.Split(string(output), "\n")

	// Skip header line
	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		namespace := fields[0]
		podName := fields[1]

		// Check if pod name contains the service name
		if strings.Contains(podName, svcName) {
			pods = append(pods, podInfo{
				namespace: namespace,
				podName:   podName,
			})
		}
	}

	return pods, nil
}

// fetchConsoleLogs gets console logs from a pod
func fetchConsoleLogs(namespace, podName string, lines int, containerName string) (string, error) {
	args := []string{"logs", "-n", namespace, podName, "--tail", fmt.Sprintf("%d", lines)}

	// Add container name if specified
	if containerName != "" {
		args = append(args, "-c", containerName)
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if error is about needing to specify container
		outputStr := string(output)
		if strings.Contains(outputStr, "a container name must be specified") {
			// Extract available containers from error message
			return "", parseContainerError(outputStr)
		}
		return "", fmt.Errorf("failed to get logs: %v, output: %s", err, outputStr)
	}
	return string(output), nil
}

// parseContainerError extracts available container names from kubectl error message
func parseContainerError(output string) error {
	// Parse error message like:
	// error: a container name must be specified for pod efast-5d9cf78dbd-bq4sz, choose one of: [efast efast-private] or one of the init containers: [copy-cfl-config]

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "choose one of:") {
			// Extract the container list
			parts := strings.Split(line, "choose one of:")
			if len(parts) > 1 {
				containers := strings.TrimSpace(parts[1])
				return fmt.Errorf("multiple containers found in pod. Please use -c <container_name> to specify one of:%s", containers)
			}
		}
	}
	return fmt.Errorf("failed to get logs: %s", output)
}

// fetchExecutorFileLogs gets log files from agent-executor container
func fetchExecutorFileLogs(namespace, podName string, lines int, containerName string) (map[string]string, error) {
	result := make(map[string]string)

	// requests.log always uses 10 lines, others use the specified lines
	logFileConfigs := map[string]int{
		"agent-executor.log": lines,
		"dolphin.log":        lines,
		"requests.log":       10, // Fixed at 10 lines
	}

	for logFile, fileLines := range logFileConfigs {
		args := []string{"exec", "-n", namespace, podName}

		// Add container name if specified
		if containerName != "" {
			args = append(args, "-c", containerName)
		}

		args = append(args, "--", "sh", "-c",
			fmt.Sprintf("tail -n %d log/%s 2>/dev/null || echo 'File not found'", fileLines, logFile))

		cmd := exec.Command("kubectl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			result[logFile] = fmt.Sprintf("Error: %v, Output: %s", err, string(output))
		} else {
			result[logFile] = string(output)
		}
	}

	return result, nil
}

// handleDebugMode handles enabling or disabling debug mode for agent-executor
func handleDebugMode(enable bool) {
	fmt.Println("🔧 Agent-Executor Debug Mode Control")
	fmt.Println(strings.Repeat("=", 80))

	// Find agent-executor pod
	pods, err := getPodsForService("agent-executor")
	if err != nil {
		fmt.Printf("❌ Error finding agent-executor pods: %v\n", err)
		os.Exit(1)
	}

	if len(pods) == 0 {
		fmt.Println("❌ Error: No agent-executor pods found")
		os.Exit(1)
	}

	namespace := pods[0].namespace
	podName := pods[0].podName

	fmt.Printf("\n📍 Found agent-executor pod:\n")
	fmt.Printf("  Namespace: %s\n", namespace)
	fmt.Printf("  Pod:       %s\n", podName)

	// Check current debug status
	debugEnabled, _, err := checkDebugStatus(namespace)
	if err != nil {
		fmt.Printf("❌ Error checking debug status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n📊 Current Debug Status:\n")
	fmt.Printf("  app.debug:                           %s\n", getYN(debugEnabled.debug))
	fmt.Printf("  app.log_level:                       %s\n", debugEnabled.logLevel)
	fmt.Printf("  dialog_logging.enable_dialog_logging: %s\n", getYN(debugEnabled.dialogLogging))
	fmt.Printf("  dialog_logging.use_single_log_file:  %s\n\n", getYN(debugEnabled.singleLogFile))

	if enable {
		// Enable debug mode
		if debugEnabled.debug && debugEnabled.dialogLogging && debugEnabled.singleLogFile && debugEnabled.logLevel == "debug" {
			fmt.Println("✅ Debug mode is already enabled!")
			fmt.Println("   You can proceed with using the tool to collect logs.")
			return
		}

		fmt.Println("⚠️  WARNING: This will modify the agent-executor ConfigMap and restart the service!")
		fmt.Println("   The following changes will be made:")
		fmt.Println("     - Set app.debug to true")
		fmt.Println("     - Set app.log_level to 'debug'")
		fmt.Println("     - Set dialog_logging.enable_dialog_logging to true")
		fmt.Println("     - Set dialog_logging.use_single_log_file to true")
		fmt.Println("     - Restart agent-executor pod(s)")

		if !confirmAction("Enable debug mode and restart agent-executor") {
			fmt.Println("\n❌ Operation cancelled by user")
			return
		}

		// Update ConfigMap
		if err := updateConfigMap(namespace, enable); err != nil {
			fmt.Printf("❌ Error updating ConfigMap: %v\n", err)
			os.Exit(1)
		}

		// Restart specific pod
		if err := restartExecutorPod(namespace, podName); err != nil {
			fmt.Printf("❌ Error restarting pod: %v\n", err)
			os.Exit(1)
		}

		// Wait for pod to be ready
		fmt.Println("\n⏳ Waiting for agent-executor pod to be ready...")
		if err := waitForPodReady(namespace, ""); err != nil {
			fmt.Printf("❌ Error waiting for pod to be ready: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ Debug mode enabled successfully!")
		fmt.Println("📝 You can now run the agent to reproduce the issue, then use this tool to collect logs.")

	} else {
		// Disable debug mode
		if !debugEnabled.debug && debugEnabled.logLevel == "info" && !debugEnabled.dialogLogging && !debugEnabled.singleLogFile {
			fmt.Println("✅ Debug mode is already disabled!")
			return
		}

		fmt.Println("⚠️  WARNING: This will modify the agent-executor ConfigMap and restart the service!")
		fmt.Println("   The following changes will be made:")
		fmt.Println("     - Set app.debug to false")
		fmt.Println("     - Set app.log_level to 'info'")
		fmt.Println("     - Set dialog_logging.enable_dialog_logging to false")
		fmt.Println("     - Set dialog_logging.use_single_log_file to false")
		fmt.Println("     - Restart agent-executor pod(s)")

		if !confirmAction("Disable debug mode and restart agent-executor") {
			fmt.Println("\n❌ Operation cancelled by user")
			return
		}

		// Update ConfigMap
		if err := updateConfigMap(namespace, enable); err != nil {
			fmt.Printf("❌ Error updating ConfigMap: %v\n", err)
			os.Exit(1)
		}

		// Restart specific pod
		if err := restartExecutorPod(namespace, podName); err != nil {
			fmt.Printf("❌ Error restarting pod: %v\n", err)
			os.Exit(1)
		}

		// Wait for pod to be ready
		fmt.Println("\n⏳ Waiting for agent-executor pod to be ready...")
		if err := waitForPodReady(namespace, ""); err != nil {
			fmt.Printf("❌ Error waiting for pod to be ready: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✅ Debug mode disabled successfully!")
		fmt.Println("📝 Agent-executor is now running in normal mode.")
	}
}

// DebugStatus represents the current debug configuration status
type DebugStatus struct {
	debug         bool
	logLevel      string
	dialogLogging bool
	singleLogFile bool
}

// checkDebugStatus checks the current debug status from ConfigMap
func checkDebugStatus(namespace string) (*DebugStatus, string, error) {
	cmd := exec.Command("kubectl", "get", "cm", "-n", namespace, "agent-executor-yaml-v2", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get ConfigMap: %v, output: %s", err, string(output))
	}

	// Parse JSON output (we'll keep it as string for easier manipulation)
	configMapStr := string(output)

	// Extract values using grep-like approach
	debug := strings.Contains(configMapStr, "debug: true") ||
		strings.Contains(configMapStr, "debug: true\n")
	logLevelDebug := strings.Contains(configMapStr, "log_level: \"debug\"") ||
		strings.Contains(configMapStr, "log_level: 'debug'")
	dialogLogging := strings.Contains(configMapStr, "enable_dialog_logging: true")
	singleLogFile := strings.Contains(configMapStr, "use_single_log_file: true")

	status := &DebugStatus{
		debug:         debug,
		logLevel:      "info",
		dialogLogging: dialogLogging,
		singleLogFile: singleLogFile,
	}

	if logLevelDebug {
		status.logLevel = "debug"
	}

	return status, configMapStr, nil
}

// updateConfigMap updates the ConfigMap with debug settings using kubectl edit
func updateConfigMap(namespace string, enable bool) error {
	fmt.Println("\n🔄 Updating ConfigMap...")

	// Get current ConfigMap in YAML format
	cmd := exec.Command("kubectl", "get", "cm", "-n", namespace, "agent-executor-yaml-v2", "-o", "yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get ConfigMap: %v, output: %s", err, string(output))
	}

	// Parse the YAML content
	configMapYaml := string(output)

	// Apply replacements to YAML content
	updatedYaml := configMapYaml

	if enable {
		// Enable debug mode
		updatedYaml = strings.ReplaceAll(updatedYaml, `debug: false`, `debug: true`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `log_level: "info"`, `log_level: "debug"`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `log_level: 'info'`, `log_level: 'debug'`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `enable_dialog_logging: false`, `enable_dialog_logging: true`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `use_single_log_file: false`, `use_single_log_file: true`)
	} else {
		// Disable debug mode
		updatedYaml = strings.ReplaceAll(updatedYaml, `debug: true`, `debug: false`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `log_level: "debug"`, `log_level: "info"`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `log_level: 'debug'`, `log_level: 'info'`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `enable_dialog_logging: true`, `enable_dialog_logging: false`)
		updatedYaml = strings.ReplaceAll(updatedYaml, `use_single_log_file: true`, `use_single_log_file: false`)
	}

	// Write to temporary file and apply
	tmpFile := fmt.Sprintf("/tmp/agent-executor-configmap-%d.yaml", time.Now().Unix())
	if err := os.WriteFile(tmpFile, []byte(updatedYaml), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Apply the ConfigMap
	cmd = exec.Command("kubectl", "apply", "-f", tmpFile)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply ConfigMap: %v, output: %s", err, string(output))
	}

	fmt.Printf("✅ ConfigMap updated successfully\n")
	return nil
}

// restartExecutorPod restarts a specific agent-executor pod
func restartExecutorPod(namespace, podName string) error {
	fmt.Printf("\n🔄 Restarting pod: %s\n", podName)

	cmd := exec.Command("kubectl", "delete", "pod", "-n", namespace, podName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete pod: %v, output: %s", err, string(output))
	}

	fmt.Printf("✅ Pod %s deleted, new pod will be created automatically\n", podName)
	return nil
}

// waitForPodReady waits for the pod to be in Running state
func waitForPodReady(namespace, oldPodName string) error {
	maxWait := 5 * 60 // 5 minutes
	checkInterval := 5 * time.Second
	startTime := time.Now()

	for time.Since(startTime) < time.Duration(maxWait)*time.Second {
		// Get current agent-executor pods
		pods, err := getPodsForService("agent-executor")
		if err != nil {
			return fmt.Errorf("failed to get pods: %v", err)
		}

		if len(pods) == 0 {
			return fmt.Errorf("no agent-executor pods found")
		}

		// Find a running pod (prefer the new one if oldPodName is provided)
		var targetPod string
		for _, pod := range pods {
			if pod.namespace == namespace {
				if oldPodName == "" || pod.podName != oldPodName {
					targetPod = pod.podName
					break
				}
			}
		}

		if targetPod == "" && len(pods) > 0 {
			targetPod = pods[0].podName
		}

		// Check pod status
		cmd := exec.Command("kubectl", "get", "pod", "-n", namespace, targetPod, "-o", "json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Pod might be in the process of being created
			fmt.Printf(".")
			time.Sleep(checkInterval)
			continue
		}

		// Check if pod is running and ready
		if strings.Contains(string(output), "\"phase\": \"Running\"") {
			// Check if all containers are ready
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "\"ready\": true") {
					fmt.Printf("\n✅ Pod %s is ready\n", targetPod)
					return nil
				}
			}
		}

		fmt.Printf(".")
		time.Sleep(checkInterval)
	}

	fmt.Println()
	return fmt.Errorf("timeout waiting for pod to be ready")
}

// confirmAction prompts user for confirmation
func confirmAction(action string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n⚠️  This action will: %s\n", action)
	fmt.Printf("📝 Type 'yes' to confirm, anything else to cancel: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "yes"
}

// selectPod prompts user to select a pod from a list
func selectPod(pods []podInfo) *podInfo {
	fmt.Printf("\n📋 Found %d pods matching the service name:\n", len(pods))
	fmt.Printf(strings.Repeat("-", 80) + "\n")
	fmt.Printf("  No.  | Pod Name                              | Namespace\n")
	fmt.Printf(strings.Repeat("-", 80) + "\n")

	for i, pod := range pods {
		fmt.Printf("  %2d   | %-36s | %s\n", i+1, pod.podName, pod.namespace)
	}
	fmt.Printf(strings.Repeat("-", 80) + "\n")

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("📝 Please enter the number (1-%d) of the pod to collect logs from: ", len(pods))

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// Parse the number
	var selection int
	_, err := fmt.Sscanf(input, "%d", &selection)
	if err != nil || selection < 1 || selection > len(pods) {
		fmt.Printf("❌ Invalid selection. Please enter a number between 1 and %d\n", len(pods))
		return nil
	}

	fmt.Printf("✅ Selected pod: %s (namespace: %s)\n", pods[selection-1].podName, pods[selection-1].namespace)
	return &pods[selection-1]
}

// getYN returns "Yes" or "No" for boolean values
func getYN(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// printUsage displays detailed usage information
func printUsage() {
	fmt.Printf(`
📋 Kubernetes 日志收集工具 - 使用说明
================================================================================

功能描述：
  自动收集 Kubernetes 集群中指定服务的日志并输出到 JSON 文件。
  不带任何参数运行时，默认收集 agent-app 和 agent-executor 的日志。

使用方法：
  fetch_log [选项]

选项说明：
  --svc_list string
      指定要收集日志的服务名称列表，多个服务用逗号分隔
      示例: --svc_list agent-factory,agent-memory
      示例: --svc_list agent-executor
      示例: --svc_list agent-app,agent-factory,agent-executor

  --list
      列出集群中所有 Pod 的名称和命名空间，不收集日志
      示例: fetch_log --list

  --preview
      在控制台中预览收集到的日志内容
      示例: fetch_log --svc_list agent-executor --preview
      注意: 预览模式最多显示 100 行日志，requests.log 全部显示（固定 10 行）

  -c string
      指定容器名称（当 Pod 有多个容器时必须指定）
      示例: fetch_log --svc_list efast -c efast-private
      注意: 如果不指定，且 Pod 包含多个容器，工具会提示可用的容器列表

  --logline int
      获取日志的行数，范围 1-300，默认 50 行
      示例: fetch_log --svc_list agent-factory --logline 200
      示例: fetch_log --logline 100
      注意: agent-executor 的 requests.log 固定为 10 行，不受此参数控制

  --enable_debug
      启用 agent-executor 调试模式（会修改 ConfigMap 并重启服务）
      这将会修改以下配置项：
        - app.debug: true
        - app.log_level: "debug"
        - dialog_logging.enable_dialog_logging: true
        - dialog_logging.use_single_log_file: true
      示例: fetch_log --enable_debug
      注意: 需要用户输入 'yes' 确认才会执行

  --disable_debug
      禁用 agent-executor 调试模式（会修改 ConfigMap 并重启服务）
      这将会恢复以下配置项：
        - app.debug: false
        - app.log_level: "info"
        - dialog_logging.enable_dialog_logging: false
        - dialog_logging.use_single_log_file: false
      示例: fetch_log --disable_debug
      注意: 需要用户输入 'yes' 确认才会执行

  --ai
      启用 AI 分析功能，使用大模型分析收集到的日志
      首次使用时会提示配置 ADP 平台的 IP、Token 和模型名称
      配置会保存到 ~/.fetch_log/ 目录，下次使用自动加载
      示例: fetch_log --ai
      示例: fetch_log --svc_list agent-executor --ai

  --token string
      更新 AI 配置中的 Token（当 Token 失效时使用）
      示例: fetch_log --ai --token "new_token_here"
      注意: 必须与 --ai 参数一起使用

  --model string
      更新 AI 配置中的模型名称
      示例: fetch_log --ai --model "deepseek-v3-1-20250821"
      注意: 必须与 --ai 参数一起使用

  --ip string
      更新 AI 配置中的 ADP 平台 IP 地址
      示例: fetch_log --ai --ip "192.168.232.11"
      注意: 必须与 --ai 参数一起使用

默认行为：
  不带任何参数运行时，等同于：
    fetch_log --svc_list agent-app,agent-executor

常用示例：
  # 收集默认服务（agent-app 和 agent-executor）的最新 50 行日志
  fetch_log

  # 收集指定服务的日志
  fetch_log --svc_list agent-factory,agent-memory

  # 收集日志并预览内容
  fetch_log --svc_list agent-executor --preview

  # 收集更多行数的日志
  fetch_log --svc_list agent-factory --logline 200

  # 收集多容器 Pod 的日志（指定容器名）
  fetch_log --svc_list efast -c efast-private

  # 列出所有 Pod
  fetch_log --list

  # 启用调试模式
  fetch_log --enable_debug

  # 使用 AI 分析日志（首次使用会提示配置）
  fetch_log --ai

  # 收集指定服务并使用 AI 分析
  fetch_log --svc_list agent-executor --ai

  # 更新 Token 并使用 AI 分析（当 Token 失效时）
  fetch_log --ai --token "your_new_token"

  # 更新模型名称并使用 AI 分析
  fetch_log --ai --model "new_model_name"

  # 更新所有配置并使用 AI 分析
  fetch_log --ai --ip "192.168.232.11" --token "new_token" --model "new_model"

输出文件：
  日志将保存到当前目录，文件名格式为：log_YYYYMMDD_HHMMSS.json
  示例: log_20260109_153045.json

日志格式说明：
  对于普通服务：log_detail 包含 1 个日志对象
  对于 agent-executor：log_detail 包含 3 个日志对象
    - agent-executor.log: 控制台日志
    - dolphin.log: Dolphin 引擎日志
    - requests.log: 请求日志（固定 10 行）

注意事项：
  1. 需要 kubectl 配置好集群访问权限
  2. --enable_debug 和 --disable_debug 会重启 agent-executor 服务
  3. agent-executor 的 requests.log 始终只获取最新 10 行
  4. --preview 预览模式下，普通日志最多显示 100 行，requests.log 全部显示
  5. 当服务有多个 Pod 时，工具会列出所有匹配的 Pod 并提示选择其中一个
  6. --ai 功能首次使用时会提示配置 ADP 平台信息，配置保存在 ~/.fetch_log/ 目录
  7. AI Token 有时效性，失效时请使用 --token 参数更新
  8. --ai 功能需要网络连接到 ADP 平台

AI 分析功能说明：
  AI 分析功能使用内置的专业分析模板，包含：
  - 系统提示词：专业的日志分析专家角色设定
  - Dolphin 语言文档：完整的 Dolphin 语法规范(28KB)
  - 错误知识库：智能体常见错误和解决方案库
  - 日志数据：实际收集到的服务日志（JSON 格式）

  所有资源文件已内置在工具中，无需额外配置

  日志记录：每次 AI 请求的 URL 和请求体会记录到 ~/.fetch_log/log/ai_request.json

`)
}

// getAIConfigPath returns the full path to the AI configuration file
func getAIConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, aiConfigDir)
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	return filepath.Join(configDir, aiConfigFile), nil
}

// loadAIConfig loads the AI configuration from file
func loadAIConfig() (*AIConfig, error) {
	configPath, err := getAIConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // Config doesn't exist yet
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// saveAIConfig saves the AI configuration to file
func saveAIConfig(config *AIConfig) error {
	configPath, err := getAIConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// readLineWithEdit reads a line from stdin and allows editing
func readLineWithEdit(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	for {
		if defaultValue != "" {
			fmt.Printf("%s [默认: %s]: ", prompt, defaultValue)
		} else {
			fmt.Printf("%s: ", prompt)
		}

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// If user just pressed Enter, use default value
		if input == "" && defaultValue != "" {
			return defaultValue
		}

		// Allow user to edit the input
		fmt.Printf("您输入的是: %s\n", input)
		fmt.Printf("确认请按 Enter，重新输入请按 r: ")

		confirm, _ := reader.ReadString('\n')
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm == "" || confirm == "y" || confirm == "yes" {
			return input
		} else if confirm == "r" {
			continue // Retry input
		}
	}
}

// getAIConfigInteractive interactively gets AI configuration from user
func getAIConfigInteractive() (*AIConfig, error) {
	fmt.Println("\n🤖 AI 分析配置向导")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("首次使用 AI 分析功能，需要配置以下信息:\n")

	ip := readLineWithEdit("请输入 ADP 平台的访问 IP 地址", "")
	token := readLineWithEdit("请输入 ADP 平台用户的 Token", "")
	model := readLineWithEdit("请输入 ADP 平台已添加的可用模型名称", "")

	config := &AIConfig{
		IP:    ip,
		Token: token,
		Model: model,
	}

	// Save configuration
	if err := saveAIConfig(config); err != nil {
		return nil, fmt.Errorf("failed to save AI configuration: %v", err)
	}

	fmt.Println("\n✅ AI 配置已保存到本地")
	return config, nil
}

// updateAIConfig updates specific fields in AI configuration
func updateAIConfig(config *AIConfig, token, model, ip string) error {
	updated := false

	if token != "" {
		config.Token = token
		updated = true
		fmt.Println("✅ Token 已更新")
	}

	if model != "" {
		config.Model = model
		updated = true
		fmt.Println("✅ Model 已更新")
	}

	if ip != "" {
		config.IP = ip
		updated = true
		fmt.Println("✅ IP 已更新")
	}

	if updated {
		if err := saveAIConfig(config); err != nil {
			return fmt.Errorf("failed to save updated configuration: %v", err)
		}
		fmt.Println("✅ 配置已更新")
	}

	return nil
}

// performAIAnalysis performs AI analysis on collected logs
func performAIAnalysis(result LogResult) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🤖 AI 日志分析")
	fmt.Println(strings.Repeat("=", 80))

	// Load or create AI configuration
	config, err := loadAIConfig()
	if err != nil {
		fmt.Printf("❌ 加载 AI 配置失败: %v\n", err)
		fmt.Println("请重新运行命令并配置 AI 信息")
		return
	}

	if config == nil {
		// First time setup
		config, err = getAIConfigInteractive()
		if err != nil {
			fmt.Printf("❌ 配置 AI 失败: %v\n", err)
			return
		}
	}

	// Update configuration if command line parameters are provided
	if *aiToken != "" || *aiModel != "" || *aiIP != "" {
		if err := updateAIConfig(config, *aiToken, *aiModel, *aiIP); err != nil {
			fmt.Printf("❌ 更新配置失败: %v\n", err)
			return
		}
	}

	// Validate configuration
	if config.IP == "" || config.Token == "" || config.Model == "" {
		fmt.Println("❌ AI 配置不完整，请使用 --token, --model, --ip 参数更新配置")
		return
	}

	// Prepare log content for analysis
	logContent := prepareLogContent(result)

	// Chat with AI
	fmt.Println("\n📝 正在将日志发送给 AI 进行分析...\n")
	if err := chatWithAI(config, logContent); err != nil {
		fmt.Printf("❌ AI 分析失败: %v\n", err)
		if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "token") {
			fmt.Println("💡 提示: Token 可能已失效，请使用 --token 参数更新 Token")
		}
		return
	}
}

// prepareLogContent prepares log content for AI analysis using embedded template
func prepareLogContent(result LogResult) string {
	// Convert log result to JSON
	logJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logJSON = []byte("无法序列化日志数据")
	}

	// Replace placeholders in embedded template
	template := string(logAnalysisTemplate)
	template = strings.ReplaceAll(template, "to_fill_error_log", string(logJSON))
	template = strings.ReplaceAll(template, "to_fill_dolphin_doc", string(dolphinDoc))
	template = strings.ReplaceAll(template, "to_fill_knowledge_items", string(knowledgeBase))

	return template
}

// logAIRequest logs AI request information to file
func logAIRequest(url string, jsonData []byte) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create log directory
	logDir := filepath.Join(homeDir, aiConfigDir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Prepare log entry
	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"url":       url,
		"request":   json.RawMessage(jsonData),
	}

	// Marshal to JSON with indentation
	logData, err := json.MarshalIndent(logEntry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %v", err)
	}

	// Write to file (overwrite each time)
	logFile := filepath.Join(logDir, "ai_request.json")
	if err := os.WriteFile(logFile, logData, 0644); err != nil {
		return fmt.Errorf("failed to write log file: %v", err)
	}

	return nil
}

// chatWithAI sends log content to AI and displays the response
func chatWithAI(config *AIConfig, logContent string) error {
	// Prepare request payload
	requestBody := map[string]interface{}{
		"model": config.Model,
		"temperature": float64(1),
		"top_p": float64(1),
		"max_tokens": int(10000),
		"top_k": int(1),
		"presence_penalty": float64(0),
		"frequency_penalty": float64(0),
		"stream": true,
		"messages": []map[string]string{
			{
				"role": "user",
				"content": logContent,
			},
		},
		"response_format": map[string]string{
			"type": "text",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("http://%s/api/mf-model-api/v1/chat/completions", config.IP)

	// Log request information
	if err := logAIRequest(url, jsonData); err != nil {
		// Log failure should not prevent the request
		fmt.Printf("⚠️  Warning: Failed to log request: %v\n", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Read streaming response
	fmt.Println("🤖 AI 分析结果：")
	fmt.Println(strings.Repeat("-", 80))

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read response: %v", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse SSE format: "data: {...}"
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// Check for [DONE] marker
			if data == "[DONE]" {
				break
			}

			// Parse JSON
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				// Skip invalid JSON
				continue
			}

			// Extract content from chunk
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							fmt.Print(content)
						}
					}
				}
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("\n✅ AI 分析完成")

	return nil
}
