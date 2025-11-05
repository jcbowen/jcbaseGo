package helper

import (
	"strings"
	"testing"
)

// TestParseUnitString 测试字符串到单位的解析功能
func TestParseUnitString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		hasError bool
	}{
		{"10MB to bytes", "10MB", 10485760, false},
		{"5.5s to nanoseconds", "5.5s", 5500000000, false},
		{"1KB to bytes", "1KB", 1024, false},
		{"2GB to bytes", "2GB", 2147483648, false},
		{"1.5h to nanoseconds", "1.5h", 5400000000000, false},
		{"invalid format", "10invalid", 0, true},
		{"empty string", "", 0, true},
		{"only number", "10", 0, true},
		{"only unit", "MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 根据输入字符串判断应该使用哪种单位类型
			var unitType UnitType
			if strings.Contains(tt.input, "MB") || strings.Contains(tt.input, "KB") || strings.Contains(tt.input, "GB") || strings.Contains(tt.input, "B") {
				unitType = UnitTypeStorage
			} else if strings.Contains(tt.input, "s") || strings.Contains(tt.input, "h") || strings.Contains(tt.input, "m") {
				unitType = UnitTypeTime
			} else {
				unitType = UnitTypeStorage // 默认使用存储类型
			}

			result, _, err := ParseUnitString(tt.input, unitType)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, got %f for input %s", tt.expected, result, tt.input)
				}
			}
		})
	}
}

