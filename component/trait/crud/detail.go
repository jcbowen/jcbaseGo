package crud

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
	"reflect"
)

func (t *Trait) ActionDetail(c *gin.Context) {
	t.checkInit(c)

	id, _ := t.ExtractPkId()
	showDeletedStr := c.DefaultQuery("show_deleted", "0")
	showDeleted := helper.Convert{Value: showDeletedStr}.ToBool()

	if helper.IsEmptyValue(id) {
		t.Result(errcode.ParamError, t.PkId+" 不能为空")
		return
	}

	tableAlias := ""
	if t.ModelTableAlias != "" {
		tableAlias = " " + t.ModelTableAlias
	}

	// 构建查询
	query := t.MysqlMain.GetDb().Table(t.ModelTableName + tableAlias)

	if !showDeleted && helper.InArray("deleted_at", t.ModelFields) {
		query = query.Where(t.TableAlias + "deleted_at IS NULL")
	}

	query = t.callCustomMethod("DetailQuery", query)[0].(*gorm.DB)

	// 动态创建模型实例
	modelType := reflect.TypeOf(t.Model).Elem()
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

func (t *Trait) DetailQuery(query *gorm.DB) *gorm.DB {
	return query
}

func (t *Trait) Detail(item interface{}) interface{} {
	return item
}

func (t *Trait) DetailReturn(detail interface{}) bool {
	t.Result(errcode.Success, "ok", detail)
	return true
}
