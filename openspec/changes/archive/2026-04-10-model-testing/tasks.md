## 1. Backend - Handler

- [x] 1.1 Create `server/internal/handler/model_test.go`
- [x] 1.2 Implement `TestProviderModel` handler (POST /api/v1/providers/:id/models/:model_id/test)
- [x] 1.3 Implement `TestModel` handler (POST /api/v1/models/:id/test)
- [x] 1.4 Implement `executeTest` helper function with httptest context
- [x] 1.5 Implement `extractResponseContent` helper for OpenAI/Anthropic response parsing

## 2. Backend - Router

- [x] 2.1 Register test routes in `server/cmd/server/main.go`

## 3. Frontend - Provider Model Test

- [x] 3.1 Add test button to provider model table in `web/src/views/Providers/Detail.vue`
- [x] 3.2 Create test result dialog component
- [x] 3.3 Add i18n translations for test UI (zh/en)

## 4. Frontend - Virtual Model Test

- [x] 4.1 Add test button to Model Detail page in `web/src/views/Models/Detail.vue`
- [x] 4.2 Display test results for all mappings
- [x] 4.3 Add i18n translations for test UI (zh/en)

## 5. Testing

- [ ] 5.1 Manual test provider model testing API
- [ ] 5.2 Manual test virtual model testing API
- [ ] 5.3 Manual test frontend UI
