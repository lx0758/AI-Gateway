## RENAMED Requirements

FROM: `AliasMapping` → TO: `ModelMapping`
FROM: `alias_mappings` → TO: `model_mappings`

---

## MODIFIED Requirements

### Requirement: ModelMapping references Provider and ProviderModel

The system SHALL associate ModelMapping with Provider (by provider_id) and reference ProviderModel by model_id string, including model capability and token information.

#### Scenario: Mapping with valid provider
- **WHEN** admin creates ModelMapping with provider_id=5, provider_model_name="gpt-4-turbo"
- **THEN** system verifies Provider.id=5 exists
- **AND** system verifies ProviderModel(provider_id=5, model_id="gpt-4-turbo") exists
- **AND** creation succeeds

#### Scenario: Mapping with invalid provider
- **WHEN** admin creates ModelMapping with provider_id=999
- **THEN** creation fails with "provider not found" error

#### Scenario: Mapping with invalid model
- **WHEN** admin creates ModelMapping with provider_id=5, provider_model_name="nonexistent"
- **THEN** creation fails with "provider model not found" error

#### Scenario: Mapping includes model info in API response
- **WHEN** API returns ModelMapping in detail page
- **THEN** each mapping includes model_info object
- **AND** model_info contains {context_window, max_output, supports_vision, supports_tools, supports_stream}
- **AND** values are retrieved from ProviderModel table in real-time

### Requirement: ModelMapping belongs to Model

The system SHALL enforce foreign key relationship between ModelMapping.model_id and Model.id.

#### Scenario: Create mapping requires existing model
- **WHEN** admin creates ModelMapping with model_id=1
- **THEN** system verifies Model.id=1 exists
- **AND** creation succeeds if model exists
- **AND** creation fails with "model not found" error if model does not exist

#### Scenario: Mapping retrieved with model
- **WHEN** router queries ModelMapping
- **THEN** system can Preload Model relationship
- **AND** mapping.model.name field is accessible

### Requirement: ModelMapping has weight for load balancing

The system SHALL use ModelMapping.weight field for routing priority.

#### Scenario: Higher weight routes first
- **WHEN** Model "gpt-4" has ModelMappings with weights [10, 50, 30]
- **THEN** router returns providers sorted by weight DESC: [50, 30, 10]

#### Scenario: Default weight is 1
- **WHEN** admin creates ModelMapping without specifying weight
- **THEN** system sets weight=1 by default

#### Scenario: Weight updated by drag-drop sorting
- **WHEN** admin drags mapping to position 1 in detail page
- **THEN** system sets weight = total_mappings - 1
- **WHEN** admin drags mapping to last position
- **THEN** system sets weight = 0

### Requirement: ModelMapping can be enabled or disabled

The system SHALL allow enabling/disabling individual ModelMapping without affecting sibling mappings.

#### Scenario: Disabled mapping excluded from routing
- **WHEN** ModelMapping.enabled=false
- **THEN** router excludes this mapping from provider list

#### Scenario: Disabled mapping still visible in UI
- **WHEN** ModelMapping.enabled=false
- **THEN** admin can see mapping in detail page
- **AND** admin can re-enable mapping via status switch

#### Scenario: UI displays model info for disabled mappings
- **WHEN** ModelMapping.enabled=false
- **THEN** admin can see model token and capability information
- **AND** information helps admin decide to re-enable or delete

### Requirement: ModelMapping supports Provider association

The system SHALL Preload Provider relationship for ModelMapping queries.

#### Scenario: Mapping includes Provider info
- **WHEN** API returns ModelMapping
- **THEN** mapping includes provider object with {id, name, openai_base_url, anthropic_base_url}
- **AND** UI can display provider name and type tags

#### Scenario: UI displays model capabilities in detail page
- **WHEN** admin views model detail page mapping table
- **THEN** system displays Capabilities column showing Vision, Tools, Stream tags
- **AND** tags display with colors: Vision(green), Tools(orange), Stream(blue)
- **AND** tags are based on model_info values

### Requirement: ModelMapping supports drag-drop reordering

The system SHALL allow reordering ModelMappings by drag-drop in detail page, automatically updating weights.

#### Scenario: Drag-drop updates weights linearly
- **WHEN** admin drags and drops mappings to reorder
- **THEN** system calculates new weights: position 1 = total - 1, position 2 = total - 2, ..., last = 0
- **AND** system calls PUT `/models/:id/mappings/order` API
- **AND** API updates all mappings weights in database

#### Scenario: Drag-drop API receives order array
- **WHEN** frontend calls PUT `/models/:id/mappings/order`
- **THEN** request body contains `{ "order": [mapping_id_1, mapping_id_2, ...] }`
- **AND** system updates weights based on array position index

#### Scenario: Drag-drop preserves other attributes
- **WHEN** system updates weights via drag-drop
- **THEN** model_id, provider_id, provider_model_name, enabled remain unchanged
- **AND** only weight values are modified
