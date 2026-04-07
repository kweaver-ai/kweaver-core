package drivenadapters

import "context"

type mockBusinessDomain struct{}

func (m *mockBusinessDomain) BindResourceInternal(ctx context.Context, params BizDomainResourceParams) error {
	return nil
}

func (m *mockBusinessDomain) UnBindResourceInternal(ctx context.Context, params BizDomainResourceParams) error {
	return nil
}

func (m *mockBusinessDomain) ListResource(ctx context.Context, params BizDomainResourceQuery, token string) (BizDomainResourceResp, error) {
	return BizDomainResourceResp{
		Total: 0,
		Items: BizDomainResources{},
	}, nil
}

func (m *mockBusinessDomain) CheckerResource(ctx context.Context, params BizDomainResourceParams, token string) (bool, error) {
	return true, nil
}
