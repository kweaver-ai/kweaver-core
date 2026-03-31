// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/knowledge_network_service.go -destination ../interfaces/mock/mock_knowledge_network_service.go
type KnowledgeNetworkService interface {
	SearchSubgraph(ctx context.Context, query *SubGraphQueryBaseOnSource) (ObjectSubGraph, error)
	SearchSubgraphByTypePath(ctx context.Context, query *SubGraphQueryBaseOnTypePath) (PathsEntries, error)
	SearchSubgraphByObjects(ctx context.Context, query *SubGraphQueryBaseOnObjects) (ObjectSubGraph, error)
}
