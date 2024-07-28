package crud

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
	"log"
	"reflect"
	"strconv"
)

func (t *Trait) ActionSave(c *gin.Context) {
	t.checkInit(c)
	id, _ := t.ExtractPkId()

	if !helper.IsEmptyValue(id) {
		t.ActionUpdate(c)
	} else {
		t.ActionCreate(c)
	}
}

// GetSaveFormData 获取表单数据
func (t *Trait) GetSaveFormData() (modelValue interface{}, mapData map[string]any, err error) {
	gpcInterface, GPCExists := t.GinContext.Get("GPC")
	if !GPCExists {
		return
	}
	formDataMap := gpcInterface.(map[string]map[string]any)["all"]

	// 安全过滤
	sanitizedMapData := security.Input{Value: formDataMap}.Sanitize().(map[interface{}]interface{})
	// 格式转换
	mapData = make(map[string]interface{})
	for key, value := range sanitizedMapData {
		strKey := key.(string)
		mapData[strKey] = value
	}

	// 动态创建模型实例
	modelType := reflect.TypeOf(t.Model).Elem()
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	modelValue = reflect.New(modelType).Interface()

	// 将 mapData 转换为适合模型字段类型的值
	err = t.BindMapToStruct(mapData, modelValue)
	if err != nil {
		log.Panic(err)
		return
	}

	return modelValue, mapData, nil
}

// BeforeSave 保存前
func (t *Trait) BeforeSave(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	// 可以在此处添加一些前置处理逻辑
	return modelValue, mapData, nil
}

// AfterSave 保存后
func (t *Trait) AfterSave(modelValue interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return nil
}

// BindMapToStruct 将 map 数据绑定到 struct，并处理类型转换
func (t *Trait) BindMapToStruct(mapData map[string]any, modelValue interface{}) error {
	val := reflect.ValueOf(modelValue)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("modelValue must be a non-nil pointer")
	}

	modelVal := val.Elem()
	modelType := modelVal.Type()

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = field.Name
		}

		if val, ok := mapData[fieldName]; ok {
			fieldVal := modelVal.Field(i)
			if !fieldVal.CanSet() {
				log.Printf("Field %s cannot be set\n", fieldName)
				continue
			}

			err := t.setValue(fieldVal, val)
			if err != nil {
				log.Printf("Error setting field %s: %v\n", fieldName, err)
				return err
			}
		}
	}

	return nil
}

// setValue 根据字段类型设置值
func (t *Trait) setValue(fieldVal reflect.Value, val interface{}) error {
	switch fieldVal.Kind() {
	case reflect.String:
		strVal, ok := val.(string)
		if !ok {
			return errors.New("cannot convert to string")
		}
		fieldVal.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(val.(string), 10, 64)
		if err != nil {
			return err
		}
		fieldVal.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(val.(string), 10, 64)
		if err != nil {
			return err
		}
		fieldVal.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(val.(string), 64)
		if err != nil {
			return err
		}
		fieldVal.SetFloat(floatVal)
	default:
		return errors.New("unsupported field type")
	}

	return nil
}
