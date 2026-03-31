package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
)

// Pypi源解析器

const (
	DefaultPypiRepo = "https://pypi.org/simple" // 默认Pypi源
)

// ParsePypiReq 解析Pypi源请求参数
type ParsePypiReq struct {
	PypiRepoURL   string `form:"pypi_repo_url" default:"https://pypi.org/simple" validate:"required,url"`
	PackageName   string `uri:"package_name" validate:"required"`
	PythonVersion string `form:"python_version" default:"3.10"`
}

// ParsePypiResp 解析Pypi源响应参数
type ParsePypiResp struct {
	PackageName string   `json:"package_name"`
	Versions    []string `json:"versions"`
}

// PypiResponse Pypi源响应参数
type PypiResponse struct {
	Info struct {
		Name           string `json:"name"`
		Version        string `json:"version"`
		RequiresPython string `json:"requires_python"`
	} `json:"info"`
	Releases map[string][]PypiRelease `json:"releases"`
}

// PypiRelease Pypi源响应参数
type PypiRelease struct {
	RequiresPython string `json:"requires_python"`
	Yanked         bool   `json:"yanked"`
	YankedReason   string `json:"yanked_reason"`
}

func ParsePypi(ctx context.Context, req *ParsePypiReq) (resp *ParsePypiResp, err error) {
	packageName := strings.TrimSpace(req.PackageName)
	pythonVersion := strings.TrimSpace(req.PythonVersion)
	if packageName == "" {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, "package_name is empty")
	}
	if pythonVersion == "" {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, "python_version is empty")
	}

	targetPy, err := parsePythonVersion(pythonVersion)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid python_version: %s", err.Error()))
	}

	baseURL := normalizeRepoURL(req.PypiRepoURL)
	url := fmt.Sprintf("%s/pypi/%s/json", baseURL, packageName)

	fmt.Printf("Fetching from Pypi: %s\n", url)

	httpClient := &http.Client{Timeout: 25 * time.Second}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtPypiParserFailed, fmt.Sprintf("create request failed: %s", err.Error()))
	}
	httpReq.Header.Set("User-Agent", "pypi-parser/1.0 (+https://pypi.org)")

	rsp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtPypiParserFailed, fmt.Sprintf("request failed: %s", err.Error()))
	}
	defer rsp.Body.Close()
	if rsp.StatusCode == http.StatusNotFound {
		return &ParsePypiResp{
			PackageName: packageName,
			Versions:    []string{},
		}, nil
	}

	if rsp.StatusCode != http.StatusOK {
		return nil, errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtPypiParserFailed, fmt.Sprintf("HTTP %s", rsp.Status))
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return nil, errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtPypiParserFailed, fmt.Sprintf("read ParsePypiResp failed: %s", err.Error()))
	}

	var pypiData PypiResponse
	if err := json.Unmarshal(body, &pypiData); err != nil {
		return nil, errors.NewHTTPError(ctx, http.StatusInternalServerError, errors.ErrExtPypiParserFailed, map[string]interface{}{
			"error": fmt.Sprintf("decode JSON failed: %s", err.Error()),
			"body":  string(body),
		})
	}
	validVersions := collectCompatibleVersions(pypiData, targetPy)
	sort.Sort(sort.Reverse(semver.Collection(validVersions)))

	versions := make([]string, 0, len(validVersions))
	for _, v := range validVersions {
		versions = append(versions, v.Original())
	}

	name := strings.TrimSpace(pypiData.Info.Name)
	if name == "" {
		name = packageName
	}

	return &ParsePypiResp{
		PackageName: name,
		Versions:    versions,
	}, nil
}

func normalizeRepoURL(repoURL string) string {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		repoURL = DefaultPypiRepo
	}
	repoURL = strings.TrimRight(repoURL, "/")
	if before, ok := strings.CutSuffix(repoURL, "/simple"); ok {
		repoURL = before
	}
	return repoURL
}

func collectCompatibleVersions(pypiData PypiResponse, targetPy pythonVer) []*semver.Version {
	validVersions := make([]*semver.Version, 0, len(pypiData.Releases))
	for verStr, releases := range pypiData.Releases {
		pkgVer, err := semver.NewVersion(verStr)
		if err != nil {
			continue
		}
		if len(releases) == 0 {
			continue
		}

		ok := false
		for _, rel := range releases {
			if rel.Yanked {
				continue
			}
			spec := strings.TrimSpace(rel.RequiresPython)
			if spec == "" {
				spec = strings.TrimSpace(pypiData.Info.RequiresPython)
			}
			if pythonSpecSatisfied(spec, targetPy) {
				ok = true
				break
			}
		}
		if ok {
			validVersions = append(validVersions, pkgVer)
		}
	}
	return validVersions
}

