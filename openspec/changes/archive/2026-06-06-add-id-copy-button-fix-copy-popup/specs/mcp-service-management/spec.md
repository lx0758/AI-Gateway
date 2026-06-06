## MODIFIED Requirements

### Requirement: 列出 MCP 服务

系统 SHALL 提供端点列出所有配置的 MCP 服务及其元数据。

#### Scenario: 列出所有服务
- **WHEN** 管理员请求服务列表
- **THEN** 系统返回所有服务，包含名称、标识符、类型、启用状态和最后同步时间

### Requirement: MCP 服务列表的 name 列支持复制

系统 SHALL 在 MCP 服务列表页面的 name 列提供复制按钮。

#### Scenario: name 列显示复制按钮
- **WHEN** 管理员查看 MCP 服务列表
- **THEN** name 列在每个服务名称旁显示一个复制图标按钮

#### Scenario: 点击复制按钮复制服务名称
- **WHEN** 管理员点击 name 旁边的复制按钮
- **THEN** 系统将服务名称复制到剪贴板
- **AND** 显示成功提示

### Requirement: MCP 详情页面的子项名称列支持复制

系统 SHALL 在 MCP 详情页面的工具、资源、提示词名称列提供复制按钮。

#### Scenario: 工具名称列显示复制按钮
- **WHEN** 管理员查看 MCP 详情页面的工具列表
- **THEN** 每个工具名称旁显示一个复制图标按钮

#### Scenario: 资源名称列显示复制按钮
- **WHEN** 管理员查看 MCP 详情页面的资源列表
- **THEN** 每个资源名称旁显示一个复制图标按钮

#### Scenario: 提示词名称列显示复制按钮
- **WHEN** 管理员查看 MCP 详情页面的提示词列表
- **THEN** 每个提示词名称旁显示一个复制图标按钮
