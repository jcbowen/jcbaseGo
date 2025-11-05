package helper

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// UnitType 定义单位类型
// UnitType defines the type of unit
type UnitType int

const (
	// UnitTypeStorage 存储单位类型
	// UnitTypeStorage storage unit type
	UnitTypeStorage UnitType = iota
	// UnitTypeTime 时间单位类型
	// UnitTypeTime time unit type
	UnitTypeTime
	// UnitTypeData 数据单位类型（位/字节）
	// UnitTypeData data unit type (bit/byte)
	UnitTypeData
)

// Unit 定义单位结构
// Unit defines the unit structure
type Unit struct {
	Name     string   // 单位名称
	Symbols  []string // 单位符号（支持多种表示方式）
	Factor   float64  // 转换因子（相对于基础单位）
	UnitType UnitType // 单位类型
}

// 预定义的单位映射
// Predefined unit mappings
var (
	// StorageUnits 存储单位映射
	// StorageUnits storage unit mapping
	StorageUnits = []Unit{
		{"Byte", []string{"B", "byte", "bytes"}, 1, UnitTypeStorage},
		{"Kilobyte", []string{"KB", "K", "kilobyte", "kilobytes"}, 1024, UnitTypeStorage},
		{"Megabyte", []string{"MB", "M", "megabyte", "megabytes"}, 1024 * 1024, UnitTypeStorage},
		{"Gigabyte", []string{"GB", "G", "gigabyte", "gigabytes"}, 1024 * 1024 * 1024, UnitTypeStorage},
		{"Terabyte", []string{"TB", "T", "terabyte", "terabytes"}, 1024 * 1024 * 1024 * 1024, UnitTypeStorage},
		{"Petabyte", []string{"PB", "P", "petabyte", "petabytes"}, 1024 * 1024 * 1024 * 1024 * 1024, UnitTypeStorage},
	}

	// TimeUnits 时间单位映射
	// TimeUnits time unit mapping
	TimeUnits = []Unit{
		{"Nanosecond", []string{"ns", "nanosecond", "nanoseconds"}, 1, UnitTypeTime},
		{"Microsecond", []string{"μs", "us", "microsecond", "microseconds"}, 1000, UnitTypeTime},
		{"Millisecond", []string{"ms", "millisecond", "milliseconds"}, 1000 * 1000, UnitTypeTime},
		{"Second", []string{"s", "sec", "second", "seconds"}, 1000 * 1000 * 1000, UnitTypeTime},
		{"Minute", []string{"m", "min", "minute", "minutes"}, 60 * 1000 * 1000 * 1000, UnitTypeTime},
		{"Hour", []string{"h", "hr", "hour", "hours"}, 60 * 60 * 1000 * 1000 * 1000, UnitTypeTime},
		{"Day", []string{"d", "day", "days"}, 24 * 60 * 60 * 1000 * 1000 * 1000, UnitTypeTime},
	}

	// DataUnits 数据单位映射
	// DataUnits data unit mapping
	DataUnits = []Unit{
		{"Bit", []string{"b", "bit", "bits"}, 1, UnitTypeData},
		{"Byte", []string{"byte", "bytes"}, 8, UnitTypeData}, // 保留"byte"符号用于数据单位
		{"Kilobit", []string{"Kb", "Kbit", "kilobit", "kilobits"}, 1024, UnitTypeData},
		{"Kilobyte", []string{"Kbyte", "kilobyte", "kilobytes"}, 1024 * 8, UnitTypeData},
		{"Megabit", []string{"Mb", "Mbit", "megabit", "megabits"}, 1024 * 1024, UnitTypeData},
		{"Megabyte", []string{"Mbyte", "megabyte", "megabytes"}, 1024 * 1024 * 8, UnitTypeData},
		{"Gigabit", []string{"Gb", "Gbit", "gigabit", "gigabits"}, 1024 * 1024 * 1024, UnitTypeData},
		{"Gigabyte", []string{"Gbyte", "gigabyte", "gigabytes"}, 1024 * 1024 * 1024 * 8, UnitTypeData},
	}

	// AllUnits 所有单位的合并映射
	// AllUnits combined mapping of all units
	AllUnits = append(append(append([]Unit{}, StorageUnits...), TimeUnits...), DataUnits...)
)