type pythonVer struct {
	major     int
	minor     int
	patch     int
	specified int
}

func parsePythonVersion(s string) (pythonVer, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return pythonVer{}, fmt.Errorf("empty python version")
	}
	parts := strings.Split(s, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return pythonVer{}, fmt.Errorf("expect 'X', 'X.Y' or 'X.Y.Z', got '%s'", s)
	}
	maj, err := parseInt(parts[0])
	if err != nil {
		return pythonVer{}, err
	}
	min := 0
	patch := 0
	if len(parts) >= 2 {
		min, err = parseInt(parts[1])
		if err != nil {
			return pythonVer{}, err
		}
	}
	if len(parts) == 3 {
		patch, err = parseInt(parts[2])
		if err != nil {
			return pythonVer{}, err
		}
	}
	return pythonVer{major: maj, minor: min, patch: patch, specified: len(parts)}, nil
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty number")
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}

func pythonSpecSatisfied(spec string, target pythonVer) bool {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return true
	}
	if i := strings.Index(spec, ";"); i >= 0 {
		spec = strings.TrimSpace(spec[:i])
		if spec == "" {
			return true
		}
	}

	clauses := strings.Split(spec, ",")
	for _, clause := range clauses {
		clause = strings.TrimSpace(clause)
		if clause == "" {
			continue
		}
		if !pythonClauseSatisfied(clause, target) {
			return false
		}
	}
	return true
}

func pythonClauseSatisfied(clause string, target pythonVer) bool {
	op, verStr := splitOperatorAndVersion(clause)
	if op == "" || verStr == "" {
		return false
	}

	if op == "~=" {
		base, wildcard, err := parseSpecVersion(verStr)
		if err != nil || wildcard {
			return false
		}
		upper := compatibleUpperBound(base)
		return comparePython(target, base) >= 0 && comparePython(target, upper) < 0
	}

	specVer, wildcard, err := parseSpecVersion(verStr)
	if err != nil {
		return false
	}

	switch op {
	case "==":
		if wildcard {
			return matchPrefix(target, specVer)
		}
		return comparePython(target, specVer) == 0
	case "!=":
		if wildcard {
			return !matchPrefix(target, specVer)
		}
		return comparePython(target, specVer) != 0
	case ">":
		return comparePython(target, specVer) > 0
	case ">=":
		return comparePython(target, specVer) >= 0
	case "<":
		return comparePython(target, specVer) < 0
	case "<=":
		return comparePython(target, specVer) <= 0
	default:
		return false
	}
}

func splitOperatorAndVersion(clause string) (string, string) {
	clause = strings.TrimSpace(clause)
	ops := []string{">=", "<=", "==", "!=", "~=", ">", "<"}
	for _, op := range ops {
		if strings.HasPrefix(clause, op) {
			return op, strings.TrimSpace(clause[len(op):])
		}
	}
	return "", ""
}

func parseSpecVersion(s string) (pythonVer, bool, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, ".*") {
		base := strings.TrimSuffix(s, ".*")
		v, err := parsePythonVersion(base)
		return v, true, err
	}
	v, err := parsePythonVersion(s)
	return v, false, err
}

func compatibleUpperBound(base pythonVer) pythonVer {
	if base.patch != 0 {
		return pythonVer{major: base.major, minor: base.minor + 1, patch: 0}
	}
	return pythonVer{major: base.major + 1, minor: 0, patch: 0}
}

func matchPrefix(target pythonVer, prefix pythonVer) bool {
	if prefix.specified <= 1 {
		return target.major == prefix.major
	}
	if prefix.specified == 2 {
		return target.major == prefix.major && target.minor == prefix.minor
	}
	return target.major == prefix.major && target.minor == prefix.minor && target.patch == prefix.patch
}

func comparePython(a, b pythonVer) int {
	if a.major != b.major {
		if a.major < b.major {
			return -1
		}
		return 1
	}
	if a.minor != b.minor {
		if a.minor < b.minor {
			return -1
		}
		return 1
	}
	if a.patch != b.patch {
		if a.patch < b.patch {
			return -1
		}
		return 1
	}
	return 0
}
