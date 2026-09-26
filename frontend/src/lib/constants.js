// 中文状态 / 检测类型文案契约 —— 与 Go 侧 boce_status.go 的常量块逐字对应。
// 任何一侧改动都必须同步另一侧，否则 boce_status_test.go 的 TestJSStatusConstants 会变红。
export const STATUS_POLLUTION = '污染'
export const STATUS_INTERCEPTED = '拦截'
export const STATUS_UNFILED = '未备案'
export const STATUS_BLACKLISTED = '黑名单'
export const STATUS_WALLED = '被墙'
export const STATUS_SUSPECT_WALLED = '疑似被墙'
export const STATUS_BAD_DOMAIN = '域名格式错误'
export const STATUS_NORMAL = '正常'
export const STATUS_FILED = '已备案'
export const STATUS_UNREGISTERED = '未注册'
export const STATUS_FAILED = '失败'
export const STATUS_UNKNOWN = '未知'

export const TYPE_POLLUTION = '污染'
export const TYPE_QQ = 'QQ拦截'
export const TYPE_WECHAT = '微信拦截'
export const TYPE_ICP = '备案查询'
export const TYPE_BLACKLIST = '备案黑名单'
export const TYPE_WALL = '被墙检测'

// 筛选下拉的排序权重，顺序即优先级。
export const STATUS_LIST = [
  STATUS_POLLUTION,
  STATUS_INTERCEPTED,
  STATUS_UNFILED,
  STATUS_BLACKLISTED,
  STATUS_WALLED,
  STATUS_SUSPECT_WALLED,
  STATUS_BAD_DOMAIN,
  STATUS_NORMAL,
  STATUS_FILED,
  STATUS_UNREGISTERED,
  STATUS_FAILED,
  STATUS_UNKNOWN,
]

// 异常判定集合。
export const ABNORMAL_STATUSES = [
  STATUS_POLLUTION,
  STATUS_INTERCEPTED,
  STATUS_UNFILED,
  STATUS_BLACKLISTED,
  STATUS_WALLED,
  STATUS_SUSPECT_WALLED,
  STATUS_BAD_DOMAIN,
  STATUS_FAILED,
  STATUS_UNKNOWN,
]
