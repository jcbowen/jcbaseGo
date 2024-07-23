package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"gorm.io/gorm"
	"net/http"
	"reflect"
)

func (t *Trait) ActionList(c *gin.Context) {
	t.checkInit(c)

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	showDeletedStr := c.DefaultQuery("show_deleted", "0")
	page := max(helper.Convert{Value: pageStr}.ToInt(), 1)
	pageSize := helper.Convert{Value: pageSizeStr}.ToInt()
	showDeleted := helper.Convert{Value: showDeletedStr}.ToBool()
	if pageSize < 1 {
		pageSize = 10
	} else if pageSize > 1000 {
		pageSize = 1000
	}

	// 构建查询
	query := t.MysqlMain.GetDb().Table(t.ModelTableName)

	if !showDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	query = t.invokeCustomMethod("ListQuery", query).(*gorm.DB)

	// 获取总数
	total := int64(0)
	err := query.Model(reflect.New(reflect.TypeOf(t.Model).Elem()).Interface()).Count(&total).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, jcbaseGo.Result{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	// 动态创建模型实例
	modelType := reflect.TypeOf(t.Model).Elem()
	sliceType := reflect.SliceOf(modelType)
	results := reflect.New(sliceType).Interface()

	err = query.Order(t.invokeCustomMethod("ListOrder")).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, jcbaseGo.Result{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
			Data: nil,
		})
		return
	}

	// 遍历结果到ListEach中
	resultsValue := reflect.ValueOf(results).Elem()
	for i := 0; i < resultsValue.Len(); i++ {
		item := resultsValue.Index(i).Addr().Interface()
		eachResult := t.invokeCustomMethod("ListEach", item)
		if reflect.TypeOf(eachResult).Kind() == reflect.Ptr {
			eachResult = reflect.ValueOf(eachResult).Elem().Interface()
		}
		resultsValue.Index(i).Set(reflect.ValueOf(eachResult))
	}

	// 返回结果
	t.invokeCustomMethod("ListReturn", c, jcbaseGo.ListData{
		List:     results,
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
	})
}

func (t *Trait) ListQuery(query *gorm.DB) *gorm.DB {
	return query
}

func (t *Trait) ListOrder() (order interface{}) {
	return t.PkId + " DESC"
}

func (t *Trait) ListEach(item interface{}) interface{} {
	return item
}

func (t *Trait) ListReturn(c *gin.Context, listData jcbaseGo.ListData) bool {
	c.JSON(http.StatusOK, jcbaseGo.Result{
		Code: 200,
		Msg:  "ok",
		Data: listData,
	})
	return true
}
