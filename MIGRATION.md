# Migration Guide: v1.x → v2.0

## Breaking Changes

### Provider Configuration

**v1.x (Old)**
```json
{
  "name": "OpenAI",
  "type": "openai",
  "base_url": "https://api.openai.com/v1",
  "api_key": "sk-xxx"
}
```

**v2.0 (New)**
```json
{
  "name": "OpenAI",
  "openai_base_url": "https://api.openai.com/v1",
  "anthropic_base_url": "",
  "api_key": "sk-xxx"
}
```

### Key Changes

1. **Removed**: `type` field
2. **Removed**: `base_url` field
3. **Added**: `openai_base_url` field
4. **Added**: `anthropic_base_url` field
5. **Validation**: At least one base URL must be provided

## Automatic Migration

The system automatically migrates existing providers on startup:

- `type="openai"` + `base_url="xxx"` → `openai_base_url="xxx"`, `anthropic_base_url=""`
- `type="anthropic"` + `base_url="xxx"` → `openai_base_url=""`, `anthropic_base_url="xxx"`

## Manual Steps

1. **Update API Clients**: Change `type` and `base_url` to `openai_base_url` and/or `anthropic_base_url`
2. **Update Frontend**: Clear browser cache and reload the management interface
3. **Verify Providers**: Check that all providers show correct API styles in the UI

## Rollback

If you need to rollback:
1. Stop the new version
2. Restore the database from backup
3. Deploy the old version

## Benefits

- **Unified Provider**: One provider can support both OpenAI and Anthropic formats
- **No Duplicate Models**: Models are stored once per provider
- **Smart Routing**: Automatically prefers direct calls over format conversion
- **Future Ready**: Provider list enables load balancing and failover