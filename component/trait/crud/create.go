package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
	"reflect"
)

func (t *Trait) ActionCreate(c *gin.Context) {

	t.InitCrud(c)

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

	// 开启事务
	tx := t.MysqlMain.GetDb().Begin()

	// 插入数据
	if err = tx.Create(modelValue).Error; err != nil {
		tx.Rollback()
		t.Result(errcode.DatabaseError, "ok")
		return
	}

	// 调用自定义的CreateAfter方法进行后置处理
	callErr := t.callCustomMethod("CreateAfter", tx, modelValue)[0]
	if callErr != nil {
		err = callErr.(error)
		if err != nil {
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

func (t *Trait) CreateFormData() (modelValue interface{}, mapData map[string]any, err error) {
	return t.SaveFormData()
}

func (t *Trait) CreateBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	// 可以在此处添加一些前置处理逻辑
	return t.SaveBefore(modelValue, mapData)
}

func (t *Trait) CreateAfter(tx *gorm.DB, modelValue interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return t.AfterSave(tx, modelValue)
}

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
