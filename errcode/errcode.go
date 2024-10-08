package errcode

const (
	// 9001 为服务端逻辑运行相关错误
	// 9002 为客户端逻辑运行相关错误
	// 9003 三方服务的错误（如：MySQL、redis连接等）
	// 2 开头一般为不同的成功状态

	// Success 操作成功
	Success = 200
	// Unknown 未知错误
	Unknown = 1

	//----- 客户端错误 ----- /

	// BadRequest 请求无效或格式错误
	BadRequest = 400
	// Unauthorized 未授权
	Unauthorized = 401
	// Forbidden 服务器理解请求但拒绝执行
	Forbidden = 403
	// NotFound 请求的资源不存在
	NotFound = 404
	// MethodNotAllowed 请求方法不被允许
	MethodNotAllowed = 405
	// Conflict 请求与服务器的状态冲突
	Conflict = 409

	// SuccessResponse 响应成功
	SuccessResponse = Success
	// SuccessChange 新建或修改成功
	SuccessChange = 2001
	// SuccessQueue 成功进入后台队列
	SuccessQueue = 2002
	// SuccessDelete 删除成功
	SuccessDelete = 2004

	// SystemBusy 系统繁忙
	SystemBusy = 9001001
	// UnderMaintenance 维护中
	UnderMaintenance = 9001002
	// InvalidConfig 无效的配置
	InvalidConfig = 9001003
	// NoConfig 没有配置
	NoConfig = 9001004

	// LostConnection 失去连接
	LostConnection = 9002001
	// InvalidSign 无效的签名
	InvalidSign = 9002002
	// InvalidTimestamp 无效的时间戳
	InvalidTimestamp = 9002003
	// LoginInvalid 登录失效
	LoginInvalid = 9002004
	// InvalidToken 无效的token
	InvalidToken = 9002028
	// NoPermission 无权访问
	// @deprecated 请使用 NoPermissionEdit
	NoPermission = 9002011
	// NoPermissionVisit 无权访问
	NoPermissionVisit = 9002005
	// NoPermissionEdit 无权编辑
	NoPermissionEdit = 9002011
	// IncorrectUsernameOrPassword 账号或密码错误
	IncorrectUsernameOrPassword = 9002006
	// DISABLE 禁用
	DISABLE = 9002007
	// ParamError 参数错误
	ParamError = 9002008
	// ParamInvalid 无效参数
	ParamInvalid = 9002009
	// ParamMissing 参数缺失
	ParamMissing = 9002010
	// ParamEmpty 部分参数为空
	ParamEmpty = 9002008 // @deprecated 请使用 ParamError
	// EXISTED 已存在
	EXISTED = 9002011
	// NotExist 不存在或已被删除
	NotExist = 9002012
	// ConflictWithExisting 和已有数据冲突
	ConflictWithExisting = 9002013
	// IllegalAccess 不合法的访问
	IllegalAccess = 9002014
	// IllegalFormat 不合法的格式
	IllegalFormat = 9002015
	// IllegalType 不合法的类型
	IllegalType = 9002016
	// IllegalSize 不合法的大小
	IllegalSize = 9002017
	// IllegalCertificate 不合法的凭证
	IllegalCertificate = 9002018
	// UNAUTHORIZED 未授权
	// @deprecated 请使用 Unauthorized
	UNAUTHORIZED = Unauthorized
	// InvalidAuthorizationInformation 无效的授权信息
	InvalidAuthorizationInformation = 9002020
	// ExpirationService 服务到期
	ExpirationService = 9002021
	// ExpirationUpgrade 大更新到期
	ExpirationUpgrade = 9002022
	// ExpirationUpdate 小更新到期
	ExpirationUpdate = 9002023
	// ExpirationUse 使用权到期
	ExpirationUse = 9002024
	// ExpirationApi 使用权到期
	ExpirationApi = 9002025
	// InoperableState 不可操作的状态
	InoperableState = 9002026
	// NotSupported 暂不支持
	NotSupported = 9002027
	// NetworkConnectionError 网络连接-错误
	NetworkConnectionError = 9002028
	// NetworkConnectionTimeout 网络连接-超时
	NetworkConnectionTimeout = 9002029
	// NetworkConnectionInterrupt 网络连接-中断
	NetworkConnectionInterrupt = 9002030
	// NetworkConnectionRefused 网络连接-拒绝
	NetworkConnectionRefused = 9002031
	// NetworkConnectionReset 网络连接-重置
	NetworkConnectionReset = 9002032
	// NoBindPhone 没有绑定手机号
	NoBindPhone = 9002033
	// NoBindEmail 没有绑定邮箱
	NoBindEmail = 9002034
	// NoBindWechat 没有绑定微信
	NoBindWechat = 9002037
	// NoBindWechatMiniProgram 没有绑定微信小程序
	NoBindWechatMiniProgram = 9002038
	// NoBindAlipay 没有绑定支付宝
	NoBindAlipay = 9002039
	// NoSetPassword 没有设置密码
	NoSetPassword = 9002035
	// NoSetPayPassword 没有设置支付密码
	NoSetPayPassword = 9002036

	// StorageError 数据存储错误
	StorageError = 9003001
	// DatabaseError 数据库错误
	DatabaseError = 9003002
	// DatabaseConnectionError 数据库连接错误
	DatabaseConnectionError = 9003003
	// DatabaseQueryError 数据库查询错误
	DatabaseQueryError = 9003004
	// DatabaseWriteError 数据库写入错误
	DatabaseWriteError = 9003005
	// DatabaseUpdateError 数据库更新错误
	DatabaseUpdateError = 9003006
	// DatabaseDeleteError 数据库删除错误
	DatabaseDeleteError = 9003007
	// DatabaseTransactionError 数据库事务错误
	DatabaseTransactionError = 9003008
	// DatabaseStoredProcedureError 数据库存储过程错误
	DatabaseStoredProcedureError = 9003009
	// DatabaseTriggerError 数据库触发器错误
	DatabaseTriggerError = 9003010
	// DatabaseViewError 数据库视图错误
	DatabaseViewError = 9003011
	// DatabaseFunctionError 数据库函数错误
	DatabaseFunctionError = 9003012
	// DatabaseIndexError 数据库索引错误
	DatabaseIndexError = 9003013
	// DatabaseSequenceError 数据库序列错误
	DatabaseSequenceError = 9003014
	// DatabaseConstraintError 数据库约束错误
	DatabaseConstraintError = 9003015
	// DatabaseLockError 数据库锁错误
	DatabaseLockError = 9003016
	// DatabaseTransactionIsolationLevelError 数据库事务隔离级别错误
	DatabaseTransactionIsolationLevelError = 9003017
	// DatabaseTransactionLockError 数据库事务锁定错误
	DatabaseTransactionLockError = 9003018
	// DatabaseTransactionTimeoutError 数据库事务超时错误
	DatabaseTransactionTimeoutError = 9003019
	// DatabaseTransactionDeadlockError 数据库事务死锁错误
	DatabaseTransactionDeadlockError = 9003020
	// DatabaseTransactionRollbackError 数据库事务回滚错误
	DatabaseTransactionRollbackError = 9003021
	// DatabaseTransactionCommitError 数据库事务提交错误
	DatabaseTransactionCommitError = 9003022
	// DatabaseTransactionSavepointError 数据库事务保存点错误
	DatabaseTransactionSavepointError = 9003023
	// DatabaseTransactionRollbackToSavepointError 数据库事务回滚到保存点错误
	DatabaseTransactionRollbackToSavepointError = 9003024
	// DatabaseTransactionReleaseSavepointError 数据库事务释放保存点错误
	DatabaseTransactionReleaseSavepointError = 9003025
	// DatabaseTransactionRollbackToReleaseSavepointError 数据库事务回滚到释放保存点错误
	DatabaseTransactionRollbackToReleaseSavepointError = 9003026
	// NetworkError 网络错误(未指定服务端还是客户端)
	NetworkError = 9003502

	// NoPreAction 未进行前置活动 9005000
	NoPreAction = 9005000
	// NoNextAction 未进行后置活动 9006000
	NoNextAction = 9005001

	// Other 其他错误
	Other = 9999999
)
