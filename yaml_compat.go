package jcbaseGo

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// yaml 键名别名兼容层。
//
// 背景：DbStruct 的部分驼峰字段（parseTime、driverName 等）自 v0.17.7 起补齐了 yaml tag，
// 导致此前在 yaml 链路上生效的全小写键（yaml.v3 对无 tag 字段只认「字段名全小写」）被静默忽略：
//   - yaml.Unmarshal 不校验未知键，旧键失效无任何报错；
//   - 失效字段留下的零值会被 CheckAndSetDefault 填成 default（如 parsetime 丢失后被填为 False，
//     时间解析被静默关闭）；
//   - 启动时的 updateConfigFile 会把丢值后的状态重写回配置文件，错误被固化。
//
// 本层为每个字段建立「规范键（yaml tag 名或小写字段名）↔ 小写别名」的映射，在解析前把
// 别名键改写为规范键：旧配置文件（小写键）与新写法（驼峰键）同时可用，同一字段两种键
// 同现时报错以消除歧义，无法识别的键打启动告警（对齐 yaml.v3 默认的忽略行为，仅可见化）。

// buildYAMLKeyMap 构建结构体字段的 yaml 键名映射表。
// 规范键的取值规则：字段有 yaml tag 时取 tag 名（逗号前段，"-" 表示跳过该字段），
// 无 tag 时取 yaml.v3 的默认规则——字段名全小写。
// 参数：
//   - t: 结构体反射类型
//
// 返回值：
//   - map[string]string: 键为「键名的小写形式」（规范键与别名共用小写形式定位），值为规范键
//
// 使用示例：
//   - keyMap := buildYAMLKeyMap(reflect.TypeOf(DbStruct{})) // 含 "parsetime" -> "parseTime"
func buildYAMLKeyMap(t reflect.Type) map[string]string {
	keyMap := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := strings.Split(field.Tag.Get("yaml"), ",")[0]
		if name == "-" {
			continue
		}
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		keyMap[strings.ToLower(name)] = name
	}
	return keyMap
}

// normalizeYAMLStructKeys 把 yaml mapping 节点的键名归一化为结构体字段对应的规范键。
// 规则：
//   - 键的小写形式命中映射表时，改写键值为规范键（别名归一化，如 parsetime -> parseTime）；
//   - 同一字段同时出现两种键名写法（如 parseTime 与 parsetime 同现）返回错误，由调用方决定如何处理；
//   - 未命中的键打启动告警后保持原样（与 yaml.v3 默认的忽略行为一致，仅让配置错误可见）。
//
// 参数：
//   - node: 待归一化的 yaml 节点；非 mapping 节点（含 null）直接跳过，交由后续 Decode 处理
//   - t: 目标结构体反射类型
//   - typeName: 结构体名称，用于告警与错误信息定位
//
// 返回值：
//   - error: 键名冲突时返回描述两种写法的错误；其余情况返回 nil
//
// 使用示例：
//   - err := normalizeYAMLStructKeys(node, reflect.TypeOf(DbStruct{}), "DbStruct")
func normalizeYAMLStructKeys(node *yaml.Node, t reflect.Type, typeName string) error {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	keyMap := buildYAMLKeyMap(t)
	seen := make(map[string]string) // 规范键 -> 首次出现的原始键，用于冲突检测

	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		// 只处理标量键；非标量键交给 Decode 按其自身规则报错
		if keyNode.Kind != yaml.ScalarNode {
			continue
		}

		canonical, ok := keyMap[strings.ToLower(keyNode.Value)]
		if !ok {
			log.Printf("[jcbaseGo] %s 配置项 %q 无法识别，已忽略，请检查键名拼写", typeName, keyNode.Value)
			continue
		}

		if prev, dup := seen[canonical]; dup {
			return fmt.Errorf("%s 配置项同时出现了 %q 与 %q（同一字段的两种键名写法），请只保留一种", typeName, prev, keyNode.Value)
		}
		seen[canonical] = keyNode.Value

		// 别名归一化：小写键改写为规范键，Decode 与后续 Marshal 写回均使用规范键
		keyNode.Value = canonical
	}
	return nil
}

// UnmarshalYAML 实现 yaml.v3 的自定义解析：解析前先做键名别名归一化。
// 通过 type alias 绕过本方法自身，避免 Decode 递归触发 UnmarshalYAML。
// 参数：
//   - value: 待解析的 yaml 节点
//
// 返回值：
//   - error: 键名冲突或节点解析失败时返回对应错误
func (c *DbStruct) UnmarshalYAML(value *yaml.Node) error {
	if err := normalizeYAMLStructKeys(value, reflect.TypeOf(*c), "DbStruct"); err != nil {
		return err
	}
	type dbStructAlias DbStruct
	return value.Decode((*dbStructAlias)(c))
}

// UnmarshalYAML 实现 yaml.v3 的自定义解析：解析前先做键名别名归一化（与 DbStruct 同规则）。
// 参数：
//   - value: 待解析的 yaml 节点
//
// 返回值：
//   - error: 键名冲突或节点解析失败时返回对应错误
func (c *SqlLiteStruct) UnmarshalYAML(value *yaml.Node) error {
	if err := normalizeYAMLStructKeys(value, reflect.TypeOf(*c), "SqlLiteStruct"); err != nil {
		return err
	}
	type sqlLiteStructAlias SqlLiteStruct
	return value.Decode((*sqlLiteStructAlias)(c))
}
