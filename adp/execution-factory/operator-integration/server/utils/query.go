package utils

import "context"

func BatchQueryWithContext[T any, U any](ctx context.Context, items []U, batchSize int, queryFunc func(context.Context, []U) ([]T, error)) ([]T, error) {
	var result []T
	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[i:end]
		items, err := queryFunc(ctx, batch)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}
