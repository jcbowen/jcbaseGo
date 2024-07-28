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
	t.checkInit(c)

	// 获取表单数据
	callResults := t.callCustomMethod("GetUpdateFormData")
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

	// 调用自定义的BeforeUpdate方法进行前置处理
	callResults = t.callCustomMethod("BeforeUpdate", modelValue, mapData)
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

	// 调用自定义的AfterUpdate方法进行后置处理
	callErr := t.callCustomMethod("AfterUpdate", modelValue)[0]
	if callErr != nil {
		err, ok := callErr.(error)
		if ok && err != nil {
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

func (t *Trait) GetUpdateFormData() (modelValue interface{}, mapData map[string]any, err error) {
	return t.GetSaveFormData()
}

func (t *Trait) BeforeUpdate(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	// 可以在此处添加一些前置处理逻辑
	return t.BeforeSave(modelValue, mapData)
}

func (t *Trait) AfterUpdate(modelValue interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return t.AfterSave(modelValue)
}

func (t *Trait) UpdateReturn(item interface{}) bool {
	t.Result(errcode.Success, "ok")
	return true
}
