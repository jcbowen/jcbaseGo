package crud

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
)

// ActionDelete 删除数据的主要处理方法
// 参数说明：
//   - c *gin.Context: Gin框架的上下文对象，包含请求和响应信息
func (t *Trait) ActionDelete(c *gin.Context) {
	t.InitCrud(c, "delete")

	// 格式转换
	mapData := t.GetSafeMapGPC("all")

	// 获取ids参数
	idsInterface, exists := mapData[t.PkId+"s"]
	if !exists {
		t.Result(errcode.ParamError, "参数缺失，请重试")
		return
	}
	ids, ok := idsInterface.([]interface{})
	if !ok {
		t.Result(errcode.ParamError, "参数类型错误，请重试")
		return
	}

	// 过滤空id并去重
	idSet := make(map[interface{}]bool)
	var validIds []interface{}
	for _, id := range ids {
		if !helper.IsEmptyValue(id) && !idSet[id] {
			idSet[id] = true
			validIds = append(validIds, id)
		}
	}

	if len(validIds) == 0 {
		t.Result(errcode.ParamError, "参数缺失，请重试")
		return
	}

	// 设置删除数据的select字段
	fields := t.callCustomMethod("DeleteFields")[0].([]string)
	t.checkFieldExistInSelect(fields, t.PkId, "设置删除数据的")

	// 查询要删除的数据
	var delArr []map[string]interface{}
	deleteQuery := t.DBI.GetDb().Table(t.ModelTableName).
		Select(fields)
	// 应用软删除条件
	if t.SoftDeleteField != "" && helper.InArray(t.SoftDeleteField, t.ModelFields) {
		deleteQuery = deleteQuery.Where(t.SoftDeleteField + " " + t.SoftDeleteCondition)
	}
	// 获取删除条件
	deleteQuery = t.callCustomMethod("GetDeleteWhere", deleteQuery, validIds)[0].(*gorm.DB)
	err := deleteQuery.Find(&delArr).Error

	if err != nil {
		t.Result(errcode.DatabaseError, err.Error())
		return
	}

	if len(delArr) == 0 {
		t.Result(errcode.NotExist, "当前操作的数据不存在或已被删除")
		return
	}

	// 提取要删除的数据的主键ID
	var delIds []interface{}
	for _, item := range delArr {
		delIds = append(delIds, item[t.PkId])
	}

	// 删除前处理
	callResults := t.callCustomMethod("DeleteBefore", delArr, delIds)
	if callResults[1] != nil {
		err, ok = callResults[1].(error)
		if ok && err != nil {
			t.Result(errcode.ParamError, err.Error())
			return
		}
	}
	delIds = callResults[0].([]interface{})

	// 使用GORM的事务方法，自动处理提交和回滚
	err = t.DBI.GetDb().Transaction(func(tx *gorm.DB) error {
		// 执行删除
		if t.SoftDeleteField != "" && helper.InArray(t.SoftDeleteField, t.ModelFields) {
			// 软删除（更新软删除字段）
			condition := t.callCustomMethod("DeleteCondition", delArr)[0].(map[string]interface{})
			deleteQuery := tx.Model(t.Model)
			deleteQuery = t.callCustomMethod("GetDeleteWhere", deleteQuery, validIds)[0].(*gorm.DB)
			err = deleteQuery.Updates(condition).Error
		} else {
			// 真实删除
			deleteQuery := tx.Model(t.Model)
			deleteQuery = t.callCustomMethod("GetDeleteWhere", deleteQuery, validIds)[0].(*gorm.DB)
			err = deleteQuery.Delete(t.Model).Error
		}

		// 检查删除结果
		if err != nil {
			return err
		}

		// 删除后处理
		callErr := t.callCustomMethod("DeleteAfter", delIds, delArr)[0]
		if callErr != nil {
			err, ok = callErr.(error)
			if ok && err != nil {
				return err
			}
		}

		return nil
	})

	// 处理事务结果
	if err != nil {
		t.Result(errcode.DatabaseError, "删除失败："+err.Error())
		return
	}

	// 清理缓存
	// t.Model.ClearCache()

	// 返回结果
	t.callCustomMethod("DeleteReturn", delIds, delArr)
}