// TestFormatUnit 测试新的格式化功能
func TestFormatUnit(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		unitType    UnitType
		precision   int
		toUnit      []string
		expected    string
		shouldError bool
	}{
		// 自动格式化测试（数值输入）
		{"auto_1024B", 1024, UnitTypeStorage, 2, []string{"auto"}, "1.00KB", false},
		{"auto_1048576B", 1048576, UnitTypeStorage, 0, []string{"auto"}, "1MB", false},
		{"auto_1000000000ns", 1000000000, UnitTypeTime, 2, []string{"auto"}, "1.00s", false},
		{"auto_3600000000000ns", 3600000000000, UnitTypeTime, 0, []string{"auto"}, "1h", false},

		// 自动格式化测试（字符串输入）
		{"auto_1024B_str", "1024B", UnitTypeStorage, 2, []string{"auto"}, "1.00KB", false},
		{"auto_1MB_str", "1MB", UnitTypeStorage, 2, []string{"auto"}, "1.00MB", false},
		{"auto_1s_str", "1s", UnitTypeTime, 2, []string{"auto"}, "1.00s", false},

		// 指定单位格式化测试（数值输入）
		{"to_KB_1024B", 1024, UnitTypeStorage, 2, []string{"KB"}, "1.00KB", false},
		{"to_MB_1048576B", 1048576, UnitTypeStorage, 0, []string{"MB"}, "1MB", false},
		{"to_s_1000000000ns", 1000000000, UnitTypeTime, 1, []string{"s"}, "1.0s", false},

		// 指定单位格式化测试（字符串输入）
		{"to_KB_1024B_str", "1024B", UnitTypeStorage, 2, []string{"KB"}, "1.00KB", false},
		{"to_MB_1MB_str", "1MB", UnitTypeStorage, 2, []string{"MB"}, "1.00MB", false},
		{"to_h_3600s_str", "3600s", UnitTypeTime, 1, []string{"h"}, "1.0h", false},

		// 默认自动格式化测试
		{"default_auto_1024B", 1024, UnitTypeStorage, 2, []string{}, "1.00KB", false},

		// 错误处理测试
		{"invalid_value_type", struct{}{}, UnitTypeStorage, 2, []string{"auto"}, "", true},
		{"invalid_unit_string", "100Invalid", UnitTypeStorage, 2, []string{"auto"}, "", true},
		{"unit_type_mismatch", "100B", UnitTypeTime, 2, []string{"auto"}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatUnit(tt.value, tt.unitType, tt.precision, tt.toUnit...)
			if tt.shouldError {
				if err == nil {
					t.Errorf("FormatUnit() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("FormatUnit() unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("FormatUnit() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// TestConvertUnit 测试单位之间的转换功能
func TestConvertUnit(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		fromUnit string
		toUnit   string
		expected float64
		hasError bool
	}{
		{"10MB to KB", 10, "MB", "KB", 10240, false},
		{"1h to minutes", 1, "h", "m", 60, false},
		{"1024KB to MB", 1024, "KB", "MB", 1, false},
		{"8b to byte", 8, "b", "byte", 0, true},                      // 默认情况下应该失败，因为b和byte类型不同
		{"8b to byte with preferred type", 8, "b", "byte", 1, false}, // 使用优先类型参数应该成功
		{"invalid unit", 10, "invalid", "KB", 0, true},
		{"type mismatch", 10, "MB", "s", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result float64
			var err error

			// 根据测试名称决定是否使用优先类型参数
			if tt.name == "8b to byte with preferred type" {
				result, err = ConvertUnit(tt.value, tt.fromUnit, tt.toUnit, UnitTypeData)
			} else {
				// 根据单位类型决定使用哪种优先类型
				var unitType UnitType
				if strings.Contains(tt.fromUnit, "MB") || strings.Contains(tt.fromUnit, "KB") || strings.Contains(tt.fromUnit, "GB") || strings.Contains(tt.fromUnit, "B") {
					unitType = UnitTypeStorage
				} else if strings.Contains(tt.fromUnit, "s") || strings.Contains(tt.fromUnit, "h") || strings.Contains(tt.fromUnit, "m") {
					unitType = UnitTypeTime
				} else {
					unitType = UnitTypeStorage // 默认使用存储类型
				}
				result, err = ConvertUnit(tt.value, tt.fromUnit, tt.toUnit, unitType)
			}

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for conversion %s to %s, but got none", tt.fromUnit, tt.toUnit)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for conversion %s to %s: %v", tt.fromUnit, tt.toUnit, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, got %f for conversion %s to %s", tt.expected, result, tt.fromUnit, tt.toUnit)
				}
			}
		})
	}
}

// TestIsValidUnit 测试单位有效性检查
func TestIsValidUnit(t *testing.T) {
	tests := []struct {
		name     string
		unit     string
		expected bool
	}{
		{"valid MB", "MB", true},
		{"valid s", "s", true},
		{"valid KB", "KB", true},
		{"invalid unit", "invalid", false},
		{"empty", "", false},
		{"valid ms", "ms", true},
		{"valid GB", "GB", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 根据单位字符串判断应该使用哪种单位类型
			var unitType UnitType
			if strings.Contains(tt.unit, "MB") || strings.Contains(tt.unit, "KB") || strings.Contains(tt.unit, "GB") || strings.Contains(tt.unit, "B") {
				unitType = UnitTypeStorage
			} else if strings.Contains(tt.unit, "s") || strings.Contains(tt.unit, "h") || strings.Contains(tt.unit, "m") {
				unitType = UnitTypeTime
			} else {
				unitType = UnitTypeStorage // 默认使用存储类型
			}

			result := IsValidUnit(tt.unit, unitType)
			if result != tt.expected {
				t.Errorf("Expected %t, got %t for unit %s", tt.expected, result, tt.unit)
			}
		})
	}
}

// TestGetUnitType 测试单位类型获取
func TestGetUnitType(t *testing.T) {
	tests := []struct {
		name     string
		unit     string
		expected UnitType
		hasError bool
	}{
		{"storage unit", "MB", UnitTypeStorage, false},
		{"time unit", "s", UnitTypeTime, false},
		{"data unit", "b", UnitTypeData, false},
		{"invalid unit", "invalid", UnitTypeStorage, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 根据单位字符串判断应该使用哪种单位类型
			var unitType UnitType
			upperUnit := strings.ToUpper(tt.unit)
			if strings.Contains(upperUnit, "MB") || strings.Contains(upperUnit, "KB") || strings.Contains(upperUnit, "GB") || (strings.Contains(upperUnit, "B") && tt.unit != "b") {
				unitType = UnitTypeStorage
			} else if strings.Contains(tt.unit, "s") || strings.Contains(tt.unit, "h") || strings.Contains(tt.unit, "m") {
				unitType = UnitTypeTime
			} else if tt.unit == "b" {
				unitType = UnitTypeData
			} else {
				unitType = UnitTypeStorage // 默认使用存储类型
			}

			result, err := GetUnitType(tt.unit, unitType)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for unit %s, but got none", tt.unit)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for unit %s: %v", tt.unit, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %v, got %v for unit %s", tt.expected, result, tt.unit)
				}
			}
		})
	}
}

