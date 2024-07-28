package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"reflect"
)

func (t *Trait) ActionCreate(c *gin.Context) {
	t.checkInit(c)

	// 获取表单数据
	callResults := t.callCustomMethod("GetCreateFormData")
	modelValue := callResults[0]
	mapData := callResults[1].(map[string]any)
	if callResults[2] != nil {
		err := callResults[2].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 调用自定义的BeforeCreate方法进行前置处理
	modelValue = t.callCustomMethod("BeforeCreate", modelValue, mapData)[0]

	// 开启事务
	tx := t.MysqlMain.GetDb().Begin()

	// 插入数据
	if err := tx.Create(modelValue).Error; err != nil {
		tx.Rollback()
		t.Result(errcode.DatabaseError, "ok")
		return
	}

	// 调用自定义的AfterCreate方法进行后置处理
	callErr := t.callCustomMethod("AfterCreate", modelValue)[0]
	if callErr != nil {
		err, ok := callErr.(error)
		if ok && err != nil {
			tx.Rollback()
			t.Result(errcode.Unknown, err.Error())
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Result(errcode.DatabaseTransactionCommitError, "事务提交失败，请重试")
		return
	}

	// 返回结果
	t.callCustomMethod("CreateReturn", modelValue)
}

func (t *Trait) GetCreateFormData() (modelValue interface{}, mapData map[string]any, err error) {
	return t.GetSaveFormData()
}

func (t *Trait) BeforeCreate(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	// 可以在此处添加一些前置处理逻辑
	return t.BeforeSave(modelValue, mapData)
}

func (t *Trait) AfterCreate(modelValue interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return t.AfterSave(modelValue)
}

func (t *Trait) CreateReturn(item any) bool {
	var (
		mapItem map[string]any
		pkId    uint
	)

	if reflect.TypeOf(item).Kind() == reflect.Ptr {
		item = reflect.ValueOf(item).Elem()
	}

	switch reflect.TypeOf(item).Kind() {
	case reflect.Struct:
		helper.JsonStruct(item).ToMap(&mapItem)
	default:
		mapItem = item.(map[string]any)
	}
	pkId = helper.Convert{Value: mapItem[t.PkId]}.ToUint()

	t.Result(errcode.Success, "ok", gin.H{
		t.PkId: pkId,
	})

	return true
}
