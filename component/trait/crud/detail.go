package crud

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
	"net/http"
	"reflect"
)

func (t *Trait) ActionDetail(c *gin.Context) {
	t.checkInit(c)

	idStr := c.DefaultQuery(t.PkId, "0")
	showDeletedStr := c.DefaultQuery("show_deleted", "0")
	id := helper.Convert{Value: idStr}.ToInt()
	showDeleted := helper.Convert{Value: showDeletedStr}.ToBool()

	if helper.IsEmptyValue(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	// 构建查询
	query := t.MysqlMain.GetDb().Table(t.ModelTableName)

	if !showDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	query = t.invokeCustomMethod("DetailQuery", query).(*gorm.DB)

	// 动态创建模型实例
	modelType := reflect.TypeOf(t.Model).Elem()
	result := reflect.New(modelType).Interface()

	err := query.Where(t.PkId+" = ?", id).First(result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, jcbaseGo.Result{
				Code: errcode.NotExist,
				Msg:  "数据不存在或已被删除",
				Data: nil,
			})
		} else {
			c.JSON(http.StatusInternalServerError, jcbaseGo.Result{
				Code: errcode.DatabaseError,
				Msg:  err.Error(),
				Data: nil,
			})
		}
		return
	}

	// 调用自定义的Detail方法处理结果
	result = t.invokeCustomMethod("Detail", result)

	// 返回结果
	t.invokeCustomMethod("DetailReturn", c, result)
}

func (t *Trait) DetailQuery(query *gorm.DB) *gorm.DB {
	return query
}

func (t *Trait) Detail(item interface{}) interface{} {
	return item
}

func (t *Trait) DetailReturn(c *gin.Context, detail interface{}) bool {
	c.JSON(http.StatusOK, jcbaseGo.Result{
		Code: errcode.Success,
		Msg:  "ok",
		Data: gin.H{
			"detail": detail,
		},
	})
	return true
}