// DeleteFields 获取删除操作时需要查询的字段列表
// 返回值：
//   - []string: 字段名称列表，默认只包含主键字段
func (t *Trait) DeleteFields() []string {
	return []string{t.PkId}
}

// GetDeleteWhere 构建删除操作的WHERE条件
// 参数说明：
//   - deleteQuery *gorm.DB: 数据库查询对象
//   - ids []interface{}: 要删除的ID列表
//
// 返回值：
//   - *gorm.DB: 添加了WHERE条件的查询对象
func (t *Trait) GetDeleteWhere(deleteQuery *gorm.DB, ids []interface{}) *gorm.DB {
	return deleteQuery.Where(t.PkId+" IN ?", ids)
}

// DeleteBefore 删除前的钩子方法，用于数据预处理和验证
// 参数说明：
//   - delArr []map[string]interface{}: 要删除的数据记录列表
//   - delIds []interface{}: 要删除的ID列表
//
// 返回值：
//   - []interface{}: 处理后的ID列表
//   - error: 处理过程中的错误信息
func (t *Trait) DeleteBefore(delArr []map[string]interface{}, delIds []interface{}) ([]interface{}, error) {
	// 可以在此处添加一些前置处理逻辑
	return delIds, nil
}

// DeleteCondition 获取软删除的条件数据
// 参数说明：
//   - delArr []map[string]interface{}: 要删除的数据记录列表
//
// 返回值：
//   - map[string]interface{}: 软删除时要更新的字段和值
func (t *Trait) DeleteCondition(delArr []map[string]interface{}) map[string]interface{} {
	// 使用配置的软删除字段名
	fieldName := t.SoftDeleteField
	if fieldName == "" {
		// 理论上不应该到这里，因为调用此方法前已经检查了软删除字段
		// 但为了代码健壮性，如果确实没有配置，则使用 deleted_at 作为后备
		fieldName = "deleted_at"
	}

	// 获取当前时间
	now := time.Now()

	// 检查模型字段类型，决定返回字符串还是时间类型
	fieldType := t.getFieldType(fieldName)
	if fieldType == "time.Time" {
		// 如果字段是时间类型，返回 time.Time 对象
		return map[string]interface{}{
			fieldName: now,
		}
	} else {
		// 如果字段是字符串类型，返回格式化的时间字符串
		return map[string]interface{}{
			fieldName: now.Format("2006-01-02 15:04:05"),
		}
	}
}

// DeleteAfter 删除后的钩子方法，用于后续处理（在事务内执行）
// 参数说明：
//   - delIds []interface{}: 已删除的ID列表
//   - delArr []map[string]interface{}: 已删除的数据记录列表
//
// 返回值：
//   - error: 处理过程中的错误信息，如果返回错误则会回滚事务
func (t *Trait) DeleteAfter(delIds []interface{}, delArr []map[string]interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return nil
}

// DeleteReturn 删除成功后的返回处理方法
// 参数说明：
//   - delIds []interface{}: 已删除的ID列表
//   - delArr []map[string]interface{}: 已删除的数据记录列表
func (t *Trait) DeleteReturn(delIds []interface{}, delArr []map[string]interface{}) {
	t.Result(errcode.Success, "删除成功", gin.H{
		"delIds": delIds,
	})
}

// checkFieldExistInSelect 检查字段是否存在于选择字段列表中的辅助方法
// 参数说明：
//   - fields []string: 字段列表
//   - field string: 要检查的字段名
//   - context string: 上下文描述，用于错误信息
func (t *Trait) checkFieldExistInSelect(fields []string, field, context string) {
	if !helper.InArray(field, fields) {
		log.Panic(context + "字段未包含主键")
	}
}
