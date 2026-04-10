package initial

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	"golang.org/x/mod/semver"
)

const ANYDATA_DIR = "resource/anydata"
const AGENT_CONFIG = "resource/anydata/agents.json"

type AgentVersionInfo struct {
	*drivenadapters.AgentInfo
	Version string `json:"version"`
}

type AgentConfig struct {
	Agents []*AgentVersionInfo `json:"agents"`
}

func include(path string) string {
	fullPath := filepath.Join(ANYDATA_DIR, path)

	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		log.Printf("[include] Failed to read file: %s, err: %s\n", fullPath, err.Error())
		return ""
	}

	bytes, _ = json.Marshal([]string{string(bytes)})

	return string(bytes[2 : len(bytes)-2])
}

func loadAgents(model string) ([]*AgentVersionInfo, error) {

	if _, err := os.Stat(AGENT_CONFIG); os.IsNotExist(err) {
		log.Printf("[initAnyData] Agent config file not found: %s\n", AGENT_CONFIG)
		return nil, err
	}

	data, err := os.ReadFile(AGENT_CONFIG)

	if err != nil {
		log.Printf("[initAnyData] Failed to read agent config file: %s, err: %s\n", AGENT_CONFIG, err.Error())
		return nil, err
	}

	tmpl, err := template.New("agent_config").Delims("<%=", "%>").Funcs(template.FuncMap{
		"include": include,
	}).Parse(string(data))

	if err != nil {
		log.Printf("[initAnyData] Failed to parse AnyData config file: %v\n", err)
		return nil, err
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, map[string]string{"model": model})

	var agentConfig AgentConfig

	err = json.Unmarshal(buf.Bytes(), &agentConfig)
	if err != nil {
		log.Printf("[initAnyData] Failed to unmarshal agent config: %v\n", err)
		return nil, err
	}

	return agentConfig.Agents, nil
}

func initAgent(ctx context.Context, agentInfo *AgentVersionInfo) {
	var err error
	ad := drivenadapters.NewAnyData()
	db := rds.GetAgentDao()
	agent, dbErr := db.GetAgentByName(ctx, agentInfo.Name)
	if dbErr != nil {
		log.Printf("[initAnyData] Failed to GetAgentByName: %s, err: %s\n", agentInfo.Name, dbErr.Error())
		return
	}

	if agent != nil && agent.ID != 0 {
		_, err = ad.GetAgentByID(ctx, agent.AgentID)

		if err != nil {
			log.Printf("[initAnyData] Failed to GetAgentByID: %s, err: %s\n", agent.AgentID, err.Error())
			dbErr = db.DeleteAgentByName(ctx, agentInfo.Name)

			if dbErr != nil {
				log.Printf("[initAnyData] Failed to DeleteAgentByName: %s, err: %s\n", agentInfo.Name, dbErr.Error())
				return
			}

			agent = nil
		}
	}

	if agent == nil || agent.ID == 0 {
		id, _ := utils.GetUniqueID()
		agent = &rds.AgentModel{
			ID:      id,
			Name:    agentInfo.Name,
			Version: agentInfo.Version,
		}

		agentInfo.Name = fmt.Sprintf("%s_%s_%d", agentInfo.Name, agentInfo.Version, agent.ID)
		agentInfo.Name = strings.ReplaceAll(agentInfo.Name, ".", "_")
		agent.AgentID, err = ad.AddAgent(ctx, agentInfo.AgentInfo)

		if err != nil {
			log.Printf("[initAnyData] Failed to AddAgent: %s, err: %s\n", agentInfo.Name, err.Error())
			return
		}

		dbErr = db.CreateAgent(ctx, agent)

		if dbErr != nil {
			log.Printf("[initAnyData] Failed to CreateAgent: %s, err: %s\n", agentInfo.Name, dbErr.Error())
			return
		}

		log.Printf("[initAnyData] Agent created: %s, ID: %d\n", agent.Name, agent.ID)

		return
	}

	if semver.Compare(agentInfo.Version, agent.Version) == 1 {
		agent.Version = agentInfo.Version
		agentInfo.Name = fmt.Sprintf("%s_%s_%d", agentInfo.Name, agentInfo.Version, agent.ID)
		agentInfo.AgentID = agent.AgentID
		err = ad.UpdateAgent(ctx, agentInfo.AgentInfo)
		if err != nil {
			log.Printf("[initAnyData] Failed to UpdateAgent: %s, err: %s\n", agentInfo.Name, err.Error())
			return
		}

		dbErr = db.UpdateAgent(ctx, agent)

		if dbErr != nil {
			log.Printf("[initAnyData] Failed to UpdateAgent: %s, err: %s\n", agentInfo.Name, dbErr.Error())
		}

		log.Printf("[initAnyData] Agent: %s updated to version: %s\n", agentInfo.Name, agentInfo.Version)

		return
	}

	log.Printf("[initAnyData] Agent: %s is already up to date\n", agentInfo.Name)
}

func initAnyData(ctx context.Context, config *common.Config) {

	var err error

	if config.AnyData.Host == "" || config.AnyData.Model == "" {
		return
	}

	ad := drivenadapters.NewAnyData()
	llmSources, err1 := ad.GetLLMSource(ctx, nil)

	found := false

	if err1 == nil {
		for _, llm := range llmSources.Res.Data {
			found = llm.ModelName == config.AnyData.Model
			if found {
				break
			}
		}
	}

	if !found {
		models, err2 := ad.GetModelList(ctx, nil)
		if err2 != nil {
			log.Printf("[initAnyData] Failed to get model list: %v\n", err)
			return
		}

		for _, model := range models.Res {
			found = model.ModelName == config.AnyData.Model
			if found {
				break
			}
		}
	}

	if !found {
		log.Printf("[initAnyData] Model %v does not exist\n", config.AnyData.Model)
		return
	}

	agents, err := loadAgents(config.AnyData.Model)

	if err != nil {
		log.Printf("[initAnyData] Failed to load agents: %v\n", err)
		return
	}

	if len(agents) > 0 {
		for _, agent := range agents {
			initAgent(ctx, agent)
		}
	}
}
