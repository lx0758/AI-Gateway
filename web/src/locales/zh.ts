export default {
  app: {
    title: 'AI 代理',
    shortTitle: 'AI'
  },
  common: {
    save: '保存',
    cancel: '取消',
    delete: '删除',
    edit: '编辑',
    detail: '详情',
    create: '创建',
    search: '搜索',
    status: '状态',
    action: '操作',
    enabled: '已启用',
    disabled: '已禁用',
    confirm: '确认',
    success: '成功',
    error: '错误',
    loading: '加载中...',
    noData: '暂无数据',
    total: '总计',
    name: '名称',
    type: '类型',
    description: '描述',
    yes: '是',
    no: '否',
    to: '至'
  },
  login: {
    title: 'AI 代理',
    username: '用户名',
    password: '密码',
    submit: '登录',
    logout: '退出登录',
    invalidCredentials: '用户名或密码错误'
  },
  menu: {
    dashboard: '仪表盘',
    providers: '厂商管理',
    models: '模型映射',
    apiKeys: 'API 密钥',
    usage: '用量统计',
    settings: '系统设置'
  },
  dashboard: {
    totalRequests: '总请求数',
    todayRequests: '今日请求',
    activeProviders: '活跃厂商',
    activeKeys: '活跃密钥',
    requestTrend: '请求趋势（近 7 天）',
    providerDistribution: '厂商分布',
    modelRanking: '模型使用排名'
  },
  provider: {
    name: '厂商名称',
    apiType: 'API 类型',
    baseUrl: '接口地址',
    apiKey: 'API 密钥',
    apiKeyPlaceholder: '留空则保持当前密钥不变',
    models: '模型',
    testConnection: '测试连接',
    syncModels: '同步模型',
    addProvider: '添加厂商',
    editProvider: '编辑厂商',
    lastSync: '最后同步',
    modelId: '模型 ID',
    contextWindow: '上下文窗口'
  },
  apiKey: {
    name: '密钥名称',
    key: 'API 密钥',
    allowedModels: '允许模型',
    quota: '配额',
    usedQuota: '已用',
    rateLimit: '速率限制',
    expiresAt: '过期时间',
    createKey: '创建 API 密钥',
    allModels: '所有模型'
  },
  usage: {
    stats: '统计',
    logs: '日志',
    totalTokens: '总 Token',
    promptTokens: '提示 Token',
    completionTokens: '补全 Token',
    successRate: '成功率',
    avgLatency: '平均延迟',
    totalRequests: '总请求数',
    time: '时间',
    model: '模型',
    startTime: '开始时间',
    endTime: '结束时间'
  },
  settings: {
    changePassword: '修改密码',
    oldPassword: '旧密码',
    newPassword: '新密码',
    confirmPassword: '确认密码',
    passwordChanged: '密码修改成功'
  },
  modelMapping: {
    alias: '模型别名',
    actualModel: '实际模型',
    weight: '权重',
    model: '模型',
    required: '此项为必填'
  }
}
