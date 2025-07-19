package crud

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

// ActionDetail 获取数据详情的主要处理方法
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
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

	// 为了方便，直接传query进去拼接就好
	query = t.callCustomMethod("DetailSelect", query)[0].(*gorm.DB)

	// 应用软删除条件
	if !showDeleted && t.SoftDeleteField != "" && helper.InArray(t.SoftDeleteField, t.ModelFields) {
		query = query.Where(t.TableAlias + t.SoftDeleteField + " " + t.SoftDeleteCondition)
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

// DetailFormData 获取详情查询的参数数据
// 返回值：
//   - map[string]any: 请求参数映射
//   - error: 处理过程中的错误信息
func (t *Trait) DetailFormData() (map[string]any, error) {
	// 获取安全过滤后的请求参数
	mapData := t.GetSafeMapGPC("all")

	return mapData, nil
}

// DetailSelect 设置详情查询的SELECT字段
// 参数说明：
//   - query *gorm.DB: 数据库查询对象
//
// 返回值：
//   - *gorm.DB: 设置了SELECT字段的查询对象
func (t *Trait) DetailSelect(query *gorm.DB) *gorm.DB {
	// 默认就是查询*，所以这里就没必要单独写query.Select("*")了
	return query
}

// DetailQuery 设置详情查询的WHERE条件和其他查询参数
// 参数说明：
//   - query *gorm.DB: 数据库查询对象
//   - mapData map[string]any: 请求参数映射
//
// 返回值：
//   - *gorm.DB: 设置了查询条件的查询对象
func (t *Trait) DetailQuery(query *gorm.DB, mapData map[string]any) *gorm.DB {
	return query
}

// Detail 对详情数据进行处理
// 参数说明：
//   - item interface{}: 查询到的详情数据
//
// 返回值：
//   - interface{}: 处理后的详情数据
func (t *Trait) Detail(item interface{}) interface{} {
	return item
}

// DetailReturn 详情查询成功后的返回处理方法
// 参数说明：
//   - detail interface{}: 详情数据
//
// 返回值：
//   - bool: 处理结果，通常返回true表示成功
func (t *Trait) DetailReturn(detail interface{}) bool {
	t.Result(errcode.Success, "ok", detail)
	return true
}
