package skill

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	"gopkg.in/yaml.v3"
)

type skillParser struct{}

type skillFrontmatter struct {
	Name         string                 `yaml:"name" validate:"required"`
	Description  string                 `yaml:"description" validate:"required"`
	Dependencies map[string]interface{} `yaml:"dependencies"`
	Metadata     map[string]interface{} `yaml:"metadata"`
}

// 设置SKILL.md统一命名

const SkillMD = "SKILL.md"

// skillAsset 技能资产
type skillAsset struct {
	RelPath  string
	FileType string
	MimeType string
	Content  []byte
}

func newSkillParser() *skillParser {
	return &skillParser{}
}

func (p *skillParser) parseRegisterReq(req *interfaces.RegisterSkillReq) (skillDB *model.SkillRepositoryDB, files []*interfaces.SkillFileSummary, assets []*skillAsset, err error) {
	switch req.FileType {
	case "content":
		content, err := decodeContent(req.File)
		if err != nil {
			return nil, nil, nil, err
		}
		skill, err := p.parseSkillContent(content, req)
		if err != nil {
			return nil, nil, nil, err
		}
		return skill, nil, nil, nil
	case "zip":
		content, files, assets, err := p.parseSkillZip(req)
		if err != nil {
			return nil, nil, nil, err
		}
		skill, err := p.parseSkillContent(content, req)
		if err != nil {
			return nil, nil, nil, err
		}
		return skill, files, assets, nil
	default:
		return nil, nil, nil, fmt.Errorf("unsupported file type: %s", req.FileType)
	}
}

func (p *skillParser) parseSkillContent(content string, req *interfaces.RegisterSkillReq) (*model.SkillRepositoryDB, error) {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid SKILL.md format: missing frontmatter")
	}

	fm := &skillFrontmatter{}
	if err := yaml.Unmarshal([]byte(parts[1]), fm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal skill frontmatter: %w", err)
	}
	if err := validator.New().Struct(fm); err != nil {
		return nil, fmt.Errorf("invalid skill frontmatter: %w", err)
	}

	skill := &model.SkillRepositoryDB{
		Name:         fm.Name,
		Description:  fm.Description,
		SkillContent: strings.TrimSpace(parts[2]),
		Version:      uuid.New().String(),
		Status:       interfaces.BizStatusUnpublish.String(),
		Source:       req.Source,
		Dependencies: utils.ObjectToJSON(fm.Dependencies),
		ExtendInfo:   utils.ObjectToJSON(fm.Metadata),
		CreateUser:   req.UserID,
		UpdateUser:   req.UserID,
	}
	return skill, nil
}

func (p *skillParser) parseSkillZip(req *interfaces.RegisterSkillReq) (string, []*interfaces.SkillFileSummary, []*skillAsset, error) {
	reader, err := zip.NewReader(bytes.NewReader(req.File), int64(len(req.File)))
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to open zip: %w", err)
	}

	var skillContent string
	files := make([]*interfaces.SkillFileSummary, 0, len(reader.File))
	assets := make([]*skillAsset, 0, len(reader.File))
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		relPath, err := normalizeZipPath(file.Name)
		if err != nil {
			return "", nil, nil, err
		}

		rc, err := file.Open()
		if err != nil {
			return "", nil, nil, err
		}
		content, readErr := io.ReadAll(rc)
		_ = rc.Close()
		if readErr != nil {
			return "", nil, nil, readErr
		}

		if strings.EqualFold(relPath, SkillMD) {
			skillContent = string(content)
			// 如果refpath为SKILL.md，转换为大写
			relPath = SkillMD
		}

		files = append(files, &interfaces.SkillFileSummary{
			RelPath:  relPath,
			FileType: detectFileType(relPath),
			Size:     int64(len(content)),
			MimeType: detectMimeType(relPath),
		})
		assets = append(assets, &skillAsset{
			RelPath:  relPath,
			FileType: detectFileType(relPath),
			MimeType: detectMimeType(relPath),
			Content:  content,
		})
	}

	if skillContent == "" {
		return "", nil, nil, fmt.Errorf("SKILL.md not found in zip")
	}
	return skillContent, files, assets, nil
}

func decodeContent(raw json.RawMessage) (string, error) {
	var content string
	if err := json.Unmarshal(raw, &content); err == nil {
		return content, nil
	}
	return string(raw), nil
}

func normalizeZipPath(path string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(path))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	if clean == "" || clean == "." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return "", fmt.Errorf("invalid skill file path: %s", path)
	}
	return clean, nil
}

func detectFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".py", ".js", ".ts", ".sh":
		return "script"
	case ".md", ".txt":
		return "reference"
	case ".yaml", ".yml", ".json", ".toml":
		return "config"
	default:
		return "asset"
	}
}

func detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md":
		return "text/markdown"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"
	default:
		return "application/octet-stream"
	}
}
