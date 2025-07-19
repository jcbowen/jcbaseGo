package crud

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

// ActionUpdate 更新数据的主要处理方法
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
func (t *Trait) ActionUpdate(c *gin.Context) {
	t.InitCrud(c)

	// 获取表单数据
	callResults := t.callCustomMethod("UpdateFormData")
	modelValue := callResults[0]
	mapData := callResults[1].(map[string]any)
	if callResults[2] != nil {
		err := callResults[2].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
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
		t.Result(errcode.ParamError, t.PkId+" 不能为空")
		return
	}

	// 使用GORM的事务方法，自动处理提交和回滚
	err := t.DBI.GetDb().Transaction(func(tx *gorm.DB) error {
		// 动态创建模型实例
		modelType := reflect.TypeOf(t.Model).Elem()
		result := reflect.New(modelType).Interface()

		// 在事务内查询数据，避免并发问题
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

		// 调用自定义的UpdateBefore方法进行前置处理
		callResults = t.callCustomMethod("UpdateBefore", modelValue, mapData, result)
		modelValue = callResults[0]
		mapData = callResults[1].(map[string]any)
		if callResults[2] != nil {
			err := callResults[2].(error)
			if err != nil {
				return err
			}
		}

		// 仅更新传入的字段
		var updateFields []string
		if helper.InArray("updated_at", t.ModelFields) {
			updateFields = append(updateFields, "updated_at")
		}
		for key := range mapData {
			if helper.InArray(key, t.ModelFields) {
				updateFields = append(updateFields, key)
			}
		}

		// 更新数据
		if err = tx.Table(t.ModelTableName).
			Select(updateFields).
			Updates(modelValue).Error; err != nil {
			return err
		}

		// 调用自定义的UpdateAfter方法进行后置处理
		callErr := t.callCustomMethod("UpdateAfter", tx, modelValue, result)[0]
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
		if err.Error() == "数据不存在或已被删除" {
			t.Result(errcode.NotExist, err.Error())
		} else {
			t.Result(errcode.DatabaseError, "更新失败："+err.Error())
		}
		return
	}

	// 返回结果
	t.callCustomMethod("UpdateReturn", modelValue)
}

// UpdateFormData 获取更新操作的表单数据
// 返回值：
//   - modelValue interface{}: 绑定后的模型实例
//   - mapData map[string]any: 原始表单数据映射
//   - err error: 处理过程中的错误信息
func (t *Trait) UpdateFormData() (modelValue interface{}, mapData map[string]any, err error) {
	return t.SaveFormData()
}

// UpdateBefore 更新前的钩子方法，用于数据预处理和验证
// 参数说明：
//   - modelValue interface{}: 要更新的模型实例（包含新数据）
//   - mapData map[string]any: 表单数据映射
//   - originalData interface{}: 数据库中的原始数据
//
// 返回值：
//   - interface{}: 处理后的模型实例
//   - map[string]any: 处理后的表单数据映射
//   - error: 处理过程中的错误信息
func (t *Trait) UpdateBefore(modelValue interface{}, mapData map[string]any, originalData interface{}) (interface{}, map[string]any, error) {
	callResults := t.callCustomMethod("SaveBefore", modelValue, mapData, originalData)
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

// UpdateAfter 更新后的钩子方法，用于后续处理（在事务内执行）
// 参数说明：
//   - tx *gorm.DB: 数据库事务对象
//   - modelValue interface{}: 已更新的模型实例
//   - originalData interface{}: 数据库中的原始数据
//
// 返回值：
//   - error: 处理过程中的错误信息，如果返回错误则会回滚事务
func (t *Trait) UpdateAfter(tx *gorm.DB, modelValue interface{}, originalData interface{}) error {
	callResults := t.callCustomMethod("SaveAfter", tx, modelValue)
	var err error
	if callResults[0] != nil {
		err = callResults[0].(error)
	} else {
		err = nil
	}

	return err
}

// UpdateReturn 更新成功后的返回处理方法
// 参数说明：
//   - item interface{}: 更新成功的数据项
//
// 返回值：
//   - bool: 处理结果，通常返回true表示成功
func (t *Trait) UpdateReturn(item interface{}) bool {
	t.Result(errcode.Success, "ok")
	return true
}
