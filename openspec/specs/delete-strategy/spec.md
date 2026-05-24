# 删除策略规格

## 概述

本规格定义了系统中业务实体的删除策略，包括软删除、硬删除和级联删除的实现方式。

---

## Requirements

### Requirement: 业务实体使用软删除

系统 SHALL 对以下业务实体使用软删除策略：User, Provider, Model, MCP, Key。

#### Scenario: 删除用户使用软删除
- **WHEN** 管理员删除 User
- **THEN** 系统设置 deleted_at 字段
- **AND** 记录仍存在于数据库中
- **AND** 普通查询不返回该记录

#### Scenario: 删除厂商使用软删除
- **WHEN** 管理员删除 Provider
- **THEN** 系统设置 deleted_at 字段
- **AND** 关联的 ProviderModel 和 ModelMapping 被硬删除

#### Scenario: 删除网关模型使用软删除
- **WHEN** 管理员删除 Model
- **THEN** 系统设置 deleted_at 字段
- **AND** 关联的 ModelMapping 被硬删除

#### Scenario: 删除 MCP 服务使用软删除
- **WHEN** 管理员删除 MCP
- **THEN** 系统设置 deleted_at 字段
- **AND** 关联的 MCPTool, MCPResource, MCPPrompt 被硬删除

#### Scenario: 删除 API Key 使用软删除
- **WHEN** 管理员删除 Key
- **THEN** 系统设置 deleted_at 字段
- **AND** 关联的 KeyModel, KeyMCPTool, KeyMCPResource, KeyMCPPrompt 被硬删除

---

### Requirement: 关联表和子资源使用硬删除

系统 SHALL 对以下关联表和子资源使用硬删除策略：ProviderModel, ModelMapping, MCPTool, MCPResource, MCPPrompt, KeyModel, KeyMCPTool, KeyMCPResource, KeyMCPPrompt。

#### Scenario: 删除 ProviderModel 使用硬删除
- **WHEN** 管理员删除 ProviderModel
- **THEN** 系统从数据库永久删除记录
- **AND** 关联的 ModelMapping 被级联硬删除

#### Scenario: 删除 ModelMapping 使用硬删除
- **WHEN** 删除 ModelMapping（通过级联删除）
- **THEN** 系统从数据库永久删除记录
- **AND** 记录无法恢复

#### Scenario: 删除 MCPTool 使用硬删除
- **WHEN** 删除 MCP 服务时级联删除 MCPTool
- **THEN** 系统从数据库永久删除所有关联的 MCPTool

---

### Requirement: 级联删除由代码实现

系统 SHALL 在 Handler 层实现级联删除逻辑，不依赖数据库 CASCADE 约束。

#### Scenario: 删除 Provider 时级联删除
- **WHEN** 管理员删除 Provider
- **THEN** 系统先硬删除所有 ProviderModel
- **AND** 然后硬删除所有关联的 ModelMapping
- **AND** 最后软删除 Provider

#### Scenario: 删除 ProviderModel 时级联删除
- **WHEN** 管理员删除 ProviderModel
- **THEN** 系统先硬删除所有关联的 ModelMapping
- **AND** 然后硬删除 ProviderModel

#### Scenario: 删除 Model 时级联删除
- **WHEN** 管理员删除 Model
- **THEN** 系统硬删除所有关联的 ModelMapping
- **AND** 软删除 Model

#### Scenario: 删除 MCP 时级联删除
- **WHEN** 管理员删除 MCP
- **THEN** 系统硬删除所有关联的 MCPTool, MCPResource, MCPPrompt
- **AND** 软删除 MCP

#### Scenario: 删除 Key 时级联删除
- **WHEN** 管理员删除 Key
- **THEN** 系统硬删除所有关联的 KeyModel, KeyMCPTool, KeyMCPResource, KeyMCPPrompt
- **AND** 软删除 Key
