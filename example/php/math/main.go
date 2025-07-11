package main

import (
	"fmt"
	"log"

	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/php"
)

func main() {
	fmt.Println("=== PHP 数学计算示例 ===")

	// 创建 PHP 组件实例
	phpComponent := php.New(jcbaseGo.Option{})

	// 示例1: 基本数学运算
	fmt.Println("\n1. 基本数学运算:")

	// 幂运算
	result, err := phpComponent.RunFunc("pow", "2", "3")
	if err != nil {
		log.Printf("调用 pow 失败: %v", err)
	} else {
		fmt.Printf("pow(2, 3) = %s\n", result)
	}

	// 平方根
	result, err = phpComponent.RunFunc("sqrt", "16")
	if err != nil {
		log.Printf("调用 sqrt 失败: %v", err)
	} else {
		fmt.Printf("sqrt(16) = %s\n", result)
	}

	// 绝对值
	result, err = phpComponent.RunFunc("abs", "-123.45")
	if err != nil {
		log.Printf("调用 abs 失败: %v", err)
	} else {
		fmt.Printf("abs(-123.45) = %s\n", result)
	}

	// 向上取整
	result, err = phpComponent.RunFunc("ceil", "3.14")
	if err != nil {
		log.Printf("调用 ceil 失败: %v", err)
	} else {
		fmt.Printf("ceil(3.14) = %s\n", result)
	}

	// 向下取整
	result, err = phpComponent.RunFunc("floor", "3.99")
	if err != nil {
		log.Printf("调用 floor 失败: %v", err)
	} else {
		fmt.Printf("floor(3.99) = %s\n", result)
	}

	// 四舍五入
	result, err = phpComponent.RunFunc("round", "3.14159", "2")
	if err != nil {
		log.Printf("调用 round 失败: %v", err)
	} else {
		fmt.Printf("round(3.14159, 2) = %s\n", result)
	}

	// 示例2: 三角函数
	fmt.Println("\n2. 三角函数:")

	// 正弦函数（弧度）
	result, err = phpComponent.RunFunc("sin", "1.5708")
	if err != nil {
		log.Printf("调用 sin 失败: %v", err)
	} else {
		fmt.Printf("sin(π/2) = %s\n", result)
	}

	// 余弦函数（弧度）
	result, err = phpComponent.RunFunc("cos", "0")
	if err != nil {
		log.Printf("调用 cos 失败: %v", err)
	} else {
		fmt.Printf("cos(0) = %s\n", result)
	}

	// 正切函数（弧度）
	result, err = phpComponent.RunFunc("tan", "0.7854")
	if err != nil {
		log.Printf("调用 tan 失败: %v", err)
	} else {
		fmt.Printf("tan(π/4) = %s\n", result)
	}

	// 反正弦函数
	result, err = phpComponent.RunFunc("asin", "1")
	if err != nil {
		log.Printf("调用 asin 失败: %v", err)
	} else {
		fmt.Printf("asin(1) = %s\n", result)
	}

	// 反余弦函数
	result, err = phpComponent.RunFunc("acos", "0")
	if err != nil {
		log.Printf("调用 acos 失败: %v", err)
	} else {
		fmt.Printf("acos(0) = %s\n", result)
	}

	// 反正切函数
	result, err = phpComponent.RunFunc("atan", "1")
	if err != nil {
		log.Printf("调用 atan 失败: %v", err)
	} else {
		fmt.Printf("atan(1) = %s\n", result)
	}

	// 示例3: 对数和指数函数
	fmt.Println("\n3. 对数和指数函数:")

	// 自然对数
	result, err = phpComponent.RunFunc("log", "2.718")
	if err != nil {
		log.Printf("调用 log 失败: %v", err)
	} else {
		fmt.Printf("ln(e) = %s\n", result)
	}

	// 以10为底的对数
	result, err = phpComponent.RunFunc("log10", "100")
	if err != nil {
		log.Printf("调用 log10 失败: %v", err)
	} else {
		fmt.Printf("log10(100) = %s\n", result)
	}

	// 自然指数函数
	result, err = phpComponent.RunFunc("exp", "1")
	if err != nil {
		log.Printf("调用 exp 失败: %v", err)
	} else {
		fmt.Printf("exp(1) = %s\n", result)
	}

	// 示例4: 随机数生成
	fmt.Println("\n4. 随机数生成:")

	// 生成随机整数
	result, err = phpComponent.RunFunc("rand", "1", "100")
	if err != nil {
		log.Printf("调用 rand 失败: %v", err)
	} else {
		fmt.Printf("rand(1, 100) = %s\n", result)
	}

	// 生成随机浮点数
	result, err = phpComponent.RunFunc("lcg_value")
	if err != nil {
		log.Printf("调用 lcg_value 失败: %v", err)
	} else {
		fmt.Printf("lcg_value() = %s\n", result)
	}

	// 生成随机字节
	result, err = phpComponent.RunFunc("random_bytes", "8")
	if err != nil {
		log.Printf("调用 random_bytes 失败: %v", err)
	} else {
		fmt.Printf("random_bytes(8) = %s\n", result)
	}

	// 示例5: 数值比较和验证
	fmt.Println("\n5. 数值比较和验证:")

	// 获取最大值
	result, err = phpComponent.RunFunc("max", "10", "20", "5", "30")
	if err != nil {
		log.Printf("调用 max 失败: %v", err)
	} else {
		fmt.Printf("max(10, 20, 5, 30) = %s\n", result)
	}

	// 获取最小值
	result, err = phpComponent.RunFunc("min", "10", "20", "5", "30")
	if err != nil {
		log.Printf("调用 min 失败: %v", err)
	} else {
		fmt.Printf("min(10, 20, 5, 30) = %s\n", result)
	}

	// 检查是否为有限数
	result, err = phpComponent.RunFunc("is_finite", "123.45")
	if err != nil {
		log.Printf("调用 is_finite 失败: %v", err)
	} else {
		fmt.Printf("is_finite(123.45) = %s\n", result)
	}

	// 检查是否为无限数
	result, err = phpComponent.RunFunc("is_infinite", "INF")
	if err != nil {
		log.Printf("调用 is_infinite 失败: %v", err)
	} else {
		fmt.Printf("is_infinite(INF) = %s\n", result)
	}

	// 检查是否为 NaN
	result, err = phpComponent.RunFunc("is_nan", "NAN")
	if err != nil {
		log.Printf("调用 is_nan 失败: %v", err)
	} else {
		fmt.Printf("is_nan(NAN) = %s\n", result)
	}

	// 示例6: 进制转换
	fmt.Println("\n6. 进制转换:")

	// 十进制转二进制
	result, err = phpComponent.RunFunc("decbin", "255")
	if err != nil {
		log.Printf("调用 decbin 失败: %v", err)
	} else {
		fmt.Printf("decbin(255) = %s\n", result)
	}

	// 十进制转八进制
	result, err = phpComponent.RunFunc("decoct", "255")
	if err != nil {
		log.Printf("调用 decoct 失败: %v", err)
	} else {
		fmt.Printf("decoct(255) = %s\n", result)
	}

	// 十进制转十六进制
	result, err = phpComponent.RunFunc("dechex", "255")
	if err != nil {
		log.Printf("调用 dechex 失败: %v", err)
	} else {
		fmt.Printf("dechex(255) = %s\n", result)
	}

	// 二进制转十进制
	result, err = phpComponent.RunFunc("bindec", "11111111")
	if err != nil {
		log.Printf("调用 bindec 失败: %v", err)
	} else {
		fmt.Printf("bindec(11111111) = %s\n", result)
	}

	// 八进制转十进制
	result, err = phpComponent.RunFunc("octdec", "377")
	if err != nil {
		log.Printf("调用 octdec 失败: %v", err)
	} else {
		fmt.Printf("octdec(377) = %s\n", result)
	}

	// 十六进制转十进制
	result, err = phpComponent.RunFunc("hexdec", "FF")
	if err != nil {
		log.Printf("调用 hexdec 失败: %v", err)
	} else {
		fmt.Printf("hexdec(FF) = %s\n", result)
	}

	// 示例7: 数学常量
	fmt.Println("\n7. 数学常量:")

	// 圆周率
	result, err = phpComponent.RunFunc("M_PI")
	if err != nil {
		log.Printf("获取 M_PI 失败: %v", err)
	} else {
		fmt.Printf("M_PI = %s\n", result)
	}

	// 自然对数的底
	result, err = phpComponent.RunFunc("M_E")
	if err != nil {
		log.Printf("获取 M_E 失败: %v", err)
	} else {
		fmt.Printf("M_E = %s\n", result)
	}

	// 示例8: 复杂数学计算
	fmt.Println("\n8. 复杂数学计算:")

	// 计算复数的模
	result, err = phpComponent.RunFunc("hypot", "3", "4")
	if err != nil {
		log.Printf("调用 hypot 失败: %v", err)
	} else {
		fmt.Printf("hypot(3, 4) = %s\n", result)
	}

	// 计算阶乘
	result, err = phpComponent.RunFunc("gmp_fact", "5")
	if err != nil {
		log.Printf("调用 gmp_fact 失败: %v", err)
	} else {
		fmt.Printf("5! = %s\n", result)
	}

	// 计算最大公约数
	result, err = phpComponent.RunFunc("gmp_gcd", "48", "18")
	if err != nil {
		log.Printf("调用 gmp_gcd 失败: %v", err)
	} else {
		fmt.Printf("gcd(48, 18) = %s\n", result)
	}

	// 计算最小公倍数
	result, err = phpComponent.RunFunc("gmp_lcm", "12", "18")
	if err != nil {
		log.Printf("调用 gmp_lcm 失败: %v", err)
	} else {
		fmt.Printf("lcm(12, 18) = %s\n", result)
	}

	fmt.Println("\n=== 数学计算示例完成 ===")
}
