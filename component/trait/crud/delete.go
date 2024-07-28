package crud

import (
	"github.com/gin-gonic/gin"
	"github.com/jcbowen/jcbaseGo/component/helper"
	"github.com/jcbowen/jcbaseGo/component/security"
	"github.com/jcbowen/jcbaseGo/errcode"
	"gorm.io/gorm"
	"log"
	"time"
)

func (t *Trait) ActionDelete(c *gin.Context) {
	t.checkInit(c)

	// 获取GPC参数
	gpcInterface, GPCExists := t.GinContext.Get("GPC")
	if !GPCExists {
		t.Result(errcode.ParamError, "参数缺失，请重试")
		return
	}
	formDataMap := gpcInterface.(map[string]map[string]any)["all"]

	// 安全过滤
	sanitizedMapData := security.Input{Value: formDataMap}.Sanitize().(map[interface{}]interface{})
	// 格式转换
	mapData := make(map[string]interface{})
	for key, value := range sanitizedMapData {
		strKey := key.(string)
		mapData[strKey] = value
	}

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
	fields := t.callCustomMethod("GetDeleteFields")[0].([]string)
	t.checkFieldExistInSelect(fields, t.PkId, "设置删除数据的")

	// 查询要删除的数据
	var delArr []map[string]interface{}
	deleteQuery := t.MysqlMain.GetDb().Table(t.ModelTableName).
		Select(fields)
	if helper.InArray("deleted_at", t.ModelFields) {
		deleteQuery = deleteQuery.Where("deleted_at IS NULL")
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

	// 开启事务
	tx := t.MysqlMain.GetDb().Begin()

	// 获取删除操作更新的属性值
	condition := t.callCustomMethod("GetDeleteCondition", delArr)[0].(map[string]interface{})

	deleteQuery = tx.Model(t.Model)
	deleteQuery = t.callCustomMethod("GetDeleteWhere", deleteQuery, validIds)[0].(*gorm.DB)
	err = deleteQuery.Updates(condition).Error

	// 执行删除
	if err != nil {
		tx.Rollback()
		t.Result(errcode.StorageError, "删除失败，未知错误")
		return
	}

	// 删除后处理
	callResults = t.callCustomMethod("DeleteAfter", delIds, delArr)
	if callResults[0] != nil {
		err, ok = callResults[0].(error)
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

	// 清理缓存
	// t.Model.ClearCache()

	// 返回结果
	t.callCustomMethod("DeleteReturn", delIds, delArr)
}

func (t *Trait) GetDeleteFields() []string {
	return []string{t.PkId}
}

func (t *Trait) GetDeleteWhere(deleteQuery *gorm.DB, ids []interface{}) *gorm.DB {
	return deleteQuery.Where(t.PkId+" IN ?", ids)
}

func (t *Trait) DeleteBefore(delArr []map[string]interface{}, delIds []interface{}) ([]interface{}, error) {
	// 可以在此处添加一些前置处理逻辑
	return delIds, nil
}

func (t *Trait) GetDeleteCondition(delArr []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"deleted_at": time.Now().Format("2006-01-02 15:04:05"),
	}
}

func (t *Trait) DeleteAfter(delIds []interface{}, delArr []map[string]interface{}) error {
	// 可以在此处添加一些后置处理逻辑
	return nil
}

func (t *Trait) DeleteReturn(delIds []interface{}, delArr []map[string]interface{}) {
	t.Result(errcode.Success, "删除成功", gin.H{
		"delIds": delIds,
	})
}

// checkFieldExistInSelect is a helper function to check if a field exists in the select fields
func (t *Trait) checkFieldExistInSelect(fields []string, field, context string) {
	if !helper.InArray(field, fields) {
		log.Panic(context + "字段未包含主键")
	}
}
