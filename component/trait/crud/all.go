package crud

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

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
	if !showDeleted && helper.InArray("deleted_at", t.ModelFields) {
		query = query.Where(t.TableAlias + "deleted_at " + t.SoftDeleteCondition)
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

func (t *Trait) AllQuery(query *gorm.DB) *gorm.DB {
	return query
}

func (t *Trait) AllOrder() interface{} {
	return t.TableAlias + t.PkId + " DESC"
}

func (t *Trait) AllEach(item interface{}) interface{} {
	return item
}

func (t *Trait) AllReturn(results interface{}) bool {
	t.Result(errcode.Success, "ok", results)
	return true
}
