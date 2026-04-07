// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

//go:generate mockgen -source=interface.go -destination=../mocks/interface.go -package=mocks
import "context"

// App Application interface
type App interface {
	Start() error
	Stop(context.Context)
}

type ResourceDeployType string

func (r ResourceDeployType) String() string {
	return string(r)
}

// IKnQuerySubgraphService Subgraph query service interface
type IKnQuerySubgraphService interface {
	// QueryInstanceSubgraph Query object subgraph
	QueryInstanceSubgraph(ctx context.Context, req *QueryInstanceSubgraphReq) (resp *QueryInstanceSubgraphResp, err error)
}

// IKnSearchService kn_search service interface
type IKnSearchService interface {
	// KnSearch Knowledge network retrieval
	KnSearch(ctx context.Context, req *KnSearchReq) (resp *KnSearchResp, err error)
}
