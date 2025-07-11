package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
	"log"
	"reflect"
)

// ActionSave 保存数据的主要处理方法（自动判断创建或更新）
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
func (t *Trait) ActionSave(c *gin.Context) {
	t.InitCrud(c)
	id, _ := t.ExtractPkId()

	if !helper.IsEmptyValue(id) {
		t.ActionUpdate(c)
	} else {
		t.ActionCreate(c)
	}
}

// SaveFormData 获取保存操作的表单数据
// 返回值：
//   - modelValue interface{}: 绑定后的模型实例
//   - mapData map[string]any: 原始表单数据映射
//   - err error: 处理过程中的错误信息
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
// 参数说明：
// - modelValue: 要保存的模型数据
// - mapData: 表单数据映射
// - originalData: 原始数据（仅在更新操作时提供，创建操作时为nil）
func (t *Trait) SaveBefore(modelValue interface{}, mapData map[string]any, originalData ...interface{}) (interface{}, map[string]any, error) {
	// 可以在此处添加一些前置处理逻辑
	// 如果是更新操作，可以通过 len(originalData) > 0 && originalData[0] != nil 来判断并获取原始数据
	return modelValue, mapData, nil
}

// SaveAfter 保存后的钩子方法，用于后续处理（在事务内执行）
// 参数说明：
//   - tx *gorm.DB: 数据库事务对象
//   - modelValue interface{}: 已保存的模型实例
//
// 返回值：
//   - error: 处理过程中的错误信息，如果返回错误则会回滚事务
func (t *Trait) SaveAfter(tx *gorm.DB, modelValue interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return nil
}
