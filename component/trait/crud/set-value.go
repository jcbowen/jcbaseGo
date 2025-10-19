package crud

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

// ActionSetValue 设置单个字段值的主要处理方法
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
func (t *Trait) ActionSetValue(c *gin.Context) {
	ctx := t.InitCrud(c, "set-value")

	// 获取表单数据
	callResults := t.callCustomMethod("SetValueFormData", ctx)
	modelValue := callResults[0]
	mapData := callResults[1].(map[string]any)
	if callResults[2] != nil {
		err := callResults[2].(error)
		if err != nil {
			ctx.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 获取主键ID
	idStr, exists := mapData[t.PkId]
	var id uint
	if exists {
		id = helper.Convert{Value: idStr}.ToUint()
	}
	if helper.IsEmptyValue(id) {
		ctx.Result(errcode.ParamError, t.PkId+" 不能为空")
		return
	}

	// 获取字段名、类型和值
	field, fieldExists := mapData["field"].(string)
	if !fieldExists || field == "" {
		ctx.Result(errcode.ParamError, "字段名不能为空")
		return
	}

	fieldType, typeExists := mapData["type"].(string)
	if !typeExists || fieldType == "" {
		ctx.Result(errcode.ParamError, "字段类型不能为空")
		return
	}

	value, valueExists := mapData["value"]
	if !valueExists {
		ctx.Result(errcode.ParamError, "字段值不能为空")
		return
	}

	// 检查字段名的有效性
	callResults = t.callCustomMethod("SetValueCheckField", ctx, field)
	if callResults[0] != nil {
		err := callResults[0].(error)
		if err != nil {
			ctx.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 根据字段值的类型，对字段值进行格式化
	switch fieldType {
	case "int":
		value = helper.Convert{Value: value}.ToInt()
	case "float":
		value = helper.Convert{Value: value}.ToFloat64()
	case "money":
		moneyNum, _ := helper.Convert{Value: value}.ToNumber()
		value = (&helper.MoneyHelper{}).SetAmount(moneyNum).FloatString()
	default:
		if reflect.TypeOf(value).Kind() == reflect.Slice || reflect.TypeOf(value).Kind() == reflect.Map {
			var jsonString string
			helper.Json(value).ToString(&jsonString)
			value = jsonString
		} else {
			value = helper.Convert{Value: value}.ToString()
		}
	}

	updatedField := reflect.ValueOf(modelValue).Elem().FieldByName(field)
	if updatedField.IsValid() && updatedField.CanSet() {
		updatedField.Set(reflect.ValueOf(value))
	}

	// 使用GORM的事务方法，自动处理提交和回滚
	err := t.DBI.GetDb().Transaction(func(tx *gorm.DB) error {
		// 查询数据
		modelType := reflect.TypeOf(t.Model).Elem()
		result := reflect.New(modelType).Interface()
		query := tx.Table(t.ModelTableName)
		// 应用软删除条件
		if t.SoftDeleteField != "" && helper.InArray(t.SoftDeleteField, t.ModelFields) {
			query = query.Where(t.SoftDeleteField + " " + t.SoftDeleteCondition)
		}
		err := query.Where(t.PkId+" = ?", id).First(result).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("数据不存在或已被删除")
		}
		if err != nil {
			return err
		}

		// 调用 SetValueBefore 钩子进行前置处理（在事务内，查询到原始数据后）
		callResults = t.callCustomMethod("SetValueBefore", ctx, modelValue, mapData, result)
		modelValue = callResults[0]
		mapData = callResults[1].(map[string]any)
		if callResults[2] != nil {
			err := callResults[2].(error)
			if err != nil {
				return err
			}
		}

		// 检查值是否发生变化
		modelValueField := reflect.ValueOf(result).Elem().FieldByName(field)
		if modelValueField.IsValid() && modelValueField.Interface() == value {
			return errors.New("值未发生改变，请确认修改内容")
		}

		// 更新数据
		updateData := map[string]interface{}{field: value}
		if err = tx.Table(t.ModelTableName).Where(t.PkId+" = ?", id).Updates(updateData).Error; err != nil {
			return err
		}

		// 调用自定义的SetValueAfter方法进行后置处理
		callResults = t.callCustomMethod("SetValueAfter", ctx, tx, modelValue)
		if callResults[0] != nil {
			err, ok := callResults[0].(error)
			if ok && err != nil {
				return err
			}
		}

		return nil
	})

	// 处理事务结果
	if err != nil {
		if err.Error() == "数据不存在或已被删除" {
			ctx.Result(errcode.NotExist, err.Error())
		} else if err.Error() == "值未发生改变，请确认修改内容" {
			ctx.Result(errcode.Success, err.Error())
		} else {
			ctx.Result(errcode.DatabaseError, "设置失败："+err.Error())
		}
		return
	}

	// 返回结果
	t.callCustomMethod("SetValueReturn", ctx, value, field, id)
}

// SetValueFormData 获取设置字段值操作的表单数据
// 参数说明：
//   - ctx *Context: crud上下文对象
//
// 返回值：
//   - modelValue interface{}: 绑定后的模型实例
//   - mapData map[string]any: 原始表单数据映射
//   - err error: 处理过程中的错误信息
func (t *Trait) SetValueFormData(ctx *Context) (modelValue interface{}, mapData map[string]any, err error) {
	return t.SaveFormData(ctx)
}

// SetValueCheckField 验证传入的字段名是否有效
// 参数说明：
//   - ctx *Context: crud上下文对象
//   - field string: 要验证的字段名
//
// 返回值：
//   - error: 验证失败时的错误信息
func (t *Trait) SetValueCheckField(ctx *Context, field string) error {
	if !helper.InArray(field, t.ModelFields) {
		return errors.New("参数错误，请传入有效的字段名")
	}
	return nil
}

// SetValueBefore 设置字段值前的钩子方法，用于数据预处理和验证
// 参数说明：
//   - ctx *Context: crud上下文对象
//   - modelValue interface{}: 表单数据绑定的模型实例
//   - mapData map[string]any: 表单数据映射
//   - originalData interface{}: 数据库中的数据
//
// 返回值：
//   - interface{}: 处理后的模型实例
//   - map[string]any: 处理后的表单数据映射
//   - error: 处理过程中的错误信息
func (t *Trait) SetValueBefore(ctx *Context, modelValue interface{}, mapData map[string]any, originalData interface{}) (interface{}, map[string]any, error) {
	// 调用通用的 SaveBefore 钩子方法，传入原始数据
	callResults := t.callCustomMethod("SaveBefore", ctx, modelValue, mapData, originalData)
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

// SetValueAfter 设置字段值后的钩子方法，用于后续处理（在事务内执行）
// 参数说明：
//   - ctx *Context: crud上下文对象
//   - tx *gorm.DB: 数据库事务对象
//   - modelValue interface{}: 包含设置字段值的模型实例(与其他钩子不同的是，其他钩子的modelValue是处理后的模型实例，而set-value的modelValue是表单数据绑定的模型实例，因为set-value是直接更新指定字段)
//
// 返回值：
//   - error: 处理过程中的错误信息，如果返回错误则会回滚事务
func (t *Trait) SetValueAfter(ctx *Context, tx *gorm.DB, modelValue interface{}) error {
	callResults := t.callCustomMethod("SaveAfter", ctx, tx, modelValue)
	var err error
	if callResults[0] != nil {
		err = callResults[0].(error)
	} else {
		err = nil
	}

	return err
}

// SetValueReturn 设置字段值成功后的返回处理方法
// 参数说明：
//   - ctx *Context: crud上下文对象
//   - value interface{}: 设置的值
//   - field string: 设置的字段名
//   - id uint: 被设置的记录ID
//
// 返回值：
//   - bool: 处理结果，通常返回true表示成功
func (t *Trait) SetValueReturn(ctx *Context, value interface{}, field string, id uint) bool {
	ctx.Result(errcode.Success, "设置成功", gin.H{
		"id":    id,
		"field": field,
		"value": value,
	})
	return true
}
