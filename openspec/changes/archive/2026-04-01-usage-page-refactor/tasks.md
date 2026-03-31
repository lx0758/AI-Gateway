## 1. Backend Changes

- [x] 1.1 Modify `/usage/logs` handler - remove `LIMIT 1000` constraint
- [x] 1.2 Remove `/usage/stats` handler method (Stats func)
- [x] 1.3 Remove `/usage/key-stats` handler method (KeyStats func)
- [x] 1.4 Remove routes for deprecated endpoints if单独注册

## 2. Frontend - Core Refactor

- [x] 2.1 Refactor Usage/index.vue - merge fetchStats/fetchKeyStats into single fetchLogs
- [x] 2.2 Implement frontend aggregation function for calculating stats from raw logs
- [x] 2.3 Remove calls to deprecated APIs (/usage/stats, /usage/key-stats)

## 3. Frontend - New Statistics

- [x] 3.1 Add source statistics card (aggregate by source field)
- [x] 3.2 Add provider statistics card (aggregate by provider_name field)
- [x] 3.3 Add provider-model statistics card (aggregate by provider_name + model)
- [x] 3.4 Simplify model statistics - remove actual_model column, aggregate by model only

## 4. Frontend - UI Reorder

- [x] 4.1 Reorder cards: source stats → key stats → model stats → provider stats → provider-model stats → logs
