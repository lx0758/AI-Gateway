# Spec: Project Naming Rename

本次变更仅涉及项目命名标识符的调整，无功能行为变更，因此无需创建或修改功能规格文件。

## Rationale

- 项目名称从 "AI代理/AI Proxy" 更改为 "AI网关/AI Gateway" 是定位调整，不改变现有功能行为
- Go module 名称、环境变量前缀、前端标题等均为标识符，不影响功能逻辑
- 所有现有功能的规格（如 alias-mapping、anthropic-api 等）保持不变

## Conclusion

无需新增或修改功能规格文件。本次变更的任务清单（tasks.md）将直接基于 design.md 和现有代码结构生成实施步骤。