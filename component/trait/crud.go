package trait

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

type CRUDTrait struct {
	Model          any // 模型指针
	MysqlMain      *mysql.Instance
	PkId           string `default:"id"` // 数据表主键
	ModelTableName string // 模型表名
	OperateTime    string // 操作时间
}

func (t *CRUDTrait) ActionList(c *gin.Context) {
	t.checkInit()

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 获取排序参数
	sortBy := c.DefaultQuery("sort_by", t.PkId)
	order := c.DefaultQuery("order", "asc")

	// 构建查询
	query := t.MysqlMain.GetDb().Table(t.ModelTableName)

	// 处理回调函数
	query = t.ListRow(query)

	var results []map[string]interface{}
	err := query.Select("*").Order(sortBy + " " + order).Offset((page - 1) * pageSize).Limit(pageSize).Find(&results).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	total := int64(0)
	err = query.Count(&total).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"data":      results,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

func (t *CRUDTrait) ListRow(query *gorm.DB) *gorm.DB {
	// 可以在这里添加用户自定义的查询逻辑
	return query
}

// ----- 私有方法 ----- /

func (t *CRUDTrait) checkInit() {
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
