package skill

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	. "github.com/smartystreets/goconvey/convey"
)

func TestParseRegisterReqContentSuccess(t *testing.T) {
	Convey("parseRegisterReq content success", t, func() {
		parser := newSkillParser()
		req := &interfaces.RegisterSkillReq{
			BusinessDomainID: "bd-test",
			UserID:           "user-1",
			FileType:         "content",
			File: json.RawMessage(`---
name: demo-skill
description: demo desc
version: 1.2.3
metadata:
  scene: test
---
Use this skill carefully.`),
			Source: "unit-test",
		}

		skill, files, assets, err := parser.parseRegisterReq(req)
		So(err, ShouldBeNil)
		So(skill.Name, ShouldEqual, "demo-skill")
		So(skill.Version, ShouldNotEqual, "1.2.3")
		_, parseErr := uuid.Parse(skill.Version)
		So(parseErr, ShouldBeNil)
		So(skill.SkillContent, ShouldEqual, "Use this skill carefully.")
		So(len(files), ShouldEqual, 0)
		So(len(assets), ShouldEqual, 0)
	})
}

func TestParseRegisterReqZipMissingSkillMD(t *testing.T) {
	Convey("parseRegisterReq zip missing SKILL.md", t, func() {
		parser := newSkillParser()
		req := &interfaces.RegisterSkillReq{
			BusinessDomainID: "bd-test",
			UserID:           "user-1",
			FileType:         "zip",
			File:             buildZip(t, map[string]string{"refs/guide.md": "hello"}),
		}

		_, _, _, err := parser.parseRegisterReq(req)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "SKILL.md not found")
	})
}

func TestParseRegisterReqZipRejectsTraversalPath(t *testing.T) {
	Convey("parseRegisterReq zip rejects traversal path", t, func() {
		parser := newSkillParser()
		req := &interfaces.RegisterSkillReq{
			BusinessDomainID: "bd-test",
			UserID:           "user-1",
			FileType:         "zip",
			File: buildZip(t, map[string]string{
				"SKILL.md":      validSkillMarkdown(),
				"../secret.txt": "bad",
			}),
		}

		_, _, _, err := parser.parseRegisterReq(req)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "invalid skill file path")
	})
}

func TestParseRegisterReqZipReturnsAssets(t *testing.T) {
	Convey("parseRegisterReq zip returns assets", t, func() {
		parser := newSkillParser()
		req := &interfaces.RegisterSkillReq{
			BusinessDomainID: "bd-test",
			UserID:           "user-1",
			FileType:         "zip",
			File: buildZip(t, map[string]string{
				"SKILL.md":       validSkillMarkdown(),
				"refs/guide.md":  "guide",
				"scripts/run.py": "print('ok')",
			}),
		}

		skill, files, assets, err := parser.parseRegisterReq(req)
		So(err, ShouldBeNil)
		So(skill.Name, ShouldEqual, "demo-skill")
		So(skill.SkillContent, ShouldEqual, "Use this skill carefully.")
		So(len(files), ShouldEqual, 2)
		So(len(assets), ShouldEqual, 2)
	})
}

func TestChecksumSHA256(t *testing.T) {
	Convey("checksumSHA256 returns stable sha256", t, func() {
		sum := checksumSHA256([]byte("demo"))
		So(len(sum), ShouldEqual, 64)
		So(sum, ShouldNotEqual, checksumSHA256([]byte("other")))
	})
}

func buildZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("Create(%s) error = %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("Write(%s) error = %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close error = %v", err)
	}
	return buf.Bytes()
}

func validSkillMarkdown() string {
	return `---
name: demo-skill
description: demo desc
---
Use this skill carefully.`
}
