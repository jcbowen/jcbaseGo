package trait

import (
	"github.com/gin-gonic/gin"
	"log"
	"reflect"
	"time"
)

type CRUDTrait struct {
	PkId           string `default:"id"` // 数据表主键
	Model          any    // 模型指针
	ModelTableName string // 模型表名
	OperateTime    string // 操作时间
}

func (t *CRUDTrait) ActionList(c *gin.Context) {
	t.checkInit(c)
	log.Println(t.ModelTableName)
}

// ----- 私有方法 ----- /

func (t *CRUDTrait) checkInit(c *gin.Context) {
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
