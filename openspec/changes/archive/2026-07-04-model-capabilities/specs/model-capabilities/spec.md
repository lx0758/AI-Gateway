## ADDED Requirements

### Requirement: 模型能力多选管理

系统 SHALL 允许管理员为每个 ProviderModel 选择多项能力类型，包括 Tools(工具)、Stream(流式)、Photo(照片)、Image(图片)、Video(视频)。

#### Scenario: 创建模型时默认选中 Tools 和 Stream
- **WHEN** 管理员打开添加模型表单
- **THEN** Tools(工具) 和 Stream(流式) 默认勾选
- **AND** Photo(照片)、Image(图片)、Video(视频) 默认未勾选

#### Scenario: 选择多种能力类型
- **WHEN** 管理员勾选 Tools、Image、Video
- **THEN** 系统保存 capabilities 为 `"tools,image,video"`
- **AND** 后端 ProviderModel 记录的 capabilities 字段存储该值

#### Scenario: 全部不选时显示 None
- **WHEN** 管理员取消所有勾选
- **THEN** 系统保存 capabilities 为空字符串
- **AND** 展示时显示 "None" 标签

#### Scenario: 编辑时回显已选能力
- **WHEN** 管理员打开编辑表单
- **AND** 已有记录的 capabilities = `"stream,vision"`
- **THEN** Stream(流式) 和 Image(图片) 复选框勾选
- **AND** 其他复选框未勾选

#### Scenario: 列表展示能力标签
- **WHEN** 管理员查看模型列表
- **THEN** 能力列显示多个彩色标签（Tools/Stream/Photo/Image/Video）
- **AND** 全部未选时显示 "None" 标签
