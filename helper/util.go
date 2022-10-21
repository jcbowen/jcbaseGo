package helper

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// ----- 结构体转json，Begin -----/
// 用法:
// 1. SetStruct(_struct).ToJson()
// 2. SetStruct(_struct).DoSort().ToJson()
// 3. SetStruct(_struct).File("path").ToJson()

type Struct2JsonOpt struct {
	Struct   interface{} // 结构体
	Sort     bool        // 是否需要排序
	NeedFile bool        // 是否需要输出json文件
	FilePath string      // 输出json文件路径
}

func SetStruct(_struct interface{}) *Struct2JsonOpt {
	return &Struct2JsonOpt{Struct: _struct}
}

func (opt *Struct2JsonOpt) DoSort() *Struct2JsonOpt {
	opt.Sort = true
	return opt
}

func (opt *Struct2JsonOpt) File(filepath string) *Struct2JsonOpt {
	opt.NeedFile = true
	opt.FilePath = filepath
	return opt
}

func (opt *Struct2JsonOpt) ToJson() (string, error) {
	_struct := opt.Struct

	jsonByte, err := json.Marshal(_struct)
	jsonStr := string(jsonByte)

	if opt.Sort {
		jsonStr = JsonStrSort(jsonStr)
	}

	// 判断是否需要输出json文件
	if opt.NeedFile {
		_ = os.MkdirAll(filepath.Dir(opt.FilePath), os.ModePerm)
		cfgFile, err2 := os.Create(opt.FilePath)
		if err2 != nil {
			panic(err2)
		}
		defer func(cfgFile *os.File) {
			err := cfgFile.Close()
			if err != nil {
				panic(err)
			}
		}(cfgFile)

		// 编码写入配置文件;
		cfgEncoder := json.NewEncoder(cfgFile)
		cfgEncoder.SetIndent("", "\t")
		if err3 := cfgEncoder.Encode(_struct); err3 != nil {
			panic(err3)
		}
	}

	return jsonStr, err
}

// ----- 结构体转json，End -----/

// ----- map[string]string 类型相关操作 -----/

type MapStrInterface struct {
	Data map[string]interface{}
	Keys []string
	Sort bool
}

func SetMapStrInterface(data map[string]interface{}) *MapStrInterface {
	return &MapStrInterface{Data: data}
}

func (d *MapStrInterface) DoSort() *MapStrInterface {
	d.Sort = true
	return d
}

func (d *MapStrInterface) ArrayKeys() []string {
	if len(d.Data) == 0 {
		return d.Keys
	}

	for k := range d.Data {
		d.Keys = append(d.Keys, k)
	}

	if d.Sort {
		sort.Strings(d.Keys)
	}

	return d.Keys
}

func (d *MapStrInterface) ArrayValues() []interface{} {
	var values []interface{}

	if len(d.Data) == 0 {
		return values
	}

	if d.Sort {
		for _, k := range d.ArrayKeys() {
			values = append(values, d.Data[k])
		}
	} else {
		for _, v := range d.Data {
			values = append(values, v)
		}
	}

	return values
}

func (d *MapStrInterface) GetData() map[string]interface{} {
	if d.Sort {
		data := make(map[string]interface{})
		for _, k := range d.ArrayKeys() {
			data[k] = d.Data[k]
		}
		return data
	}

	return d.Data
}

// ----- []string 类型相关操作 -----/

type ArrStr struct {
	Arr  []string // 数组
	Sort bool     // 执行ArrayValue方法时是否排序
}

func SetArrStr(str []string) *ArrStr {
	return &ArrStr{Arr: str, Sort: false}
}

// DoSort 设置ArrayValue方法¬是否排序
func (a *ArrStr) DoSort() *ArrStr {
	a.Sort = true
	return a
}

func (a *ArrStr) ArrayValue() (value []string) {
	if len(a.Arr) == 0 {
		return
	}
	for _, v := range a.Arr {
		value = append(value, v)
	}
	if a.Sort {
		sort.Strings(value)
	}
	return
}

