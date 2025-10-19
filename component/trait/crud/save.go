package crud

import (
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
)

// ActionSave 保存数据的主要处理方法（自动判断创建或更新）
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象
func (t *Trait) ActionSave(c *gin.Context) {
	ctx := t.InitCrud(c, "save")
	id, _ := t.ExtractPkId(ctx)

	if !helper.IsEmptyValue(id) {
		t.ActionUpdate(ctx.GinContext)
	} else {
		t.ActionCreate(ctx.GinContext)
	}
}

// SaveFormData 获取保存操作的表单数据
// 参数说明：
//   - ctx *Context: crud上下文对象
//
// 返回值：
//   - modelValue interface{}: 绑定后的模型实例
//   - mapData map[string]any: 原始表单数据映射
//   - err error: 处理过程中的错误信息
func (t *Trait) SaveFormData(ctx *Context) (modelValue interface{}, mapData map[string]any, err error) {
	mapData = ctx.GetSafeMapGPC("all")

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
//   - ctx *Context: crud上下文对象
//   - modelValue interface{}: 要保存的模型数据
//   - mapData map[string]any: 表单数据映射
//   - originalData ...interface{}: 原始数据（仅在更新操作时提供，创建操作时为nil）
//
// 返回值：
//   - modelValue interface{}: 处理后的模型实例
//   - mapData map[string]any: 处理后的表单数据映射
//   - err error: 处理过程中的错误信息
func (t *Trait) SaveBefore(ctx *Context, modelValue interface{}, mapData map[string]any, originalData ...interface{}) (interface{}, map[string]any, error) {
	// 可以在此处添加一些前置处理逻辑
	return modelValue, mapData, nil
}

// SaveAfter 保存后的钩子方法，用于后续处理（在事务内执行）
// 参数说明：
//   - ctx *Context: crud上下文对象
//   - tx *gorm.DB: 数据库事务对象
//   - modelValue interface{}: 已保存的模型实例
//
// 返回值：
//   - error: 处理过程中的错误信息，如果返回错误则会回滚事务
func (t *Trait) SaveAfter(ctx *Context, tx *gorm.DB, modelValue interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return nil
}
