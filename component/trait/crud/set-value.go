package crud

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

func (t *Trait) ActionSetValue(c *gin.Context) {
	t.InitCrud(c)

	// 获取表单数据
	callResults := t.callCustomMethod("SetValueFormData")
	mapData := callResults[1].(map[string]any)
	if callResults[2] != nil {
		err := callResults[2].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 获取主键ID
	idStr, exists := mapData[t.PkId]
	var id uint
	if exists {
		id = helper.Convert{Value: idStr}.ToUint()
	}
	if helper.IsEmptyValue(id) {
		t.Result(errcode.ParamError, t.PkId+" 不能为空")
		return
	}

	// 获取字段名、类型和值
	field, fieldExists := mapData["field"].(string)
	if !fieldExists || field == "" {
		t.Result(errcode.ParamError, "字段名不能为空")
		return
	}

	fieldType, typeExists := mapData["type"].(string)
	if !typeExists || fieldType == "" {
		t.Result(errcode.ParamError, "字段类型不能为空")
		return
	}

	value, valueExists := mapData["value"]
	if !valueExists {
		t.Result(errcode.ParamError, "字段值不能为空")
		return
	}

	// 检查字段名的有效性
	callResults = t.callCustomMethod("SetValueCheckField", field)
	if callResults[0] != nil {
		err := callResults[0].(error)
		if err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}

	// 根据字段值的类型，对字段值进行格式化
	switch fieldType {
	case "int":
		value = helper.Convert{Value: value}.ToInt()
	case "float":
		value = helper.Convert{Value: value}.ToFloat64()
	case "money":
		moneyNum, _ := helper.Convert{Value: value}.ToNumber()
		value = (&helper.MoneyHelper{}).SetAmount(moneyNum).FloatString()
	default:
		if reflect.TypeOf(value).Kind() == reflect.Slice || reflect.TypeOf(value).Kind() == reflect.Map {
			var jsonString string
			helper.Json(value).ToString(&jsonString)
			value = jsonString
		} else {
			value = helper.Convert{Value: value}.ToString()
		}
	}

	// 开始事务
	tx := t.DBI.GetDb().Begin()

	// 查询数据
	modelType := reflect.TypeOf(t.Model).Elem()
	result := reflect.New(modelType).Interface()
	query := tx.Table(t.ModelTableName)
	if helper.InArray("deleted_at", t.ModelFields) {
		query = query.Where("deleted_at IS NULL")
	}
	err := query.Where(t.PkId+" = ?", id).First(result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		t.Result(errcode.NotExist, "数据不存在或已被删除")
		return
	}

	// 检查值是否发生变化
	modelValueField := reflect.ValueOf(result).Elem().FieldByName(field)
	if modelValueField.IsValid() && modelValueField.Interface() == value {
		tx.Rollback()
		t.Result(errcode.Success, "值未发生改变，请确认修改内容")
		return
	}

	// 更新数据
	updateData := map[string]interface{}{field: value}
	if err = tx.Table(t.ModelTableName).Where(t.PkId+" = ?", id).Updates(updateData).Error; err != nil {
		tx.Rollback()
		t.Result(errcode.DatabaseError, err.Error())
		return
	}

	// 调用自定义的SetValueAfter方法进行后置处理
	callResults = t.callCustomMethod("SetValueAfter", tx, id, field, value)
	if callResults[0] != nil {
		err, ok := callResults[0].(error)
		if ok && err != nil {
			tx.Rollback()
			t.Result(errcode.Unknown, err.Error())
			return
		}
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		t.Result(errcode.DatabaseTransactionCommitError, "事务提交失败，请重试")
		return
	}

	// 返回结果
	t.callCustomMethod("SetValueReturn", value, field, id)
}

func (t *Trait) SetValueFormData() (modelValue interface{}, mapData map[string]any, err error) {
	return t.SaveFormData()
}

// SetValueCheckField 验证传入的字段名是否有效
func (t *Trait) SetValueCheckField(field string) error {
	if !helper.InArray(field, t.ModelFields) {
		return errors.New("参数错误，请传入有效的字段名")
	}
	return nil
}

func (t *Trait) SetValueBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
	return modelValue, mapData, nil
}

func (t *Trait) SetValueAfter(tx *gorm.DB, id uint, field string, value any) error {
	return nil
}

func (t *Trait) SetValueReturn(value interface{}, field string, id uint) bool {
	t.Result(errcode.Success, "设置成功", gin.H{
		"id":    id,
		"field": field,
		"value": value,
	})
	return true
}
