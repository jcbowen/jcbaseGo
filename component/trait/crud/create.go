package crud

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

// ActionCreate 创建数据的主要处理方法
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
func (t *Trait) ActionCreate(c *gin.Context) {
	t.InitCrud(c, "create")

	var err error

	// 获取表单数据
	callResults := t.callCustomMethod("CreateFormData")
	modelValue := callResults[0]
	mapData := callResults[1].(map[string]any)
	if callResults[2] != nil {
		err = callResults[2].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 调用自定义的CreateBefore方法进行前置处理
	callResults = t.callCustomMethod("CreateBefore", modelValue, mapData)
	modelValue = callResults[0]
	mapData = callResults[1].(map[string]any)
	if callResults[2] != nil {
		err = callResults[2].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 使用GORM的事务方法，自动处理提交和回滚
	err = t.DBI.GetDb().Transaction(func(tx *gorm.DB) error {
		// 插入数据
		if err = tx.Create(modelValue).Error; err != nil {
			return err
		}

		// 调用自定义的CreateAfter方法进行后置处理
		callErr := t.callCustomMethod("CreateAfter", tx, modelValue)[0]
		if callErr != nil {
			err = callErr.(error)
			if err != nil {
				return err
			}
		}

		return nil
	})

	// 处理事务结果
	if err != nil {
		t.Result(errcode.DatabaseError, "创建失败："+err.Error())
		return
	}

	// 返回结果
	t.callCustomMethod("CreateReturn", modelValue)
}

// CreateFormData 获取创建操作的表单数据
// 返回值：
//   - modelValue interface{}: 绑定后的模型实例
//   - mapData map[string]any: 原始表单数据映射
//   - err error: 处理过程中的错误信息
func (t *Trait) CreateFormData() (modelValue interface{}, mapData map[string]any, err error) {
	return t.SaveFormData()
}

// CreateBefore 创建前的钩子方法，用于数据预处理和验证
// 参数说明：
//   - modelValue interface{}: 要创建的模型实例
//   - mapData map[string]any: 表单数据映射
//
// 返回值：
//   - interface{}: 处理后的模型实例
//   - map[string]any: 处理后的表单数据映射
//   - error: 处理过程中的错误信息
func (t *Trait) CreateBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	callResults := t.callCustomMethod("SaveBefore", modelValue, mapData, nil)
	modelValue = callResults[0]
	mapData = callResults[1].(map[string]any)
	var err error
	if callResults[2] != nil {
		err = callResults[2].(error)
	} else {
		err = nil
	}

	return modelValue, mapData, err
}

// CreateAfter 创建后的钩子方法，用于后续处理（在事务内执行）
// 参数说明：
//   - tx *gorm.DB: 数据库事务对象
//   - modelValue interface{}: 已创建的模型实例
//
// 返回值：
//   - error: 处理过程中的错误信息，如果返回错误则会回滚事务
func (t *Trait) CreateAfter(tx *gorm.DB, modelValue interface{}) error {
	callResults := t.callCustomMethod("SaveAfter", tx, modelValue)
	var err error
	if callResults[0] != nil {
		err = callResults[0].(error)
	} else {
		err = nil
	}

	return err
}

// CreateReturn 创建成功后的返回处理方法
// 参数说明：
//   - item any: 创建成功的数据项
//
// 返回值：
//   - bool: 处理结果，通常返回true表示成功
func (t *Trait) CreateReturn(item any) bool {
	var (
		mapItem map[string]any
		pkId    uint
	)

	// 判断是否为指针
	if reflect.TypeOf(item).Kind() == reflect.Ptr {
		item = reflect.ValueOf(item).Elem().Interface()
	}

	// 将 item 转换为 map，方便取值
	switch reflect.TypeOf(item).Kind() {
	case reflect.Struct:
		helper.Json(item).ToMap(&mapItem)
	default:
		mapItem = item.(map[string]any)
	}

	// 获取主键
	pkIdAny, _ := mapItem[t.PkId]
	pkId = helper.Convert{Value: pkIdAny}.ToUint()

	t.Result(errcode.Success, "ok", gin.H{
		t.PkId: pkId,
	})

	return true
}