func (a *ArrStr) ArrayDiff(oArr ...[]string) (diff []string) {
	if len(a.Arr) == 0 {
		return
	}
	if len(a.Arr) > 0 && len(oArr) == 0 {
		diff = a.Arr
		return
	}
	for _, o := range oArr {
		for _, item := range a.Arr {
			if !InArray(item, o) {
				diff = append(diff, item)
			}
		}
	}
	return
}

func (a *ArrStr) ArrayIntersect(oArr ...[]string) (intersects []string) {
	if len(a.Arr) == 0 {
		return
	}
	if len(a.Arr) > 0 && len(oArr) == 0 {
		intersects = a.Arr
		return
	}
	var tmp = make(map[string]int, len(a.Arr))
	for _, v := range a.Arr {
		tmp[v] = 1
	}
	for _, param := range oArr {
		for _, arg := range param {
			if tmp[arg] != 0 {
				tmp[arg]++
			}
		}
	}
	for k, v := range tmp {
		if v > 1 {
			intersects = append(intersects, k)
		}
	}
	return
}

// ----- Json -----/

func JsonStr2Map(str string) map[string]interface{} {
	var tempMap map[string]interface{}
	err := json.Unmarshal([]byte(str), &tempMap)
	if err != nil {
		panic(err)
	}
	return tempMap
}

// JsonStrSort 对json字符串进行排序
func JsonStrSort(jsonStr string) string {
	jsonMap := JsonStr2Map(jsonStr)
	nData := SetMapStrInterface(jsonMap).DoSort().GetData()
	jsonByte, _ := json.Marshal(nData)
	return string(jsonByte)
}

// Str2Int 字符串转数字
//
// Deprecated: As of jcbaseGo 0.2.1, this function simply calls ToInt.
func Str2Int(str string) int {
	return ToInt(str)
}

// StrReplace 类似于php中的str_replace
func StrReplace(search interface{}, replace interface{}, subject interface{}, count int) (interface{}, error) {
	switch search.(type) {
	case string:
		switch replace.(type) {
		case string:
			switch subject.(type) {
			case string:
				return strings.Replace(subject.(string), search.(string), replace.(string), count), nil
			case []string:
				var slice []string
				for _, v := range subject.([]string) {
					slice = append(slice, strings.Replace(v, search.(string), replace.(string), count))
				}
				return slice, nil
			default:
				return nil, errors.New("invalid parameters")
			}
		default:
			return nil, errors.New("invalid parameters")
		}
	case []string:
		switch replace.(type) {
		case string:
			switch subject.(type) {
			case string:
				sub := subject.(string)

				for _, v := range search.([]string) {
					sub = strings.Replace(sub, v, replace.(string), count)
				}
				return sub, nil

			case []string:
				var slice []string
				for _, v := range subject.([]string) {
					sli, err := StrReplace(search, replace, v, count)
					if err != nil {
						return nil, err
					}
					slice = append(slice, sli.(string))
				}
				return slice, nil
			default:
				return nil, errors.New("invalid parameters")
			}
		case []string:
			switch subject.(type) {
			case string:
				rep := replace.([]string)
				sub := subject.(string)
				for i, s := range search.([]string) {
					if i < len(rep) {
						sub = strings.Replace(sub, s, rep[i], count)
					} else {
						sub = strings.Replace(sub, s, "", count)
					}
				}
				return sub, nil
			case []string:
				var slice []string
				for _, v := range subject.([]string) {
					sli, err := StrReplace(search, replace, v, count)
					if err != nil {
						return nil, err
					}
					slice = append(slice, sli.(string))
				}
				return slice, nil
			default:
				return nil, errors.New("invalid parameters")
			}
		default:
			return nil, errors.New("invalid parameters")
		}
	default:
		return nil, errors.New("invalid parameters")
	}
}

func InArray(val interface{}, array interface{}) (exists bool) {
	exists = false
	//index = -1
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				//index = i
				exists = true
				return
			}
		}
	}
	return
}

// ParseIP 解析IP地址，输出是ipv4或ipv6
// 0: invalid ip
// 4: ipv4
// 6: ipv6
func ParseIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}
