package crud

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

func (t *Trait) ActionDetail(c *gin.Context) {
	t.InitCrud(c)

	// detail参数获取回调
	callResults := t.callCustomMethod("DetailFormData")
	mapData := callResults[0].(map[string]any)
	if callResults[1] != nil {
		err := callResults[1].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 获取泛类型参数参数
	idAny, ok := mapData[t.PkId]
	if !ok {
		t.Result(errcode.ParamError, t.PkId+" 不能为空")
	}
	showDeletedAny, ok := mapData["show_deleted"]
	if !ok {
		showDeletedAny = "0"
	}

	// 转换参数为正确的类型
	id := helper.Convert{Value: idAny}.ToUint()
	showDeleted := helper.Convert{Value: showDeletedAny}.ToBool()

	// 判断必要参数是否为空
	if helper.IsEmptyValue(id) {
		t.Result(errcode.ParamError, t.PkId+" 不能为空")
		return
	}

	tableAlias := ""
	if t.ModelTableAlias != "" {
		tableAlias = " " + t.ModelTableAlias
	}

	// 构建查询
	query := t.DBI.GetDb().Table(t.ModelTableName + tableAlias)

	if !showDeleted && helper.InArray("deleted_at", t.ModelFields) {
		query = query.Where(t.TableAlias + "deleted_at IS NULL")
	}

	query = t.callCustomMethod("DetailQuery", query, mapData)[0].(*gorm.DB)

	// 动态创建模型实例
	if t.DetailResultStruct == nil {
		t.DetailResultStruct = t.Model
	}

	resultStructType := reflect.TypeOf(t.DetailResultStruct)
	if resultStructType.Kind() == reflect.Ptr {
		resultStructType = resultStructType.Elem()
	}

	if resultStructType.Kind() != reflect.Struct {
		t.DetailResultStruct = t.Model
		resultStructType = reflect.TypeOf(t.DetailResultStruct)
		if resultStructType.Kind() == reflect.Ptr {
			resultStructType = resultStructType.Elem()
		}
	}

	modelType := resultStructType
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	result := reflect.New(modelType).Interface()

	err := query.Where(t.TableAlias+t.PkId+" = ?", id).First(result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t.Result(errcode.NotExist, "数据不存在或已被删除")
		} else {
			t.Result(errcode.DatabaseError, err.Error())
		}
		return
	}

	// 调用自定义的Detail方法处理结果
	result = t.callCustomMethod("Detail", result)[0]

	// 返回结果
	t.callCustomMethod("DetailReturn", result)
}

func (t *Trait) DetailFormData() (map[string]any, error) {
	// 获取安全过滤后的请求参数
	mapData := t.GetSafeMapGPC("all")

	return mapData, nil
}

func (t *Trait) DetailQuery(query *gorm.DB, mapData map[string]any) *gorm.DB {
	return query
}

func (t *Trait) Detail(item interface{}) interface{} {
	return item
}

func (t *Trait) DetailReturn(detail interface{}) bool {
	t.Result(errcode.Success, "ok", detail)
	return true
}
