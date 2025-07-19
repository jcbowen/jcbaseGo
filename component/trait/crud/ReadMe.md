# CRUD系统使用指南

本文档提供了CRUD系统的完整使用指南，包括快速入门、API参考和钩子方法详解。

## 目录

1. [快速入门](#快速入门)
2. [使用示例](#使用示例)
3. [钩子方法详解](#钩子方法详解)
4. [API参考](#api参考)
5. [最佳实践](#最佳实践)

## 快速入门

CRUD系统提供了完整的增删改查功能，支持自定义钩子方法进行业务逻辑扩展。

### 主要特性

- 🚀 **开箱即用**: 提供完整的CRUD操作
- 🔧 **高度可扩展**: 丰富的钩子方法支持自定义逻辑
- 🛡️ **事务安全**: 自动事务管理，支持回滚
- 📊 **分页支持**: 内置分页功能
- 🗑️ **软删除**: 支持软删除和硬删除，可自定义软删除字段名
- 🔍 **灵活查询**: 支持复杂查询条件

## 软删除配置

CRUD系统支持灵活的软删除配置，您可以自定义软删除字段名和删除条件。

### 配置方式

通过在模型字段的 `gorm` 标签中添加 `soft_delete` 标签来配置软删除：

```golang
import "github.com/jcbowen/jcbaseGo/component/orm/base"

type User struct {
    base.MysqlBaseModel  // 或者 base.SqliteBaseModel
    ID        uint   `gorm:"column:id;primaryKey" json:"id"`
    Name      string `gorm:"column:name;size:100" json:"name"`

    // 方式1: 使用系统默认的 deleted_at 字段
    DeletedAt string `gorm:"column:deleted_at;type:DATETIME;default:NULL" json:"deleted_at"`

    // 方式2: 自定义软删除字段，条件为 IS NULL
    // IsDeleted string `gorm:"column:is_deleted;type:DATETIME;soft_delete:IS NULL" json:"is_deleted"`

    // 方式3: 使用状态字段作为软删除，条件为 = 1（表示正常状态）
    // Status int `gorm:"column:status;type:INT;soft_delete:= 1" json:"status"`

    // 方式4: 使用特殊默认值的 deleted_at 字段
    // DeletedAt string `gorm:"column:deleted_at;type:DATETIME;default:0000-00-00 00:00:00" json:"deleted_at"`
}
```

### 配置说明

1. **默认配置**：如果模型中存在 `deleted_at` 字段且没有 `soft_delete` 标签，系统会自动将其作为软删除字段，条件为 `IS NULL`

2. **自定义字段名**：通过 `soft_delete` 标签可以指定任意字段作为软删除字段

3. **自定义条件**：`soft_delete` 标签的值就是软删除的判断条件，如：
   - `IS NULL`：字段为空表示未删除
   - `= 1`：字段值为1表示未删除
   - `= 'active'`：字段值为'active'表示未删除

4. **特殊默认值**：如果 `deleted_at` 字段的默认值为 `0000-00-00 00:00:00`，系统会自动使用 `= '0000-00-00 00:00:00'` 作为软删除条件

### 使用效果

- **查询时**：系统会自动添加软删除条件，只查询未删除的数据
- **删除时**：执行软删除操作，更新软删除字段的值
- **显示已删除数据**：通过 `show_deleted=1` 参数可以查看已删除的数据

## 使用示例

### 控制器
```golang
package user

import (
	"github.com/jcbowen/jcbaseGo/component/trait/crud"
	"github.com/jcbowen/jcbaseGo/component/orm/mysql"  // MySQL
	"github.com/jcbowen/jcbaseGo/component/orm/sqlite" // SQLite
	"officeAutomation/controllers/base"
	"officeAutomation/library"
	userModel "officeAutomation/model/common/user"
	"github.com/jcbowen/jcbaseGo/component/orm"
)

type Index struct {
	base.Controller
	*crud.Trait
}

// New 初始化并传递数据模型、数据库连接、当前控制器给crud
func New() *Index {
	// MySQL 示例
	index := &Index{
		Trait: &crud.Trait{
			Model: &userModel.Account{},
			DBI:    library.Mysql,
		},
	}
	
	// SQLite 示例
	/*
	sqliteDb, err := sqlite.New(jcbaseGo.SqlLiteStruct{
		DbFile: "path/to/your/database.db",
	})
	if err != nil {
		panic(err)
	}
	index := &Index{
		Trait: &crud.Trait{
			Model: &userModel.Account{},
			DBI:    sqliteDb,
		},
	}
	*/
	
	index.Trait.Controller = index
	return index
}

// ListEach 自定义一个ListEach方法替换crud中的ListEach
func (i Index) ListEach(item interface{}) interface{} {
	log.Println(item, "666", item.(*userModel.Account).Id)
	return item
}
```

### 路由配置
```golang
systemGroup := r.Group("/system")
systemGroup.Use(middleware.LoginRequired())
{
    systemUserGroup := systemGroup.Group("/user")
    {
		// 直接调用crud中的方法即可
        systemUser := systemUserController.New()
        systemUserGroup.GET("/list", systemUser.ActionList)      // 获取列表
        systemUserGroup.GET("/detail", systemUser.ActionDetail) // 获取详情
        systemUserGroup.POST("/create", systemUser.ActionCreate) // 创建
        systemUserGroup.POST("/update", systemUser.ActionUpdate) // 更新
        systemUserGroup.POST("/save", systemUser.ActionSave)     // 保存（自动判断创建/更新）
        systemUserGroup.POST("/delete", systemUser.ActionDelete) // 删除
        systemUserGroup.GET("/all", systemUser.ActionAll)        // 获取所有数据
        systemUserGroup.POST("/set-value", systemUser.ActionSetValue) // 设置字段值
    }
}
```

### 自定义钩子方法示例

```golang
// 创建前的数据验证
func (i *Index) CreateBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
    user := modelValue.(*userModel.Account)

    // 验证用户名是否已存在
    var count int64
    i.DBI.GetDb().Model(&userModel.Account{}).Where("username = ?", user.Username).Count(&count)
    if count > 0 {
        return nil, nil, errors.New("用户名已存在")
    }

    // 密码加密
    user.Password = helper.Md5(user.Password)

    return user, mapData, nil
}

// 更新前的数据验证（可以访问原始数据）
func (i *Index) UpdateBefore(modelValue interface{}, mapData map[string]any, originalData interface{}) (interface{}, map[string]any, error) {
    user := modelValue.(*userModel.Account)
    original := originalData.(*userModel.Account)

    // 如果密码有变化，进行加密
    if user.Password != "" && user.Password != original.Password {
        user.Password = helper.Md5(user.Password)
    }

    return user, mapData, nil
}

// 列表数据处理
func (i *Index) ListEach(item interface{}) interface{} {
    user := item.(*userModel.Account)
    // 隐藏敏感信息
    user.Password = ""
    return user
}

// 自定义查询条件
func (i *Index) ListQuery(query *gorm.DB) (*gorm.DB, error) {
    // 只显示启用的用户
    return query.Where("status = ?", 1), nil
}
```

## 钩子方法详解

CRUD系统提供了丰富的钩子方法，允许在不同阶段插入自定义逻辑。

### 执行时机图

```
创建流程:
FormData → CreateBefore → [事务开始] → 插入数据 → CreateAfter → [事务提交] → CreateReturn

更新流程:
FormData → 查询原始数据 → UpdateBefore → [事务开始] → 更新数据 → UpdateAfter → [事务提交] → UpdateReturn

删除流程:
获取参数 → 查询要删除的数据 → DeleteBefore → [事务开始] → 删除数据 → DeleteAfter → [事务提交] → DeleteReturn
```

### 重要改进说明

#### UpdateBefore方法优化 ⭐
- **调整时机**: 从"获取表单数据后"移动到"查询原始数据后、开始事务前"
- **新增参数**: 增加了`originalData`参数，可以访问数据库中的原始数据
- **优势**:
  - 可以基于原始数据进行更精确的验证和处理
  - 与DeleteBefore的设计模式保持一致
  - 避免无效处理（只有在数据确实存在时才会执行前置处理）

#### SaveBefore方法兼容性优化 ⭐
- **参数调整**: 使用可变参数`originalData ...interface{}`
- **兼容性**: 同时支持创建场景（不传originalData）和更新场景（传入originalData）
- **使用方式**: 通过`len(originalData) > 0 && originalData[0] != nil`判断是否为更新操作

```golang
func (i *Index) SaveBefore(modelValue interface{}, mapData map[string]any, originalData ...interface{}) (interface{}, map[string]any, error) {
    if len(originalData) > 0 && originalData[0] != nil {
        // 更新操作 - 可以访问原始数据
        original := originalData[0].(*userModel.Account)
        // 基于原始数据进行处理...
    } else {
        // 创建操作 - 没有原始数据
        // 进行创建前的处理...
    }

    return modelValue, mapData, nil
}
```

## API参考

### 主要操作方法

#### ActionCreate
- **功能**: 创建数据的主要处理方法
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: POST
- **请求参数**: 模型对应的字段数据

#### ActionUpdate
- **功能**: 更新数据的主要处理方法
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: POST
- **请求参数**: 包含主键ID和要更新的字段数据

#### ActionDelete
- **功能**: 删除数据的主要处理方法
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: POST
- **请求参数**: `{主键名}s` - ID数组，如 `ids: [1,2,3]`

#### ActionList
- **功能**: 获取数据列表的主要处理方法（分页）
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: GET
- **查询参数**:
  - `page`: 页码（默认1）
  - `page_size`: 每页数量（默认10）

#### ActionDetail
- **功能**: 获取数据详情的主要处理方法
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: GET
- **查询参数**: 主键ID

#### ActionAll
- **功能**: 获取所有数据的主要处理方法（不分页）
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: GET
- **查询参数**:
  - `show_deleted`: 是否显示已删除数据（0/1，默认0）

#### ActionSave
- **功能**: 保存数据的主要处理方法（自动判断创建或更新）
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: POST
- **请求参数**: 模型对应的字段数据，如果包含主键ID则为更新，否则为创建

#### ActionSetValue
- **功能**: 设置单个字段值的主要处理方法
- **参数**: `c *gin.Context` - Gin框架的上下文对象
- **HTTP方法**: POST
- **请求参数**:
  - 主键ID
  - `field`: 字段名
  - `value`: 字段值

### 钩子方法详细说明

#### 创建操作钩子

##### CreateFormData
- **功能**: 获取创建操作的表单数据
- **返回值**:
  - `modelValue interface{}` - 绑定后的模型实例
  - `mapData map[string]any` - 原始表单数据映射
  - `err error` - 处理过程中的错误信息

##### CreateBefore (钩子方法)
- **功能**: 创建前的钩子方法，用于数据预处理和验证
- **参数**:
  - `modelValue interface{}` - 要创建的模型实例
  - `mapData map[string]any` - 表单数据映射
- **返回值**:
  - `interface{}` - 处理后的模型实例
  - `map[string]any` - 处理后的表单数据映射
  - `error` - 处理过程中的错误信息

##### CreateAfter (钩子方法)
- **功能**: 创建后的钩子方法，用于后续处理（在事务内执行）
- **参数**:
  - `tx *gorm.DB` - 数据库事务对象
  - `modelValue interface{}` - 已创建的模型实例
- **返回值**: `error` - 处理过程中的错误信息，如果返回错误则会回滚事务

##### CreateReturn
- **功能**: 创建成功后的返回处理方法
- **参数**: `item any` - 创建成功的数据项
- **返回值**: `bool` - 处理结果，通常返回true表示成功

#### 更新操作钩子

##### UpdateFormData
- **功能**: 获取更新操作的表单数据
- **返回值**:
  - `modelValue interface{}` - 绑定后的模型实例
  - `mapData map[string]any` - 原始表单数据映射
  - `err error` - 处理过程中的错误信息

##### UpdateBefore (钩子方法) ⭐ 已优化
- **功能**: 更新前的钩子方法，用于数据预处理和验证
- **参数**:
  - `modelValue interface{}` - 要更新的模型实例（包含新数据）
  - `mapData map[string]any` - 表单数据映射
  - `originalData interface{}` - 数据库中的原始数据
- **返回值**:
  - `interface{}` - 处理后的模型实例
  - `map[string]any` - 处理后的表单数据映射
  - `error` - 处理过程中的错误信息

##### UpdateAfter (钩子方法)
- **功能**: 更新后的钩子方法，用于后续处理（在事务内执行）
- **参数**:
  - `tx *gorm.DB` - 数据库事务对象
  - `modelValue interface{}` - 已更新的模型实例
- **返回值**: `error` - 处理过程中的错误信息，如果返回错误则会回滚事务

##### UpdateReturn
- **功能**: 更新成功后的返回处理方法
- **参数**: `item interface{}` - 更新成功的数据项
- **返回值**: `bool` - 处理结果，通常返回true表示成功

#### 删除操作钩子

##### DeleteFields
- **功能**: 获取删除操作时需要查询的字段列表
- **返回值**: `[]string` - 字段名称列表，默认只包含主键字段

##### GetDeleteWhere
- **功能**: 构建删除操作的WHERE条件
- **参数**:
  - `deleteQuery *gorm.DB` - 数据库查询对象
  - `ids []interface{}` - 要删除的ID列表
- **返回值**: `*gorm.DB` - 添加了WHERE条件的查询对象

##### DeleteBefore (钩子方法)
- **功能**: 删除前的钩子方法，用于数据预处理和验证
- **参数**:
  - `delArr []map[string]interface{}` - 要删除的数据记录列表
  - `delIds []interface{}` - 要删除的ID列表
- **返回值**:
  - `[]interface{}` - 处理后的ID列表
  - `error` - 处理过程中的错误信息

##### DeleteCondition
- **功能**: 获取软删除的条件数据
- **参数**: `delArr []map[string]interface{}` - 要删除的数据记录列表
- **返回值**: `map[string]interface{}` - 软删除时要更新的字段和值

##### DeleteAfter (钩子方法)
- **功能**: 删除后的钩子方法，用于后续处理（在事务内执行）
- **参数**:
  - `delIds []interface{}` - 已删除的ID列表
  - `delArr []map[string]interface{}` - 已删除的数据记录列表
- **返回值**: `error` - 处理过程中的错误信息，如果返回错误则会回滚事务

##### DeleteReturn
- **功能**: 删除成功后的返回处理方法
- **参数**:
  - `delIds []interface{}` - 已删除的ID列表
  - `delArr []map[string]interface{}` - 已删除的数据记录列表

#### 查询操作钩子

##### ListSelect
- **功能**: 设置列表查询的SELECT字段
- **参数**: `query *gorm.DB` - 数据库查询对象
- **返回值**: `*gorm.DB` - 设置了SELECT字段的查询对象

##### ListQuery
- **功能**: 设置列表查询的WHERE条件和其他查询参数
- **参数**: `query *gorm.DB` - 数据库查询对象
- **返回值**:
  - `*gorm.DB` - 设置了查询条件的查询对象
  - `error` - 处理过程中的错误信息

##### ListOrder
- **功能**: 设置列表查询的排序规则
- **返回值**: `interface{}` - 排序规则，可以是字符串或其他GORM支持的排序格式

##### ListEach
- **功能**: 对列表中的每个数据项进行处理
- **参数**: `item interface{}` - 列表中的单个数据项
- **返回值**: `interface{}` - 处理后的数据项

##### ListReturn
- **功能**: 列表查询成功后的返回处理方法
- **参数**: `listData jcbaseGo.ListData` - 包含列表数据和分页信息的结构体
- **返回值**: `bool` - 处理结果，通常返回true表示成功

##### DetailSelect
- **功能**: 设置详情查询的SELECT字段
- **参数**: `query *gorm.DB` - 数据库查询对象
- **返回值**: `*gorm.DB` - 设置了SELECT字段的查询对象

##### DetailQuery
- **功能**: 设置详情查询的WHERE条件和其他查询参数
- **参数**:
  - `query *gorm.DB` - 数据库查询对象
  - `mapData map[string]any` - 请求参数映射
- **返回值**: `*gorm.DB` - 设置了查询条件的查询对象

##### Detail
- **功能**: 对详情数据进行处理
- **参数**: `item interface{}` - 查询到的详情数据
- **返回值**: `interface{}` - 处理后的详情数据

##### DetailReturn
- **功能**: 详情查询成功后的返回处理方法
- **参数**: `detail interface{}` - 详情数据
- **返回值**: `bool` - 处理结果，通常返回true表示成功

#### 保存操作钩子

##### SaveBefore (钩子方法) ⭐ 已优化
- **功能**: 保存前的钩子方法，用于数据预处理和验证
- **参数**:
  - `modelValue interface{}` - 要保存的模型数据
  - `mapData map[string]any` - 表单数据映射
  - `originalData ...interface{}` - 原始数据（仅在更新操作时提供，创建操作时为nil）
- **返回值**:
  - `interface{}` - 处理后的模型实例
  - `map[string]any` - 处理后的表单数据映射
  - `error` - 处理过程中的错误信息

##### SaveAfter (钩子方法)
- **功能**: 保存后的钩子方法，用于后续处理（在事务内执行）
- **参数**:
  - `tx *gorm.DB` - 数据库事务对象
  - `modelValue interface{}` - 已保存的模型实例
- **返回值**: `error` - 处理过程中的错误信息，如果返回错误则会回滚事务

## 最佳实践

### 1. 钩子方法使用建议

- **Before方法**: 用于数据验证、权限检查、数据预处理等
- **After方法**: 用于日志记录、缓存更新、消息推送等后续处理
- **Return方法**: 用于自定义返回格式和内容
- **错误处理**: Before和After方法返回错误时会中断操作并回滚事务（如果在事务内）

### 2. 性能优化建议

- 在Before方法中进行轻量级验证，避免复杂的数据库查询
- 使用After方法进行异步处理，如发送邮件、更新缓存等
- 合理使用Select方法，只查询需要的字段
- 在Query方法中添加适当的索引条件

### 3. 安全建议

- 在Before方法中进行权限验证
- 对敏感字段进行加密处理
- 在Each方法中过滤敏感信息
- 使用参数绑定防止SQL注入

### 4. 事务管理和锁定优化 ⭐

#### 事务最佳实践
- **使用GORM的Transaction方法**: 自动处理提交和回滚，避免手动管理事务
- **避免长时间事务**: 在After方法中避免执行耗时的外部API调用或文件处理
- **事务超时控制**: 设置合理的事务超时时间，防止长时间锁定

```golang
// 推荐的事务处理方式
err := db.Transaction(func(tx *gorm.DB) error {
    // 数据库操作
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    
    // 轻量级的后续处理
    if err := tx.Create(&profile).Error; err != nil {
        return err
    }
    
    return nil
})
```

#### 避免锁定的建议
- **避免嵌套事务**: 不要在钩子方法中开启新的事务
- **及时释放锁**: 确保事务尽快提交或回滚
- **使用乐观锁**: 对于并发更新场景，考虑使用版本号或时间戳
- **连接池配置**: 合理配置连接池参数，防止连接泄漏

```golang
// 乐观锁示例
type User struct {
    ID      uint   `gorm:"primaryKey"`
    Name    string
    Version int    `gorm:"default:1"` // 版本号字段
}

// 更新时检查版本号
func (u *User) UpdateWithOptimisticLock(tx *gorm.DB, newName string) error {
    result := tx.Model(u).
        Where("id = ? AND version = ?", u.ID, u.Version).
        Updates(map[string]interface{}{
            "name":    newName,
            "version": u.Version + 1,
        })
    
    if result.RowsAffected == 0 {
        return errors.New("数据已被其他用户修改，请重试")
    }
    
    return nil
}
```

#### 连接池配置建议
```golang
// MySQL连接池配置
sqlDB, err := db.DB()
if err == nil {
    sqlDB.SetMaxOpenConns(100)        // 最大连接数
    sqlDB.SetMaxIdleConns(10)         // 最大空闲连接数
    sqlDB.SetConnMaxLifetime(5 * time.Minute)  // 连接最大生命周期
    sqlDB.SetConnMaxIdleTime(3 * time.Minute)  // 空闲连接超时时间
}
```

### 5. 错误处理

```golang
func (i *Index) CreateBefore(modelValue interface{}, mapData map[string]any) (interface{}, map[string]any, error) {
    // 数据验证
    if err := i.validateData(modelValue); err != nil {
        return nil, nil, err // 返回错误会中断操作
    }

    return modelValue, mapData, nil
}
```

### 6. 日志记录

```golang
func (i *Index) CreateAfter(tx *gorm.DB, modelValue interface{}) error {
    user := modelValue.(*userModel.Account)

    // 记录操作日志
    log.Printf("用户创建成功: ID=%d, Username=%s", user.Id, user.Username)

    return nil
}
```

所有钩子方法都支持在继承的控制器中重写，以实现自定义的业务逻辑。