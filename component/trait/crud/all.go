package crud

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

// ActionAll 获取所有数据的主要处理方法（不分页）
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
func (t *Trait) ActionAll(c *gin.Context) {
	t.InitCrud(c)

	// 获取是否显示已删除数据的参数
	showDeletedStr := c.DefaultQuery("show_deleted", "0")
	showDeleted := helper.Convert{Value: showDeletedStr}.ToBool()

	tableAlias := ""
	if t.ModelTableAlias != "" {
		tableAlias = " " + t.ModelTableAlias
	}

	// 构建查询
	query := t.DBI.GetDb().Table(t.ModelTableName + tableAlias)

	// 应用软删除条件
	if !showDeleted && t.SoftDeleteField != "" && helper.InArray(t.SoftDeleteField, t.ModelFields) {
		query = query.Where(t.TableAlias + t.SoftDeleteField + " " + t.SoftDeleteCondition)
	}

	query = t.callCustomMethod("AllQuery", query)[0].(*gorm.DB)

	// 获取排序
	order := t.callCustomMethod("AllOrder")[0]
	if order != nil {
		query = query.Order(order)
	}

	// 动态创建模型实例
	modelType := reflect.TypeOf(t.Model).Elem()
	sliceType := reflect.SliceOf(modelType)
	results := reflect.New(sliceType).Interface()

	// 批量处理数据
	err := query.FindInBatches(results, 1000, func(tx *gorm.DB, batch int) error {
		// 在这里处理每一批次的数据
		return nil
	}).Error
	if err != nil {
		t.Result(errcode.DatabaseError, "FindInBatches："+err.Error())
		return
	}

	// 遍历处理所有结果
	resultsValue := reflect.ValueOf(results).Elem()
	for i := 0; i < resultsValue.Len(); i++ {
		item := resultsValue.Index(i).Addr().Interface()
		eachResult := t.callCustomMethod("AllEach", item)[0]
		if reflect.TypeOf(eachResult).Kind() == reflect.Ptr {
			eachResult = reflect.ValueOf(eachResult).Elem().Interface()
		}
		resultsValue.Index(i).Set(reflect.ValueOf(eachResult))
	}

	// 返回结果
	t.callCustomMethod("AllReturn", results)
}

// AllQuery 设置获取所有数据的WHERE条件和其他查询参数
// 参数说明：
//   - query *gorm.DB: 数据库查询对象
//
// 返回值：
//   - *gorm.DB: 设置了查询条件的查询对象
func (t *Trait) AllQuery(query *gorm.DB) *gorm.DB {
	return query
}

// AllOrder 设置获取所有数据的排序规则
// 返回值：
//   - interface{}: 排序规则，可以是字符串或其他GORM支持的排序格式
func (t *Trait) AllOrder() interface{} {
	return t.TableAlias + t.PkId + " DESC"
}

// AllEach 对获取的每个数据项进行处理
// 参数说明：
//   - item interface{}: 数据列表中的单个数据项
//
// 返回值：
//   - interface{}: 处理后的数据项
func (t *Trait) AllEach(item interface{}) interface{} {
	return item
}

// AllReturn 获取所有数据成功后的返回处理方法
// 参数说明：
//   - results interface{}: 查询到的所有数据
//
// 返回值：
//   - bool: 处理结果，通常返回true表示成功
func (t *Trait) AllReturn(results interface{}) bool {
	t.Result(errcode.Success, "ok", results)
	return true
}
