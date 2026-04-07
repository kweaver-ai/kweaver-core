package business_domain

import (
	"context"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

type noopBusinessDomainService struct{}

func (n *noopBusinessDomainService) ValidateBusinessDomain(ctx context.Context) (err error) {
	return nil
}

func (n *noopBusinessDomainService) AssociateResource(ctx context.Context, bdID, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopBusinessDomainService) DisassociateResource(ctx context.Context, bdID, resourceID string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopBusinessDomainService) BatchDisassociateResource(ctx context.Context, bdID string, resourceIds []string, resourceType interfaces.AuthResourceType) error {
	return nil
}

func (n *noopBusinessDomainService) ResourceList(ctx context.Context, bdID string, resourceType interfaces.AuthResourceType) ([]string, error) {
	return []string{interfaces.ResourceIDAll}, nil
}

func (n *noopBusinessDomainService) BatchResourceList(ctx context.Context, bdIds []string, resourceType interfaces.AuthResourceType) (map[string]string, error) {
	return map[string]string{interfaces.ResourceIDAll: ""}, nil
}