// TestStrUnitMethods 测试Str结构体的单位转换方法
func TestStrUnitMethods(t *testing.T) {
	t.Run("ToUnitValue", func(t *testing.T) {
		s := NewStr("10MB")
		result, err := s.ToUnitValue(UnitTypeStorage)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result != 10485760 {
			t.Errorf("Expected 10485760, got %f", result)
		}
	})

	t.Run("IsUnitString", func(t *testing.T) {
		s1 := NewStr("10MB")
		if !s1.IsUnitString(UnitTypeStorage) {
			t.Error("Expected true for valid unit string")
		}

		s2 := NewStr("10invalid")
		if s2.IsUnitString(UnitTypeStorage) {
			t.Error("Expected false for invalid unit string")
		}
	})

	t.Run("FormatAsUnit", func(t *testing.T) {
		s := NewStr("")
		s.FormatAsUnit(10485760, UnitTypeStorage, 2, "auto")
		if s.String != "10.00MB" {
			t.Errorf("Expected 10.00MB, got %s", s.String)
		}
	})

}

// TestUnitTypeParameter 测试单位类型参数功能
func TestUnitTypeParameter(t *testing.T) {
	t.Run("ParseUnitString with preferred types", func(t *testing.T) {
		// 测试"byte"单位在不同优先类型下的解析

		// 默认情况下，"byte"应该被解析为存储单位
		_, unitInfo, err := ParseUnitString("8byte", UnitTypeStorage)
		if err != nil {
			t.Errorf("Unexpected error parsing '8byte': %v", err)
		}
		if unitInfo.UnitType != UnitTypeStorage {
			t.Errorf("Expected UnitTypeStorage for 'byte' by default, got %v", unitInfo.UnitType)
		}

		// 指定优先类型为数据单位时，"byte"应该被解析为数据单位
		_, unitInfo, err = ParseUnitString("8byte", UnitTypeData)
		if err != nil {
			t.Errorf("Unexpected error parsing '8byte' with UnitTypeData: %v", err)
		}
		if unitInfo.UnitType != UnitTypeData {
			t.Errorf("Expected UnitTypeData for 'byte' with preferred type, got %v", unitInfo.UnitType)
		}
	})

	t.Run("ConvertUnit with preferred types", func(t *testing.T) {
		// 测试"8b to byte"转换在不同优先类型下的行为

		// 默认情况下应该失败，因为b和byte类型不同
		result, err := ConvertUnit(8, "b", "byte", UnitTypeStorage)
		if err == nil {
			t.Error("Expected error for '8b to byte' conversion by default, but got none")
		}

		// 指定优先类型为数据单位时应该成功
		result, err = ConvertUnit(8, "b", "byte", UnitTypeData)
		if err != nil {
			t.Errorf("Unexpected error for '8b to byte' conversion with UnitTypeData: %v", err)
		}
		if result != 1 {
			t.Errorf("Expected 1, got %f for '8b to byte' conversion", result)
		}
	})

	t.Run("IsValidUnit with preferred types", func(t *testing.T) {
		// 测试单位有效性检查在不同优先类型下的行为

		// "byte"在默认情况下是有效的
		if !IsValidUnit("byte", UnitTypeStorage) {
			t.Error("Expected 'byte' to be valid by default")
		}

		// "byte"在指定数据单位类型时也是有效的
		if !IsValidUnit("byte", UnitTypeData) {
			t.Error("Expected 'byte' to be valid with UnitTypeData")
		}
	})

	t.Run("GetUnitType with preferred types", func(t *testing.T) {
		// 测试获取单位类型在不同优先类型下的行为

		// 默认情况下，"byte"应该是存储单位
		unitType, err := GetUnitType("byte", UnitTypeStorage)
		if err != nil {
			t.Errorf("Unexpected error getting unit type for 'byte': %v", err)
		}
		if unitType != UnitTypeStorage {
			t.Errorf("Expected UnitTypeStorage for 'byte' by default, got %v", unitType)
		}

		// 指定优先类型为数据单位时，"byte"应该是数据单位
		unitType, err = GetUnitType("byte", UnitTypeData)
		if err != nil {
			t.Errorf("Unexpected error getting unit type for 'byte' with UnitTypeData: %v", err)
		}
		if unitType != UnitTypeData {
			t.Errorf("Expected UnitTypeData for 'byte' with preferred type, got %v", unitType)
		}
	})
}

