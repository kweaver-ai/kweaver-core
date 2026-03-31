package normalize

import "go.mongodb.org/mongo-driver/bson/primitive"

// NormalizeContainer converts only container-like BSON values into plain maps/slices.
// Scalar BSON primitive values are preserved as-is.
func NormalizeContainer(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, value := range v {
			result[key] = NormalizeContainer(value)
		}
		return result
	case primitive.M:
		result := make(map[string]interface{}, len(v))
		for key, value := range v {
			result[key] = NormalizeContainer(value)
		}
		return result
	case primitive.D:
		result := make(map[string]interface{}, len(v))
		for _, elem := range v {
			result[elem.Key] = NormalizeContainer(elem.Value)
		}
		return result
	case primitive.E:
		return map[string]interface{}{
			v.Key: NormalizeContainer(v.Value),
		}
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = NormalizeContainer(item)
		}
		return result
	case primitive.A:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = NormalizeContainer(item)
		}
		return result
	default:
		return data
	}
}

func AsMap(data interface{}) (map[string]interface{}, bool) {
	result, ok := NormalizeContainer(data).(map[string]interface{})
	return result, ok
}

func AsSlice(data interface{}) ([]interface{}, bool) {
	result, ok := NormalizeContainer(data).([]interface{})
	return result, ok
}

func PrimitiveToMap(data interface{}) map[string]interface{} {
	result, ok := AsMap(data)
	if !ok {
		return map[string]interface{}{}
	}

	return result
}
