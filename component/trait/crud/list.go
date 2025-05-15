package crud

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

func (t *Trait) ActionList(c *gin.Context) {
	t.InitCrud(c)

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

	tableAlias := ""
	if t.ModelTableAlias != "" {
		tableAlias = " " + t.ModelTableAlias
	}

	// 构建查询
	query := t.DBI.GetDb().Table(t.ModelTableName + tableAlias)

	if !showDeleted && helper.InArray("deleted_at", t.ModelFields) {
		query = query.Where(t.TableAlias + "deleted_at IS NULL")
	}

	callResults := t.callCustomMethod("ListQuery", query)
	query = callResults[0].(*gorm.DB)
	if callResults[1] != nil {
		err := callResults[1].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 获取总数
	total := int64(0)
	err := query.
		Model(reflect.New(reflect.TypeOf(t.Model).Elem()).Interface()).
		Count(&total).Error
	if err != nil {
		t.Result(http.StatusInternalServerError, err.Error())
		return
	}

	// Select不能在Count前，否则会报错
	// 为了方便，直接传query进去拼接就好
	query = t.callCustomMethod("ListSelect", query)[0].(*gorm.DB)

	// 动态创建模型实例
	if t.ListResultStruct == nil {
		t.ListResultStruct = t.Model
	}

	resultStructType := reflect.TypeOf(t.ListResultStruct)
	if resultStructType.Kind() == reflect.Ptr {
		resultStructType = resultStructType.Elem()
	}

	if resultStructType.Kind() != reflect.Struct {
		t.ListResultStruct = t.Model
		resultStructType = reflect.TypeOf(t.ListResultStruct)
		if resultStructType.Kind() == reflect.Ptr {
			resultStructType = resultStructType.Elem()
		}
	}
	sliceType := reflect.SliceOf(resultStructType)
	results := reflect.New(sliceType).Interface()

	err = query.Order(t.callCustomMethod("ListOrder")[0]).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(results).Error
	if err != nil {
		t.Result(http.StatusInternalServerError, err.Error())
		return
	}

	// 遍历结果到ListEach中
	resultsValue := reflect.ValueOf(results).Elem()
	// 创建新的切片来存储修改后的结果
	modifiedResults := make([]interface{}, resultsValue.Len())
	for i := 0; i < resultsValue.Len(); i++ {
		item := resultsValue.Index(i).Addr().Interface()
		eachResult := t.callCustomMethod("ListEach", item)[0]

		// 处理返回结果
		if reflect.TypeOf(eachResult).Kind() == reflect.Map {
			// 如果是 map 类型，直接使用
			modifiedResults[i] = eachResult
		} else if reflect.TypeOf(eachResult).Kind() == reflect.Ptr {
			// 如果是指针类型，获取其值
			modifiedResults[i] = reflect.ValueOf(eachResult).Elem().Interface()
		} else {
			// 其他类型直接使用
			modifiedResults[i] = eachResult
		}
	}

	// 返回结果
	t.callCustomMethod("ListReturn", jcbaseGo.ListData{
		List:     modifiedResults, // 使用新的切片
		Total:    int(total),
		Page:     page,
		PageSize: pageSize,
	})
}

func (t *Trait) ListSelect(query *gorm.DB) *gorm.DB {
	// 默认就是查询*，所以这里就没必要单独写query.Select("*")了
	return query
}

func (t *Trait) ListQuery(query *gorm.DB) (*gorm.DB, error) {
	return query, nil
}

func (t *Trait) ListOrder() (order interface{}) {
	return t.TableAlias + t.PkId + " DESC"
}

func (t *Trait) ListEach(item interface{}) interface{} {
	return item
}

func (t *Trait) ListReturn(listData jcbaseGo.ListData) bool {
	t.Result(200, "ok", listData)
	return true
}
