## Context

当前管理后台使用 Element Plus 构建，涉及四个核心列表页面。模型映射表 `ModelMapping` 为单表设计，同一 alias 在多条记录中重复，代理逻辑通过 `router.Route(name)` 按 alias 查询映射并按权重排序选择厂商。

本次重构将模型映射拆分为一对多结构，涉及：
- 数据模型新增 `Alias` 和 `AliasMapping` 表
- API 新增嵌套资源 `/aliases/:id/mappings`
- 代理路由逻辑更新
- 前端别名管理页面重构

## Goals / Non-Goals

**Goals:**
- 数据模型重构为一对多，alias 唯一，AliasMapping 属于 Alias
- RESTful 嵌套 API 设计
- 前端别名管理页面支持折叠展开、厂商类型列显示
- 厂商和模型列表按名称排序
- 用量日志错误信息支持复制

**Non-Goals:**
- 不迁移旧数据（原 ModelMapping 表删除）
- 不删除现有数据库
- 不实现复杂的多列排序或高级筛选

## Decisions

### 1. 数据模型设计

**决策**: `Alias` 主表 + `AliasMapping` 子表

```
Alias (主表)
├── ID uint (primaryKey)
├── Name string (uniqueIndex) ← 用户调用的模型名
├── Enabled bool (default:true)
├── CreatedAt time.Time
├── UpdatedAt time.Time
├── DeletedAt gorm.DeletedAt
├── Mappings []AliasMapping (hasMany)

AliasMapping (子表)
├── ID uint (primaryKey)
├── AliasID uint (index) ← 外键
├── ProviderID uint (index)
├── ProviderModelName string
├── Weight int (default:1)
├── Enabled bool (default:true)
├── CreatedAt time.Time
├── UpdatedAt time.Time
├── DeletedAt gorm.DeletedAt
├── Provider *Provider (belongsTo)
```

**理由**:
- alias 唯一，避免重复数据
- 映射归属清晰，支持别名级别的启用/禁用
- GORM 支持 hasMany/belongsTo 关联，查询方便

### 2. API 设计

**决策**: 嵌套资源 RESTful API

```
Alias 管理:
GET    /aliases                    → 别名列表（含 mapping_count）
POST   /aliases                    → 创建别名
GET    /aliases/:id                → 别名详情 + mappings
PUT    /aliases/:id                → 更别名名称/状态
DELETE /aliases/:id                → 删除别名（级联 mappings）

Mapping 管理:
GET    /aliases/:id/mappings       → 该别名的映射列表
POST   /aliases/:id/mappings       → 添加映射
PUT    /aliases/:id/mappings/:mid  → 更映射
DELETE /aliases/:id/mappings/:mid  → 删映射
```

### 3. 代理路由逻辑更新

**决策**: 更新 `router.Route(name)` 使用新表结构

**新逻辑**:
```go
var alias model.Alias
if err := model.DB.Where("name = ? AND enabled = ?", name, true).First(&alias).Error; err != nil {
    return nil, nil
}

var mappings []model.AliasMapping
model.DB.Preload("Provider").
    Where("alias_id = ? AND enabled = ?", alias.ID, true).
    Order("weight DESC").
    Find(&mappings)
```

### 4. 前端页面设计

**决策**: 折叠卡片 + 内嵌表格

```vue
<el-collapse>
  <el-collapse-item v-for="alias in aliases" :key="alias.id">
    <template #title>
      Alias: {{ alias.name }} ({{ alias.mapping_count }} 映射)
    </template>
    <el-table :data="alias.mappings">
      <!-- columns -->
    </el-table>
  </el-collapse-item>
</el-collapse>
```

### 5. 厂商类型判断

**决策**: 根据 `provider.openai_base_url` 和 `provider.anthropic_base_url` 显示 Tag

### 6. 错误信息复制

**决策**: 复制按钮 + `navigator.clipboard.writeText()` + ElMessage 提示

## Risks / Trade-offs

- **API 变更影响** → 旧 `/model-mappings` 接口已删除
- **代理逻辑切换时机** → 新表可用后立即切换
- **前端交互** → 使用 Element Plus Collapse，避免手写折叠逻辑
- **级联删除** → GORM 约束级联删除，删除 alias 时自动删除 AliasMappings