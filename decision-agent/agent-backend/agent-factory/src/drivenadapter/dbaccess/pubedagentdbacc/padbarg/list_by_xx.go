package padbarg

type PublishedToWhereCondition struct {
	IsToCustomSpace bool `json:"is_to_custom_space"`
	IsToSquare      bool `json:"is_to_square"`
}

type GetPaPoListByXxArg struct {
	AgentKeys      []string
	AgentIDs       []string
	PubToWhereCond *PublishedToWhereCondition
}

func NewGetPAPoListByIDArg(agentIDs []string, whereCondition *PublishedToWhereCondition) *GetPaPoListByXxArg {
	return &GetPaPoListByXxArg{
		AgentIDs:       agentIDs,
		PubToWhereCond: whereCondition,
	}
}

func NewGetPaPoListByKeyArg(agentKeys []string, whereCondition *PublishedToWhereCondition) *GetPaPoListByXxArg {
	return &GetPaPoListByXxArg{
		AgentKeys:      agentKeys,
		PubToWhereCond: whereCondition,
	}
}
