package trait

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"reflect"
	"time"
)

type CRUDTrait struct {
	Model          any // 模型指针
	MysqlMain      *mysql.Instance
	PkId           string      `default:"id"` // 数据表主键
	ModelTableName string      // 模型表名
	OperateTime    string      // 操作时间
	Custom         interface{} // 自定义控制器
}

func (t *CRUDTrait) ActionList(c *gin.Context) {
	t.checkInit()

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (t *CRUDTrait) ListQuery(query *gorm.DB) *gorm.DB {
	return query
}

func (t *CRUDTrait) ListOrder() (order interface{}) {
	return t.PkId + " DESC"
}

func (t *CRUDTrait) ListEach(item interface{}) interface{} {
	return item
}

func (t *CRUDTrait) ListReturn(c *gin.Context, listData jcbaseGo.ListData) bool {
	c.JSON(http.StatusOK, jcbaseGo.Result{
		Code: 200,
		Msg:  "ok",
		Data: listData,
	})
	return true
}

// ----- 私有方法 ----- /

// 初始化
func (t *CRUDTrait) checkInit() {
	_ = helper.CheckAndSetDefault(t)

	// 判断模型是否为空
	if t.Model == nil {
		log.Panic("模型不能为空")
	}

	modelValue := reflect.ValueOf(t.Model)
	modelType := reflect.TypeOf(t.Model)

	// 检查是否传入的是指针
	if modelType.Kind() == reflect.Ptr {
		modelValue = modelValue.Elem()
		modelType = modelType.Elem()
	}

	// 确保获取到具体模型的名称
	model := reflect.New(modelType).Interface()

	// 获取表名
	if tableNameProvider, ok := model.(interface{ GetTableName(modelName string) string }); ok {
		t.ModelTableName = tableNameProvider.GetTableName(modelType.Name())
	} else {
		log.Panic("模型未实现 GetTableName 方法")
	}

	// 设置操作时间
	t.OperateTime = time.Now().Format("2006-01-02 15:04:05")
}

// 调用自定义方法，如果方法不存在则调用默认方法
func (t *CRUDTrait) invokeCustomMethod(methodName string, args ...interface{}) interface{} {
	// 调用自定义方法
	method := reflect.ValueOf(t.Custom).MethodByName(methodName)
	if method.IsValid() {
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			in[i] = reflect.ValueOf(arg)
		}
		results := method.Call(in)
		if len(results) > 0 {
			return results[0].Interface()
		}
		return nil
	}

	// 调用默认方法
	defaultMethod := reflect.ValueOf(t).MethodByName(methodName)
	if !defaultMethod.IsValid() {
		log.Panic("默认方法不存在：" + methodName)
	}
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}
	results := defaultMethod.Call(in)
	if len(results) > 0 {
		return results[0].Interface()
	}
	return nil
}
