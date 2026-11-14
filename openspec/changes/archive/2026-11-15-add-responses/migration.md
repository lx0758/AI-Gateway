# 上线迁移手册：Provider BaseURL 字段词表改名

> 本变更将两个 BaseURL 字段按外部产品名全链改名，并新增 Responses 字段。代码**不读旧列名、无兼容层**，因此数据库列重命名必须在部署新版本**之前**完成。

## 一、变更内容

| 旧列名 | 新列名 |
|---|---|
| `openai_base_url` | `openai_chat_completions_base_url` |
| `anthropic_base_url` | `anthropic_messages_base_url` |
| （无） | `openai_responses_base_url`（新增，默认空串） |

- 新列 `openai_responses_base_url` 由 GORM `AutoMigrate` 在**首次启动新版本**时自动创建（空串），无需手工加列。
- 旧两列必须**先**手工改名；`AutoMigrate` 只会新增列，不会重命名列。
- 顺序强制：**先执行 DDL 改列 → 再部署新版本**。顺序颠倒会导致新版本启动即读不到列而失败（快速失败，不会静默错数据）。

## 二、SQLite

```sql
ALTER TABLE providers RENAME COLUMN openai_base_url TO openai_chat_completions_base_url;
ALTER TABLE providers RENAME COLUMN anthropic_base_url TO anthropic_messages_base_url;
```

> SQLite 需 ≥ 3.25（`RENAME COLUMN` 支持）。执行前建议停机或确认无写入。

## 三、PostgreSQL

```sql
ALTER TABLE providers RENAME COLUMN openai_base_url TO openai_chat_completions_base_url;
ALTER TABLE providers RENAME COLUMN anthropic_base_url TO anthropic_messages_base_url;
```

## 四、执行步骤

1. **备份数据库**（SQLite 直接复制文件；PostgreSQL 使用 `pg_dump`）。
2. 停止旧版本服务。
3. 在数据库上执行上述两条 `RENAME COLUMN`。
4. 校验列名：

   ```sql
   -- SQLite
   PRAGMA table_info(providers);
   -- PostgreSQL
   SELECT column_name FROM information_schema.columns WHERE table_name = 'providers';
   ```

   应看到 `openai_chat_completions_base_url`、`anthropic_messages_base_url`，且不再有旧名。
5. 部署并启动新版本。`AutoMigrate` 自动新增 `openai_responses_base_url`（空串）。
6. 校验：管理后台 Provider 列表三列正常回显；存量 Provider 的 Responses 列为空（不具备原生 Responses 能力 → 自动走协议转换）。

## 五、回滚

回滚需把两列名改回旧名（与上线操作对称），再回退到旧版本：

```sql
ALTER TABLE providers RENAME COLUMN openai_chat_completions_base_url TO openai_base_url;
ALTER TABLE providers RENAME COLUMN anthropic_messages_base_url TO anthropic_base_url;
```

`openai_responses_base_url` 列留空对旧版本无副作用，无需删除。

## 六、上线后配置建议

- 上游已支持原生 Responses（如千问 `compatible-mode/v1`、智谱 `api/v1`、MiMo 等）：补填 `openai_responses_base_url`，Responses 请求走原生透传。
- 智谱类 chat 与 responses 前缀不一致的厂商必须分别填写（如 chat `https://open.bigmodel.cn/api/paas/v4`、responses `https://open.bigmodel.cn/api/v1`）。
- 不填 `openai_responses_base_url`：安全降级为跨协议转换，不影响既有流量。
