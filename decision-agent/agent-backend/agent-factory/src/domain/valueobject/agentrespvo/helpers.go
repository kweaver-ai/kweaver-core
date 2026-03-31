package agentrespvo

import (
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/util"
)

// -------prompt start--------

func IsPromptTypeInterface(obj interface{}) (isValid bool, err error) {
	byt, err := sonic.Marshal(obj)
	if err != nil {
		return
	}

	isValid, err = IsPromptType(string(byt))
	if err != nil {
		return
	}

	return
}

func IsPromptType(jsonStr string) (isValid bool, err error) {
	isValid, err = util.IsJsonschemaValid(getPromptTypeJsonSchemaStr(), jsonStr)
	if err != nil {
		return
	}

	if !isValid {
		return
	}

	return
}

func getPromptTypeJsonSchemaStr() string {
	return `{
  "type": "object",
  "properties": {
    "answer": {
      "type": "string"
    },
    "think": {
      "type": "string"
    }
  },
  "required": [
    "answer",
    "think"
  ]
}
`
}

// -------prompt end--------

// -------explore start--------

func IsExploreTypeInterface(obj interface{}) (isValid bool, err error) {
	byt, err := sonic.Marshal(obj)
	if err != nil {
		return
	}

	isValid, err = IsExploreType(string(byt))
	if err != nil {
		return
	}

	return
}

func IsExploreType(jsonStr string) (isValid bool, err error) {
	isValid, err = util.IsJsonschemaValid(getExploreTypeJsonSchemaStr(), jsonStr)
	if err != nil {
		return
	}

	if !isValid {
		return
	}

	return
}

func getExploreTypeJsonSchemaStr() string {
	return `{
  "type": "array",
  "items": {
    "type": "object",
    "properties": {
      "agent_name": {
        "type": "string"
      },
      "answer": {},
      "think": {
        "type": "string"
      },
      "status": {
        "type": "string"
      },
      "interrupted": {
        "type": "boolean"
      }
    },
    "required": [
      "agent_name",
      "answer",
      "think",
      "status",
      "interrupted"
    ]
  }
}
`
}

// -------explore end--------
