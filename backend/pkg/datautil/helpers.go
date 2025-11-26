package datautil

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func ParseFilterJSON(filterJSON string) ([]map[string]interface{}, error) {
	if filterJSON == "" {
		return nil, nil
	}

	var single map[string]interface{}
	if err := json.Unmarshal([]byte(filterJSON), &single); err == nil {
		return []map[string]interface{}{single}, nil
	}

	var multi []map[string]interface{}
	if err := json.Unmarshal([]byte(filterJSON), &multi); err == nil {
		return multi, nil
	}

	return nil, fmt.Errorf("invalid filter json: %s", filterJSON)
}

func ParseFilterKey(key string) (field string, operator string) {
	parts := strings.Split(key, "__")
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.ToLower(parts[1])
}

func ToInt32(value interface{}) (int32, error) {
	switch v := value.(type) {
	case float64:
		return int32(v), nil
	case float32:
		return int32(v), nil
	case int:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case json.Number:
		i64, err := v.Int64()
		return int32(i64), err
	case string:
		if v == "" {
			return 0, nil
		}
		i64, err := strconv.ParseInt(v, 10, 32)
		return int32(i64), err
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", value)
	}
}

func ToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case float64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case int:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case json.Number:
		return v.Int64()
	case string:
		if v == "" {
			return 0, nil
		}
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", value)
	}
}

func ToInt32Slice(value interface{}) ([]int32, error) {
	switch v := value.(type) {
	case []interface{}:
		result := make([]int32, 0, len(v))
		for _, item := range v {
			i, err := ToInt32(item)
			if err != nil {
				return nil, err
			}
			result = append(result, i)
		}
		return result, nil
	case []int:
		result := make([]int32, len(v))
		for i, item := range v {
			result[i] = int32(item)
		}
		return result, nil
	case []int32:
		return v, nil
	case []int64:
		result := make([]int32, len(v))
		for i, item := range v {
			result[i] = int32(item)
		}
		return result, nil
	case string:
		if v == "" {
			return []int32{}, nil
		}
		var raw []interface{}
		if err := json.Unmarshal([]byte(v), &raw); err != nil {
			return nil, err
		}
		return ToInt32Slice(raw)
	default:
		return nil, fmt.Errorf("unsupported slice type %T", value)
	}
}

func EncodeUint32Slice(values []uint32) (*string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	bytes, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	result := string(bytes)
	return &result, nil
}

func DecodeUint32Slice(raw string) ([]uint32, error) {
	if raw == "" {
		return []uint32{}, nil
	}
	var values []uint32
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func CloneString(val string) *string {
	v := val
	return &v
}

func CloneUint32(val uint32) *uint32 {
	v := val
	return &v
}
