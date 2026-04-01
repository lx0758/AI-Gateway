## Context

当前系统采用"Provider.Type 决定 Provider 实现类"的设计：
- Provider 有 Type 字段（"openai" 或 "anthropic"）
- Provider 有单一 BaseURL 字段
- Factory.Create(provider) 根据 Type 返回 OpenAIProvider 或 AnthropicProvider
- OpenAIProvider 和 AnthropicProvider 各自实现 4 条通路（透传 + 转换）

问题场景：一个后端同时支持 OpenAI 和 Anthropic 格式时，需要创建两个 Provider，导致：
- ProviderModel 数据重复（同样的模型在两个 Provider 下各存储一份）
- ModelMapping 配置复杂（需要为每个 Provider 配置映射）
- 管理界面冗余（同一个后端在列表中出现两次）

## Goals / Non-Goals

**Goals:**
- Provider 数据模型支持配置多个 BaseURL（OpenAI 和 Anthropic 可同时配置）
- 消除同一后端的 Provider 和 ProviderModel 数据重复
- 路由逻辑增加格式匹配优先级，减少不必要的转换开销
- 保持向后兼容（支持平滑迁移）

**Non-Goals:**
- 不合并 OpenAIProvider 和 AnthropicProvider 实现类（保持 4 条通路的清晰性）
- 不改变 ProviderModel 和 ModelMapping 的数据结构（它们已经是"一个模型属于一个 Provider"）
- 不改变 Provider 的认证方式（APIKey 仍然共用，OpenAI 和 Anthropic 使用相同认证）
- 不改变两个 Provider 实现类的内部逻辑（转换代码不变）

## Decisions

### 1. Provider 数据模型设计

**决策：** Provider 结构改为 `OpenAIBaseURL + AnthropicBaseURL`，删除 Type 字段

**理由：**
- 通过 BaseURL 是否为空判断支持的能力，比 Type 字段更直观
- 一个 Provider 对应一个厂商，概念统一
- 避免同一厂商创建多个 Provider 实例

**替代方案：**
- 方案 A：保留 Type，改为数组 `Types: ["openai", "anthropic"]`
  - 问题：BaseURL 如何处理？OpenAI 和 Anthropic 的 endpoint 路径不同（/chat/completions vs /messages）
  - 需要额外配置或路径转换规则，复杂度增加
  
- 方案 B：嵌套配置 `OpenAI: {BaseURL, APIKey}, Anthropic: {BaseURL, APIKey}`
  - 问题：APIKey 相同时造成配置冗余
  - 当前场景中 APIKey 是共用的，不需要分开

**选择方案：** OpenAIBaseURL + AnthropicBaseURL（简洁，满足需求）

### 2. Router 和 Factory 合并设计

**决策：** 删除 Factory 层，Router.Route() 直接返回 Provider 实例列表

**理由：**
- Router 已经掌握所有信息（Provider 数据 + requestStyle）
- 无需通过 Factory 再转换一层
- 返回实例列表支持未来负载均衡和故障转移
- 简化调用链：Router → Handler（减少一层）

**Router 内部逻辑：**
```
Router.Route(alias, requestStyle):
  1. 查询 ModelMapping(alias) → Provider数据列表（按weight排序）
  
  2. 分组：
     GroupA: Provider.OpenAIBaseURL != "" (支持OpenAI)
     GroupB: Provider.AnthropicBaseURL != "" (支持Anthropic)
  
  3. 根据请求格式重组候选顺序：
     if requestStyle == "openai":
       candidates = GroupA + GroupB  // 优先OpenAI
     else:
       candidates = GroupB + GroupA  // 优先Anthropic
  
  4. 遍历 Provider数据，创建实例：
     for each provider in candidates:
       // 确定调用格式
       if provider在优先组:
         style = requestStyle
       else:
         style = opposite(requestStyle)
       
       // 根据style和BaseURL创建实例
       if style == "openai":
         if provider.OpenAIBaseURL != "":
           instance = OpenAIProvider(BaseURL=OpenAIBaseURL)
         else:
           instance = AnthropicProvider(BaseURL=AnthropicBaseURL)
       else:
         if provider.AnthropicBaseURL != "":
           instance = AnthropicProvider(BaseURL=AnthropicBaseURL)
         else:
           instance = OpenAIProvider(BaseURL=OpenAIBaseURL)
       
       instances.append(instance)
  
  5. 返回实例列表：
     return []Provider{instances}
```

