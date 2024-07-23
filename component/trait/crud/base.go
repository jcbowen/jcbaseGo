package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/mysql"
	"log"
	"reflect"
	"time"
)

type Trait struct {
	PkId           string `default:"id"` // 数据表主键
	ModelTableName string // 模型表名
	OperateTime    string // 操作时间

	Model      any             // 模型指针
	MysqlMain  *mysql.Instance // 数据库实例
	Controller interface{}     // 控制器
	GinContext *gin.Context    // 请求上下文
}

// 初始化
func (t *Trait) checkInit(c *gin.Context) {
	_ = helper.CheckAndSetDefault(t)
	t.GinContext = c

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
func (t *Trait) invokeCustomMethod(methodName string, args ...interface{}) interface{} {
	// 调用自定义方法
	method := reflect.ValueOf(t.Controller).MethodByName(methodName)
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
