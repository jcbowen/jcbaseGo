### 使用示例

#### 控制器
```golang
package user

import (
	"github.com/jcbowen/jcbaseGo/component/trait/crud"
	"log"
	"officeAutomation/controllers/base"
	"officeAutomation/library"
	userModel "officeAutomation/model/common/user"
)

type Index struct {
	base.Controller
	*crud.Trait
}

// New 初始化并传递数据模型、数据库连接、当前控制器给crud
func New() *Index {
	index := &Index{
		Trait: &crud.Trait{
			Model:     &userModel.Account{},
			MysqlMain: library.Mysql,
		},
	}
	index.Trait.Controller = index
	return index
}

// ListEach 自定义一个ListEach方法替换crud中的ListEach
func (i Index) ListEach(item interface{}) interface{} {
	log.Println(item, "666", item.(*userModel.Account).Id)
	return item
}
```

#### router
```golang
systemGroup := r.Group("/system")
systemGroup.Use(middleware.LoginRequired())
{
    systemUserGroup := systemGroup.Group("/user")
    {
		// 直接调用crud中的方法即可
        systemUser := systemUserController.New()
        systemUserGroup.GET("/list", systemUser.ActionList)
    }
    
}
```