# Proposal

## Why

控制台两处信息展示不完整，管理员需要横向滚动或靠猜测才能读全内容：

- 厂商详情页模型列表的模型 ID 列固定 `width="200"`，长模型 ID 被截断且无任何查看完整值的途径
- 用量页日志表"厂商/模型"列固定 `width="150"`，`el-tag` 内容超长时直接被裁剪；"厂商模型统计"卡片的模型列同样无保护

## What Changes

- 厂商详情页模型 ID 列**固定加宽至 300px**（列后紧跟复制按钮，加宽后两者仍有足够空间）
- 用量页日志表"厂商/模型"列增加 **hover tooltip**，悬浮时显示完整的 `厂商/模型` 内容（沿用同页 `client_ips`、`error_msg` 已有的 tooltip 写法）
- 用量页"厂商模型统计"卡片的模型列一并增加 tooltip
- 长内容必须可完整查看，成为该两页的展示约定

**不做**：不改动其他页面的列宽与展示；不引入虚拟滚动等重型方案。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `manual-model-ui`: 厂商详情页模型列表中，模型 ID 必须完整可读（加宽列宽）。
- `usage-tracking`: 使用页的日志与统计表中，长内容（厂商/模型名）需可通过悬浮完整查看。

## Impact

- `web/src/views/Providers/Detail.vue`：模型 ID 列 `width` 调整
- `web/src/views/ModelUsage/index.vue`：日志表"厂商/模型"列与统计卡片模型列增加 `el-tooltip`
- 无后端改动、无 DB 变更
- 属像素级 UI 调整，spec 只约束"完整可读 / 可悬浮查看"的行为，具体宽度数值留在 tasks 作为实现细节
