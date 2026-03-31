## Why

当前 Handler 直接使用 `model` 结构体作为 API 响应结构，存在以下问题：

1. **耦合严重**：数据库模型和 API 契约绑定，DB 改了 API 就变
2. **暴露风险**：虽然用了 `json:"-"`，但同一结构体混用不安全
3. **无法精细控制**：DB 模型可能有 20 个字段，API 可能只需 5 个
4. **验证混杂**：DB 约束和 API 输入验证混在一起

## What Changes

在现有 Handler 文件中定义 DTO 结构，不再新建文件：

```
internal/handler/
├── key.go            # Handler + DTO 在一起
├── provider.go       # Handler + DTO 在一起
└── ...
```

## Capabilities

### New Capabilities

- `dto-separation`：将 API 响应 DTO 与数据库 model 分离

### Modified Capabilities

- 无（不影响现有功能，只是重构）

## Impact

**改动的文件：**

| 文件 | 改动内容 |
|------|----------|
| `key.go` | 新增 APIKey 相关 DTO |
| `provider.go` | 新增 Provider 相关 DTO |
| `model_mapping.go` | 新增 ModelMapping 相关 DTO |
| `provider_model.go` | 新增 ProviderModel 相关 DTO |

**原则：**
- Handler 只使用 DTO 与前端交互
- model 只在 handler 内部用于数据库操作
- DTO 命名：`CreateXXXRequest`, `UpdateXXXRequest`, `XXXResponse`
- 在现有 handler 文件末尾添加 DTO 定义
