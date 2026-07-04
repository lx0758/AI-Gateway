## 0. 已有数据迁移（手动，不写代码）

- [x] 0.1 在 design.md 中生成手动迁移 SQL 示例
- [ ] 0.2 部署前手动执行 SQL 迁移

## 1. 后端 - 移除旧 bool 字段 & 新增 capabilities

- [x] 1.1 在 db.go ProviderModel 实体中移除 SupportsVision / SupportsTools / SupportsStream 字段
- [x] 1.2 在 db.go ProviderModel 实体中新增 Capabilities string 字段
- [x] 1.3 在 provider_model.go DTO (createProviderModelRequest / providerModelResponse) 中移除三 bool 字段
- [x] 1.4 在 provider_model.go DTO 中新增 capabilities string 字段
- [x] 1.5 在 Create handler 中用 capabilities 替代旧 bool 字段
- [x] 1.6 在 Update handler 中用 capabilities 替代旧 bool 字段
- [x] 1.7 在 toProviderModelResponse 中传递 capabilities
- [x] 1.8 在 Sync handler 中同步传递 capabilities，移除旧 bool 字段写入

## 2. 后端 - 更新依赖旧 bool 字段的逻辑

- [x] 2.1 更新 handler/model.go 的 calculateCapabilitiesIntersection → calculateCapabilitiesUnion，从 capabilities 字符串推导
- [x] 2.2 更新 handler/model.go 的 toMappingResponse，使用 capabilities 替代旧 bool 字段
- [x] 2.3 更新 handler/key.go 的 response struct，移除三 bool 字段，新增 capabilities
- [x] 2.4 更新 internal/provider/provider_openai.go 的 ProviderModel struct，移除三 bool 字段
- [x] 2.5 更新 internal/provider/provider_anthropic.go 的 ProviderModel struct，移除三 bool 字段

## 3. 后端 - Context 解析工具

- [x] 3.1 新建 parsing.go，实现 parseContextString(str string) (int, error) 函数
- [x] 3.2 支持纯数字、K/k、M/m、B/b 后缀解析（1024-based）
- [x] 3.3 在 ProviderModelHandler 的 Create/Update 中调用 parseContextString 转换输入

## 4. 前端 - Format 工具函数

- [x] 4.1 修改 format.ts 中 formatToken 使用 1024-based 单位 (1K=1024, 1M=1048576)
- [x] 4.2 新增 parseContextString(str string) → number 函数
- [x] 4.3 新增 formatContextInput(num: number) → string 函数（显示用，如 131072 → "128K"）

## 5. 前端 - ProviderModel 编辑表单

- [x] 5.1 在 Detail.vue 中添加能力多选 checkbox 组（Tools/Stream/Photo/Image/Video）
- [x] 5.2 表单初始化默认勾选 Tools 和 Stream
- [x] 5.3 将 context_window 输入改为文本框，使用 parseContextString 解析
- [x] 5.4 编辑回显时将 context_window 数字格式化为 "128K" 显示
- [x] 5.5 将能力展示改为标签组形式（替代原有三个独立 tag）
- [x] 5.6 更新 API 请求体，移除旧 bool 字段，新增 capabilities

## 6. 前端 - i18n 翻译

- [x] 6.1 在 zh.ts 的 provider 命名空间新增翻译键（capabilities, tools, stream, photo, image, video, contextInputPlaceholder）
- [x] 6.2 在 en.ts 的 provider 命名空间新增翻译键

## 7. 验证

- [x] 7.1 运行 `go build ./...` 确保后端编译通过
- [x] 7.2 运行 `vue-tsc --noEmit` 确保前端类型检查通过
- [ ] 7.3 验证已有数据迁移后 capabilities 值正确
- [ ] 7.4 手动测试添加/编辑模型的表单交互