// ParseUnitString 解析包含单位的字符串，返回数值和单位信息
// ParseUnitString parses a string containing units and returns the numeric value and unit information
//
// 参数:
//   - str: 包含单位的字符串，如 "10MB", "5.5s", "1.2KB"
//   - unitType: 单位类型
//
// 返回值:
//   - float64: 转换后的数值（以基础单位表示）
//   - *Unit: 解析出的单位信息
//   - error: 解析过程中的错误
//
// 示例:
//   - ParseUnitString("10MB", UnitTypeStorage) -> 10485760, Unit{Name: "Megabyte", ...}, nil
//   - ParseUnitString("5.5s", UnitTypeTime) -> 5500000000, Unit{Name: "Second", ...}, nil
//   - ParseUnitString("8byte", UnitTypeData) -> 64, Unit{Name: "Byte", ...}, nil
func ParseUnitString(str string, unitType UnitType) (float64, *Unit, error) {
	// 使用正则表达式匹配数值和单位
	// Use regex to match numeric value and unit
	re := regexp.MustCompile(`^\s*([-+]?\d*\.?\d+)\s*([a-zA-Z]+)\s*$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(str))

	if len(matches) != 3 {
		return 0, nil, fmt.Errorf("invalid unit string format: %s", str)
	}

	// 解析数值部分
	// Parse numeric value
	valueStr := matches[1]
	unitStr := strings.ToUpper(matches[2])

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, nil, fmt.Errorf("invalid numeric value: %s", valueStr)
	}

	// 查找匹配的单位
	// Find matching unit
	var matchedUnit *Unit
	var exactMatches []*Unit

	// 首先查找精确匹配（大小写敏感）
	// First look for exact matches (case sensitive)
	for i := range AllUnits {
		unit := &AllUnits[i]
		for _, symbol := range unit.Symbols {
			if symbol == matches[2] { // 使用原始大小写进行精确匹配
				exactMatches = append(exactMatches, unit)
				break
			}
		}
	}

	// 如果有精确匹配，优先使用指定类型的匹配
	// If there are exact matches, prioritize specified type
	if len(exactMatches) > 0 {
		// 查找匹配的指定类型
		// Look for matching specified type
		for _, unit := range exactMatches {
			if unit.UnitType == unitType {
				matchedUnit = unit
				break
			}
		}

		// 如果没有找到指定类型的匹配，返回错误
		// If no specified type match found, return error
		if matchedUnit == nil {
			return 0, nil, fmt.Errorf("unit type mismatch: expected %v but found %v", unitType, exactMatches[0].UnitType)
		}
	} else {
		// 如果没有精确匹配，使用大小写不敏感的匹配
		// If no exact match, use case-insensitive matching
		for i := range AllUnits {
			unit := &AllUnits[i]
			for _, symbol := range unit.Symbols {
				if strings.ToUpper(symbol) == unitStr {
					matchedUnit = unit
					break
				}
			}
			if matchedUnit != nil {
				break
			}
		}

		// 检查大小写不敏感匹配的单位类型
		// Check case-insensitive matched unit type
		if matchedUnit != nil && matchedUnit.UnitType != unitType {
			return 0, nil, fmt.Errorf("unit type mismatch: expected %v but found %v", unitType, matchedUnit.UnitType)
		}
	}

	if matchedUnit == nil {
		return 0, nil, fmt.Errorf("unknown unit: %s", unitStr)
	}

	// 计算基础单位的值
	// Calculate value in base unit
	baseValue := value * matchedUnit.Factor

	return baseValue, matchedUnit, nil
}

// FormatUnit 格式化单位值，支持多种输入格式和输出格式
//
// 参数:
//   - unitType: 单位类型
//   - precision: 小数位数精度
//   - toUnit: 输出格式，可以是：
//     -- "auto": 自动选择最合适的单位，默认值
//     -- 具体的单位符号（如 "MB", "s", "h" 等）
//     -- value: 输入值，可以是：
//     -- 数值（基础单位）
//     -- 带单位的字符串（如 "10MB", "5.5s"）
//
// 返回值:
//   - string: 格式化后的字符串
//   - error: 转换过程中的错误
//
// 示例:
//   - FormatUnit(10485760, UnitTypeStorage, 2, "auto") -> "10.00MB", nil
//   - FormatUnit(10485760, UnitTypeStorage, 2, "MB") -> "10.00MB", nil
//   - FormatUnit("10MB", UnitTypeStorage, 2, "auto") -> "10.00MB", nil
//   - FormatUnit("10MB", UnitTypeStorage, 2, "KB") -> "10240.00KB", nil
//   - FormatUnit(5500000000, UnitTypeTime, 1, "auto") -> "5.5s", nil
func FormatUnit(value interface{}, unitType UnitType, precision int, toUnit ...string) (string, error) {
	var baseValue float64

	// 解析输入值
	switch v := value.(type) {
	case float64:
		// 数值类型，假设为基础单位
		baseValue = v
	case int:
		baseValue = float64(v)
	case int64:
		baseValue = float64(v)
	case string:
		// 字符串类型，解析带单位的字符串
		parsedValue, _, parseErr := ParseUnitString(v, unitType)
		if parseErr != nil {
			return "", fmt.Errorf("failed to parse input value: %w", parseErr)
		}
		baseValue = parsedValue
	default:
		return "", fmt.Errorf("unsupported value type: %T", value)
	}

	// 处理输出格式（默认自动选择单位）
	if len(toUnit) == 0 || toUnit[0] == "auto" {
		// 自动选择最合适的单位
		return formatAutoUnit(baseValue, unitType, precision), nil
	} else {
		// 指定目标单位
		return formatToSpecificUnit(baseValue, toUnit[0], unitType, precision)
	}
}

// formatAutoUnit 自动选择最合适的单位进行格式化
func formatAutoUnit(value float64, unitType UnitType, precision int) string {
	// 获取对应类型的单位列表
	var units []Unit
	switch unitType {
	case UnitTypeStorage:
		units = StorageUnits
	case UnitTypeTime:
		units = TimeUnits
	case UnitTypeData:
		units = DataUnits
	default:
		units = StorageUnits // 默认使用存储单位
	}

	// 如果值为0，直接返回0和最小单位
	if value == 0 {
		return fmt.Sprintf("0%s", units[0].Symbols[0])
	}

	// 找到最合适的单位
	var bestUnit Unit
	var bestValue float64

	for i := len(units) - 1; i >= 0; i-- {
		unit := units[i]
		unitValue := value / unit.Factor

		if unitValue >= 1.0 || i == 0 {
			bestUnit = unit
			bestValue = unitValue
			break
		}
	}

	// 格式化数值
	format := fmt.Sprintf("%%.%df%%s", precision)
	return fmt.Sprintf(format, bestValue, bestUnit.Symbols[0])
}

// formatToSpecificUnit 格式化为指定单位
func formatToSpecificUnit(value float64, toUnit string, unitType UnitType, precision int) (string, error) {
	// 获取基础单位符号
	baseUnitSymbol := getBaseUnitSymbol(unitType)

	// 将基础单位值转换为目标单位
	convertedValue, err := ConvertUnit(value, baseUnitSymbol, toUnit, unitType)
	if err != nil {
		return "", fmt.Errorf("failed to convert to target unit: %w", err)
	}

	// 格式化数值
	format := fmt.Sprintf("%%.%df%%s", precision)
	return fmt.Sprintf(format, convertedValue, toUnit), nil
}

// getBaseUnitSymbol 获取基础单位的符号
// getBaseUnitSymbol gets the symbol of the base unit
func getBaseUnitSymbol(unitType UnitType) string {
	switch unitType {
	case UnitTypeStorage:
		return "B" // 字节
	case UnitTypeTime:
		return "ns" // 纳秒
	case UnitTypeData:
		return "b" // 位
	default:
		return "B" // 默认字节
	}
}

// ConvertUnit 在不同单位之间转换数值
// ConvertUnit converts values between different units
//
// 参数:
//   - value: 原始数值
//   - fromUnit: 原始单位符号
//   - toUnit: 目标单位符号
//   - unitType: 单位类型
//
// 返回值:
//   - float64: 转换后的数值
//   - error: 转换过程中的错误
//
// 示例:
//   - ConvertUnit(10, "MB", "KB", UnitTypeStorage) -> 10240, nil
//   - ConvertUnit(1, "h", "m", UnitTypeTime) -> 60, nil
//   - ConvertUnit(8, "b", "byte", UnitTypeData) -> 1, nil
func ConvertUnit(value float64, fromUnit, toUnit string, unitType UnitType) (float64, error) {
	// 解析原始单位
	// Parse source unit
	_, fromUnitInfo, err := ParseUnitString(fmt.Sprintf("1%s", fromUnit), unitType)
	if err != nil {
		return 0, fmt.Errorf("invalid source unit: %s", fromUnit)
	}

	// 解析目标单位
	// Parse target unit
	_, toUnitInfo, err := ParseUnitString(fmt.Sprintf("1%s", toUnit), unitType)
	if err != nil {
		return 0, fmt.Errorf("invalid target unit: %s", toUnit)
	}

	// 检查单位类型是否匹配
	// Check if unit types match
	if fromUnitInfo.UnitType != toUnitInfo.UnitType {
		return 0, fmt.Errorf("unit type mismatch: cannot convert from %s to %s", fromUnit, toUnit)
	}

	// 执行转换
	// Perform conversion
	baseValue := value * fromUnitInfo.Factor
	convertedValue := baseValue / toUnitInfo.Factor

	return convertedValue, nil
}

// IsValidUnit 检查单位字符串是否有效
// IsValidUnit checks if a unit string is valid
//
// 参数:
//   - unitStr: 单位字符串
//   - unitType: 单位类型
//
// 返回值:
//   - bool: 单位是否有效
//
// 示例:
//   - IsValidUnit("MB", UnitTypeStorage) -> true
//   - IsValidUnit("h", UnitTypeTime) -> true
//   - IsValidUnit("b", UnitTypeData) -> true
func IsValidUnit(unitStr string, unitType UnitType) bool {
	_, _, err := ParseUnitString(fmt.Sprintf("1%s", unitStr), unitType)
	return err == nil
}

// GetUnitType 获取单位的类型
// GetUnitType gets the type of a unit
//
// 参数:
//   - unitStr: 单位字符串
//   - unitType: 单位类型
//
// 返回值:
//   - UnitType: 单位类型
//   - error: 解析错误
//
// 示例:
//   - GetUnitType("MB", UnitTypeStorage) -> UnitTypeStorage, nil
//   - GetUnitType("h", UnitTypeTime) -> UnitTypeTime, nil
//   - GetUnitType("b", UnitTypeData) -> UnitTypeData, nil
func GetUnitType(unitStr string, unitType UnitType) (UnitType, error) {
	_, unitInfo, err := ParseUnitString(fmt.Sprintf("1%s", unitStr), unitType)
	if err != nil {
		return UnitTypeStorage, err
	}
	return unitInfo.UnitType, nil
}

// GetAvailableUnits 获取指定类型的所有可用单位
// GetAvailableUnits gets all available units of the specified type
//
// 参数:
//   - unitType: 单位类型
//
// 返回值:
//   - []string: 可用的单位符号列表
func GetAvailableUnits(unitType UnitType) []string {
	var units []Unit
	switch unitType {
	case UnitTypeStorage:
		units = StorageUnits
	case UnitTypeTime:
		units = TimeUnits
	case UnitTypeData:
		units = DataUnits
	default:
		return []string{}
	}

	var symbols []string
	for _, unit := range units {
		symbols = append(symbols, unit.Symbols[0])
	}
	return symbols
}

// RoundUnitValue 对单位值进行四舍五入
// RoundUnitValue rounds a unit value
//
// 参数:
//   - value: 原始值
//   - unit: 单位符号
//   - precision: 小数位数精度
//   - unitType: 单位类型
//
// 返回值:
//   - float64: 四舍五入后的值
//   - error: 转换错误
func RoundUnitValue(value float64, unit string, precision int, unitType UnitType) (float64, error) {
	converted, err := ConvertUnit(value, unit, unit, unitType)
	if err != nil {
		return 0, err
	}

	factor := math.Pow(10, float64(precision))
	return math.Round(converted*factor) / factor, nil
}
