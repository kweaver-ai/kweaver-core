package daresvo

type RelatedQuestion struct {
	Query string `json:"query"`
}

func (r *DataAgentRes) RelatedQueries() (relatedQuestions []*RelatedQuestion) {
	_relatedQuestions := r.Answer.GetFieldSliceStr("related_questions")

	for _, _relatedQuestion := range _relatedQuestions {
		relatedQuestions = append(relatedQuestions, &RelatedQuestion{
			Query: _relatedQuestion,
		})
	}

	return
}
