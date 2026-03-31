package drivenadapters

import "context"

type mockAuthorization struct{}

func (m *mockAuthorization) CreatePolicy(ctx context.Context, params []PermPolicyParams) error {
	return nil
}

func (m *mockAuthorization) UpdatePolicy(ctx context.Context, plocyIDs []string, params []PermPolicyParams) error {
	return nil
}

func (m *mockAuthorization) DeletePolicy(ctx context.Context, params PermPolicyDeleteParams) error {
	return nil
}

func (m *mockAuthorization) ResourceFilter(ctx context.Context, params ResourceFilterParmas) ([]Resource, error) {
	return []Resource{{
		ID:   "mock-resource-id",
		Type: "mock-resource",
		Name: "mock-resource",
	}}, nil
}

func (m *mockAuthorization) ListResourceOperation(ctx context.Context, params ListResourceOperationParmas) ([]ListResourceOperationRes, error) {
	resources := make([]ListResourceOperationRes, 0, len(params.Resources))
	resources = append(resources, ListResourceOperationRes{
		ID:        "mock-resource-id",
		Operation: []string{"run_with_app", "create", "modify", "delete", "view", "manual_exec", "run_statistics", "list", "display"},
	})
	return resources, nil
}

func (m *mockAuthorization) OperationPermCheck(ctx context.Context, params OperationPermCheckParams) (bool, error) {
	return true, nil
}

func (m *mockAuthorization) ListResource(ctx context.Context, params ListResourceParams) ([]Resource, error) {
	if params.Resource.ID != "" || params.Resource.Type != "" {
		return []Resource{params.Resource}, nil
	}
	return []Resource{{
		ID:   "mock-resource-id",
		Type: "mock-resource",
		Name: "mock-resource",
	}}, nil
}