**Handler 调用：**
```
candidates := router.Route(model, "openai")
if len(candidates) == 0 {
  return 404
}

// 当前：取第一个
provider := candidates[0]
provider.ExecuteOpenAIRequest(ctx, pm)

// 未来负载均衡：
// provider := loadBalancer.Select(candidates)

// 未来故障转移：
// for provider := range candidates {
//   if provider.Execute() == nil {
//     break
//   }
// }
```

**替代方案：**
- 方案 A：保留 Factory，Router 返回 RouteResult{Provider数据, CallStyle}
  - 问题：多一层抽象，Handler 需要调用 Factory
  - RouteResult.DirectCall 和 CallStyle 冗余（实例内部已包含这些信息）
  
- 方案 B：Router 返回单个 Provider 实例
  - 问题：无法支持负载均衡和故障转移
  - 丢失候选列表信息

**选择方案：** Router 直接返回 []Provider 实例列表（简洁，支持未来扩展）

**循环依赖处理：**
```
当前依赖：router → model, provider → model
新增依赖：router → provider

检查循环：router → provider → model → router ✗

解决方案：Provider 不依赖 model.Provider，只依赖 Config
  - provider.Config 包含 BaseURL、APIKey
  - Router 提取 model.Provider 数据，创建 Config，传给 Provider 构造函数
  - Provider 实例不持有 model.Provider 引用
```

### 3. Router 路由算法设计

**决策：** Router.Route(alias, requestStyle) 返回按格式优先级排序的 Provider 实例列表

**理由：**
- Weight 是第一优先级（用户配置的优先级）
- 格式匹配是第二优先级（减少转换开销）
- 返回完整列表支持负载均衡和故障转移

**算法：**
```
1. 查询 ModelMapping(alias) → Provider数据列表
2. 按 weight DESC 排序
3. 分组：
   - GroupA: 支持 requestStyle 的 Provider
   - GroupB: 不支持 requestStyle 的 Provider
4. 合并：GroupA + GroupB（GroupA 优先，内部按 weight 排序）
5. 遍历创建实例（见决策2）
6. 返回 []Provider
```

**示例：**
```
请求: OpenAI 格式，alias="gpt-4"
ModelMapping:
  ProviderA (weight=100, OpenAIBaseURL="x", AnthropicBaseURL="")  # 支持OpenAI
  ProviderB (weight=80,  OpenAIBaseURL="", AnthropicBaseURL="y")  # 只支持Anthropic
  ProviderC (weight=50,  OpenAIBaseURL="z", AnthropicBaseURL="")  # 支持OpenAI

分组：
  GroupA (支持OpenAI): [ProviderA, ProviderC]
  GroupB (不支持): [ProviderB]

排序后：
  GroupA: [ProviderA(100), ProviderC(50)]
  GroupB: [ProviderB(80)]

最终候选顺序：
  [ProviderA, ProviderC, ProviderB]

创建实例：
  [
    OpenAIProvider(BaseURL="x"),   // 透传
    OpenAIProvider(BaseURL="z"),   // 透传
    AnthropicProvider(BaseURL="y"),          // 降级，转换O→A
  ]

Handler:
  candidates[0].ExecuteOpenAIRequest()  // 使用ProviderA，透传
```

**替代方案：**
- 方案 A：格式匹配优先，weight 作为第二优先级
  - 问题：违背用户配置意图（weight 是用户定义的优先级）
  
- 方案 B：只路由到支持同格式的 Provider，否则报错
  - 问题：失去转换能力，降级体验差
  
- 方案 C：Router 只返回单个 Provider 实例
  - 问题：无法支持负载均衡和故障转移

