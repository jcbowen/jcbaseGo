package main

import (
	"fmt"

	"github.com/jcbowen/jcbaseGo/component/helper"
)

func main() {
	fmt.Println("=== Helper 类型转换工具使用示例 ===\n")

	// 1. 字符串转换示例
	fmt.Println("1. 字符串转换示例:")
	stringConvertExample()

	// 2. 数字转换示例
	fmt.Println("\n2. 数字转换示例:")
	numberConvertExample()

	// 3. 布尔值转换示例
	fmt.Println("\n3. 布尔值转换示例:")
	boolConvertExample()

	// 4. 切片转换示例
	fmt.Println("\n4. 切片转换示例:")
	sliceConvertExample()
}

// stringConvertExample 字符串转换示例
func stringConvertExample() {
	// 字符串转整数
	str := "123"
	convert := helper.Convert{Value: str}
	num := convert.ToInt()
	fmt.Printf("字符串 '%s' 转换为整数: %d\n", str, num)

	// 字符串转浮点数
	floatStr := "123.45"
	convertFloat := helper.Convert{Value: floatStr}
	floatNum := convertFloat.ToFloat64()
	fmt.Printf("字符串 '%s' 转换为浮点数: %f\n", floatStr, floatNum)

	// 整数转字符串
	intNum := 456
	convertInt := helper.Convert{Value: intNum}
	intStr := convertInt.ToString()
	fmt.Printf("整数 %d 转换为字符串: '%s'\n", intNum, intStr)

	// 浮点数转字符串
	floatNum2 := 789.12
	convertFloat2 := helper.Convert{Value: floatNum2}
	floatStr2 := convertFloat2.ToString()
	fmt.Printf("浮点数 %f 转换为字符串: '%s'\n", floatNum2, floatStr2)
}

// numberConvertExample 数字转换示例
func numberConvertExample() {
	// 整数转浮点数
	intNum := 100
	convertInt := helper.Convert{Value: intNum}
	floatNum := convertInt.ToFloat64()
	fmt.Printf("整数 %d 转换为浮点数: %f\n", intNum, floatNum)

	// 浮点数转整数
	floatNum2 := 200.7
	convertFloat := helper.Convert{Value: floatNum2}
	intNum2 := convertFloat.ToInt()
	fmt.Printf("浮点数 %f 转换为整数: %d\n", floatNum2, intNum2)

	// 64位整数转换
	int64Num := int64(300)
	convertInt64 := helper.Convert{Value: int64Num}
	int32Num := int32(convertInt64.ToInt())
	fmt.Printf("64位整数 %d 转换为32位整数: %d\n", int64Num, int32Num)

	// 32位整数转64位
	int32Num2 := int32(400)
	convertInt32 := helper.Convert{Value: int32Num2}
	int64Num2 := convertInt32.ToInt64()
	fmt.Printf("32位整数 %d 转换为64位整数: %d\n", int32Num2, int64Num2)
}

// boolConvertExample 布尔值转换示例
func boolConvertExample() {
	// 字符串转布尔值
	trueStr := "true"
	falseStr := "false"
	invalidStr := "invalid"

	convertTrue := helper.Convert{Value: trueStr}
	bool1 := convertTrue.ToBool()
	fmt.Printf("字符串 '%s' 转换为布尔值: %t\n", trueStr, bool1)

	convertFalse := helper.Convert{Value: falseStr}
	bool2 := convertFalse.ToBool()
	fmt.Printf("字符串 '%s' 转换为布尔值: %t\n", falseStr, bool2)

	convertInvalid := helper.Convert{Value: invalidStr}
	bool3 := convertInvalid.ToBool()
	fmt.Printf("无效字符串 '%s' 转换为布尔值: %t\n", invalidStr, bool3)

	// 布尔值转字符串
	boolVal := true
	convertBool := helper.Convert{Value: boolVal}
	boolStr := convertBool.ToString()
	fmt.Printf("布尔值 %t 转换为字符串: '%s'\n", boolVal, boolStr)
}

// sliceConvertExample 切片转换示例
func sliceConvertExample() {
	// 字符串切片转接口切片
	strSlice := []string{"a", "b", "c"}
	convertStrSlice := helper.Convert{Value: strSlice}
	interfaceSlice := convertStrSlice.ToString()
	fmt.Printf("字符串切片 %v 转换为字符串: %s\n", strSlice, interfaceSlice)

	// 接口切片转字符串
	interfaceSlice2 := []interface{}{"x", "y", "z"}
	convertInterfaceSlice := helper.Convert{Value: interfaceSlice2}
	strSlice2 := convertInterfaceSlice.ToString()
	fmt.Printf("接口切片 %v 转换为字符串: %s\n", interfaceSlice2, strSlice2)

	// 整数切片转换
	intSlice := []int{1, 2, 3}
	convertIntSlice := helper.Convert{Value: intSlice}
	intSliceStr := convertIntSlice.ToString()
	fmt.Printf("整数切片 %v 转换为字符串: %s\n", intSlice, intSliceStr)

	// 64位整数切片转字符串
	int64Slice2 := []int64{4, 5, 6}
	convertInt64Slice := helper.Convert{Value: int64Slice2}
	int64SliceStr := convertInt64Slice.ToString()
	fmt.Printf("64位整数切片 %v 转换为字符串: %s\n", int64Slice2, int64SliceStr)
}
