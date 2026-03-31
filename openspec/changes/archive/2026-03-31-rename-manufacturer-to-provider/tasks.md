## 1. 重命名目录和包

- [x] 1.1 将 `server/internal/manufacturer/` 目录重命名为 `server/internal/provider/`
- [x] 1.2 更新 `provider.go` 中的包声明（原 `manufacturer.go`）
- [x] 1.3 更新 `factory.go` 中的包声明

## 2. 重命名接口和类型

- [x] 2.1 将 `Manufacturer` 接口重命名为 `Provider`
- [x] 2.2 将 `NewOpenAICompatibleManufacturer` 函数重命名为 `NewOpenAICompatibleProvider`
- [x] 2.3 将 `NewAnthropicManufacturer` 函数重命名为 `NewAnthropicProvider`

## 3. 重命名实现文件

- [x] 3.1 将 `manufacturer_anthropic.go` 重命名为 `provider_anthropic.go`
- [x] 3.2 将 `manufacturer_anthropic_test.go` 重命名为 `provider_anthropic_test.go`
- [x] 3.3 将 `manufacturer_openai_compatible.go` 重命名为 `provider_openai_compatible.go`
- [x] 3.4 将 `manufacturer_openai_compatible_test.go` 重命名为 `provider_openai_compatible_test.go`

## 4. 更新所有引用

- [x] 4.1 更新 `handler/provider_model.go` 中的导入路径和类型引用
- [x] 4.2 更新 `handler/proxy_openai.go` 中的导入路径和类型引用
- [x] 4.3 更新 `handler/proxy_anthropic.go` 中的导入路径和类型引用

## 5. 验证

- [x] 5.1 运行 `go build ./...` 确保编译通过
- [x] 5.2 运行 `go test ./...` 确保所有测试通过
