package converter

import (
	"fmt"
	"strconv"
)

// AnyToFloat64 for gauge
func AnyToFloat64(value any) (float64, error) {
	var varFloat64 float64
	var err error

	switch ty := value.(type) {

	case int:
		return float64(ty), nil

	case uint32:
		varFloat64, err = strconv.ParseFloat(strconv.Itoa(int(ty)), 64)
		if err != nil {
			return varFloat64, fmt.Errorf("can not convert value '%v' to float64: %v", value, err)
		}
		return varFloat64, nil

	case uint64:
		varFloat64, err = strconv.ParseFloat(strconv.FormatUint(ty, 10), 64)
		if err != nil {
			return varFloat64, fmt.Errorf("can not convert value '%v' to float64: %v", value, err)
		}
		return varFloat64, nil

	case float64:
		return value.(float64), nil

	case string:
		varFloat64, err = strconv.ParseFloat(value.(string), 64)
		if err != nil {
			return varFloat64, fmt.Errorf("can not convert value '%v' to float64: %v", value, err)
		}
		return varFloat64, nil

	default:
		return varFloat64, fmt.Errorf("can not convert value '%v' to float64: %v", value, err)
	}
}

// AnyToInt64 for counter
func AnyToInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case uint64:
		// Handle uint64 separately to avoid potential overflow
		if v <= uint64(int64(^uint64(0)>>1)) {
			return int64(v), nil
		} else {
			return 0, fmt.Errorf("value '%v' is too large to convert to varInt64", value)
		}
	case string:
		varInt64, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("can not convert value '%v' to int64: %v", v, err)
		}
		return varInt64, nil
	default:
		return 0, fmt.Errorf("can not convert value '%v' to int64: unsupported type", v)
	}
}
