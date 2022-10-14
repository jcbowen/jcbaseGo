package data

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Struct2JsonOpt struct {
	Struct interface{}
	//Sort     bool
	NeedFile bool
	FilePath string
}

func SetStruct(_struct interface{}) *Struct2JsonOpt {
	return &Struct2JsonOpt{Struct: _struct}
}

// 由于map是无序的，所以json不支持排序，只能有序的一个个好输出
/*func (opt *Struct2JsonOpt) DoSort(sort bool) *Struct2JsonOpt {
	opt.Sort = sort
	return opt
}*/

func (opt *Struct2JsonOpt) File(filepath string) *Struct2JsonOpt {
	opt.NeedFile = true
	opt.FilePath = filepath
	return opt
}

func (opt *Struct2JsonOpt) ToJson() (string, error) {
	_struct := opt.Struct

	jsonByte, err := json.Marshal(_struct)
	jsonStr := string(jsonByte)

	/*if opt.Sort {
		jsonMap := JsonStr2Map(jsonStr)
		nData := SetMapStrInterface(jsonMap).DoSort(true).GetData()
		jsonByte, err = json.Marshal(nData)
		jsonStr = string(jsonByte)
	}*/

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

// ToString 获取变量的字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func ToString(value interface{}) (key string) {
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return
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

// ----- map[string]string 类型相关操作 -----/

type MapStrInterface struct {
	Data map[string]interface{}
	Keys []string
	Sort bool
}

func SetMapStrInterface(data map[string]interface{}) *MapStrInterface {
	return &MapStrInterface{Data: data}
}

func (d *MapStrInterface) DoSort(sort bool) *MapStrInterface {
	d.Sort = sort
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

// ----- []string 类型相关操作 -----/

type ArrStr struct {
	Arr  []string
	Sort bool // 执行ArrayValue方法时是否排序
}

func SetArrStr(str []string) *ArrStr {
	return &ArrStr{Arr: str, Sort: true}
}

// DoSort 设置ArrayValue方法¬是否排序
func (a *ArrStr) DoSort(sort bool) *ArrStr {
	a.Sort = sort
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
