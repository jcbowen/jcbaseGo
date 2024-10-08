package crud

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
	"reflect"
)

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

	// 调用自定义的UpdateBefore方法进行前置处理
	callResults = t.callCustomMethod("UpdateBefore", modelValue, mapData)
	modelValue = callResults[0]
	mapData = callResults[1].(map[string]any)
	if callResults[2] != nil {
		err := callResults[2].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 开启事务
	tx := t.MysqlMain.GetDb().Begin()

	// 动态创建模型实例
	modelType := reflect.TypeOf(t.Model).Elem()
	result := reflect.New(modelType).Interface()

	// 查询数据
	query := tx.Table(t.ModelTableName)
	if helper.InArray("deleted_at", t.ModelFields) {
		query = query.Where("deleted_at IS NULL")
	}
	err := query.Where(t.PkId+" = ?", id).First(result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		t.Result(errcode.NotExist, "数据不存在或已被删除")
		return
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
		tx.Rollback()
		t.Result(errcode.DatabaseError, err.Error())
		return
	}

	// 调用自定义的UpdateAfter方法进行后置处理
	callErr := t.callCustomMethod("UpdateAfter", tx, modelValue)[0]
	if callErr != nil {
		err = callErr.(error)
		if err != nil {
			tx.Rollback()
			t.Result(errcode.Unknown, err.Error())
			return
		}
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		t.Result(errcode.DatabaseTransactionCommitError, "事务提交失败，请重试")
		return
	}

	// 返回结果
	t.callCustomMethod("UpdateReturn", modelValue)
}

func (t *Trait) UpdateFormData() (modelValue interface{}, mapData map[string]any, err error) {
	return t.SaveFormData()
}

func (t *Trait) UpdateBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	callResults := t.callCustomMethod("SaveBefore", modelValue, mapData)
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

func (t *Trait) UpdateAfter(tx *gorm.DB, modelValue interface{}) error {
	callResults := t.callCustomMethod("SaveAfter", tx, modelValue)
	var err error
	if callResults[0] != nil {
		err = callResults[0].(error)
	} else {
		err = nil
	}

	return err
}

func (t *Trait) UpdateReturn(item interface{}) bool {
	t.Result(errcode.Success, "ok")
	return true
}