**选择方案：** Weight 优先 + 格式匹配作为第二优先级 + 返回完整列表（平衡用户意图、性能和扩展性）

### 4. 数据迁移策略

**决策：** 提供迁移脚本，将现有 Provider.Type 和 BaseURL 映射到新的字段

**迁移规则：**
```
Type="openai"    → OpenAIBaseURL=原BaseURL, AnthropicBaseURL=""
Type="anthropic" → OpenAIBaseURL="", AnthropicBaseURL=原BaseURL
```

**实施方式：**
- GORM AutoMigrate 自动创建新字段
- 手动迁移脚本填充数据（type → baseurl）
- 迁移后删除 type 列

**替代方案：**
- 方案 A：保留 type 字段作为"兼容字段"，deprecated
  - 问题：数据模型混乱，长期维护负担
  
- 方案 B：不迁移，要求用户手动更新配置
  - 问题：用户体验差，可能导致服务中断

**选择方案：** 自动迁移脚本（平滑升级）

## Risks / Trade-offs

**[Risk] 数据迁移失败导致服务不可用**
- Mitigation：迁移脚本在启动时自动执行，失败时回滚；提供手动迁移命令

**[Risk] 前端向后兼容性（旧版本前端可能继续发送 type 字段）**
- Mitigation：Handler 兼容处理，如果收到 type 字段，忽略并验证 BaseURL；文档明确标记 breaking change

**[Risk] 用户不理解新的配置方式（两个 BaseURL 输入框）**
- Mitigation：前端增加提示"至少填写一个"；文档更新配置示例；提供迁移后的配置预览

**[Risk] Router 对 provider 包的依赖引入循环依赖**
- Mitigation：Provider 只依赖 Config 结构体，不依赖 model.Provider；Router 提取数据创建 Config

**[Trade-off] Router 职责增加（路由 + Provider 创建）**
- 优势：简化调用链，删除 Factory 层；返回实例列表支持未来扩展
- 劣势：Router 代码复杂度增加
- 结论：合理的职责合并，Provider 创建逻辑本身就是路由的一部分

**[Trade-off] Router 增加 requestStyle 参数**
- 优势：路由逻辑清晰，支持格式优先级
- 劣势：Route 方法签名变更，调用方需要传递参数
- 结论：必要的变更，参数语义明确

## Migration Plan

### Phase 1: 数据模型迁移（启动时）
1. GORM AutoMigrate 创建新列：`openai_base_url`, `anthropic_base_url`
2. 迁移脚本执行：
   ```sql
   UPDATE providers SET openai_base_url = base_url WHERE type = 'openai';
   UPDATE providers SET anthropic_base_url = base_url WHERE type = 'anthropic';
   ```
3. 验证迁移结果（所有 Provider 至少一个 BaseURL 不为空）

### Phase 2: 代码部署
1. Provider 实现类调整：构造函数接受 Config，不依赖 model.Provider
2. Router.Route 改造：增加 requestStyle 参数，返回 []Provider，内部创建实例
3. 删除 Factory 文件或保留作为内部辅助
4. Handler 调用逻辑调整：直接使用 Router 返回的实例列表
5. 前端表单更新

### Phase 3: 清理
1. 删除 Provider.Type 字段（手动执行）
2. 删除 Factory 相关代码（如果保留）
3. 删除旧的 Handler 兼容逻辑

### Rollback Strategy
- 如果迁移失败：回滚数据库（删除新列），部署旧版本代码
- 如果代码问题：快速回滚到上一版本（数据模型兼容，type 字段仍可使用）

## Open Questions

- 是否需要在迁移后保留 type 字段一段时间（标记 deprecated），以便渐进式清理？
- 前端表单是否需要"一键配置"功能（填写一个 BaseURL 后，自动生成另一个的路径）？
- Router 返回的 Provider 实例列表是否需要携带 weight 信息（用于负载均衡权重分配）？
- 是否需要在 Provider 接口增加方法判断是否需要转换（用于日志和监控）？