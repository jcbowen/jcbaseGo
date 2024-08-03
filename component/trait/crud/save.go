package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
	"log"
	"reflect"
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

// SaveFormData 获取表单数据
func (t *Trait) SaveFormData() (modelValue interface{}, mapData map[string]any, err error) {
	mapData = t.GetSafeMapGPC("all")

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

// SaveBefore 保存前
func (t *Trait) SaveBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	// 可以在此处添加一些前置处理逻辑
	return modelValue, mapData, nil
}

// AfterSave 保存后
func (t *Trait) AfterSave(tx *gorm.DB, modelValue interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return nil
}