// TestGetAvailableUnits 测试获取可用单位功能
func TestGetAvailableUnits(t *testing.T) {
	tests := []struct {
		name     string
		unitType UnitType
		expected []string
	}{
		{"storage units", UnitTypeStorage, []string{"B", "KB", "MB", "GB", "TB", "PB"}},
		{"time units", UnitTypeTime, []string{"ns", "μs", "ms", "s", "m", "h", "d"}},
		{"data units", UnitTypeData, []string{"b", "byte", "Kb", "Kbyte", "Mb", "Mbyte", "Gb", "Gbyte"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAvailableUnits(tt.unitType)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d units, got %d", len(tt.expected), len(result))
			}

			for i, expectedUnit := range tt.expected {
				if i >= len(result) {
					break
				}
				if result[i] != expectedUnit {
					t.Errorf("Expected unit %s at index %d, got %s", expectedUnit, i, result[i])
				}
			}
		})
	}
}

// TestRoundUnitValue 测试单位值四舍五入功能
func TestRoundUnitValue(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		unit      string
		precision int
		expected  float64
		hasError  bool
	}{
		{"round 10.456 MB to 2 decimals", 10.456, "MB", 2, 10.46, false},
		{"round 5.555 s to 1 decimal", 5.555, "s", 1, 5.6, false},
		{"round 1.234 KB to 0 decimals", 1.234, "KB", 0, 1, false},
		{"invalid unit", 10, "invalid", 2, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 根据单位确定单位类型
			var unitType UnitType
			if strings.Contains(strings.ToUpper(tt.unit), "MB") || strings.Contains(strings.ToUpper(tt.unit), "KB") || strings.Contains(strings.ToUpper(tt.unit), "GB") || strings.Contains(strings.ToUpper(tt.unit), "B") {
				unitType = UnitTypeStorage
			} else if strings.Contains(strings.ToUpper(tt.unit), "S") || strings.Contains(strings.ToUpper(tt.unit), "H") || strings.Contains(strings.ToUpper(tt.unit), "M") {
				unitType = UnitTypeTime
			} else {
				unitType = UnitTypeStorage // 默认使用存储类型
			}

			result, err := RoundUnitValue(tt.value, tt.unit, tt.precision, unitType)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for unit %s, but got none", tt.unit)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for unit %s: %v", tt.unit, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, got %f for value %f", tt.expected, result, tt.value)
				}
			}
		})
	}
}
