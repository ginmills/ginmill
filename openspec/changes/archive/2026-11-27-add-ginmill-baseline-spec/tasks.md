# Tasks

## 1. 修正測試盲點與補強 `Features` 單元測試

- [x] 1.1 修正 `ginmill_test.go` 的測試輔助函式 `routes()`（正確回傳 `e.Routes()`、修正未使用參數 `_ *gin.Context`），並以表驅動測試涵蓋 `NewFeatures` 非空、空切片與 `nil` 輸入及 `GetRoutes` 防禦性複本不污染內部狀態；執行 `make test-race` 驗證全數通過
- [x] 1.2 消除 `ginmill.go` 與 `ginmill_test.go` 既有格式與 lint 問題（`gofumpt` 空行、`gocritic` `initClause`），執行 `make lint` 確認 0 issues

## 2. 補齊 `Server.With` 真實 HTTP 行為與邊界測試

- [x] 2.1 新增以 `net/http/httptest` 驗證單一 `Features` 掛載、含 `r.Group("/api/v1")` 與路徑參數 `:id` 的路由分派，以及掛載空 `Features` 不影響既有路由的測試；執行 `make test-race` 驗證通過
- [x] 2.2 更新多組掛載測試為鏈式掛載兩組不衝突 `Features`（`s.With(f1).With(f2)`）並以 `httptest` 驗證兩組路由皆正常回應；另新增測試驗證重複掛載相同 `Method + Path` 觸發 Gin 路由衝突 panic；執行 `make test-race` 驗證通過
- [x] 2.3 新增「同一份共享路由契約分別綁定兩套不同業務實作並掛載於兩個獨立伺服器實例，各自回應且互不污染」的 `httptest` 測試；執行 `make test-race` 驗證通過

## 3. 整合與品質閘門驗證

- [x] 3.1 驗證 `go.mod` 直接第三方依賴僅含 `github.com/gin-gonic/gin`，並執行 `make check`、`make coverage-check` 與 `GOWORK=off make check` 確認格式、vet、lint、`-race` 測試與覆蓋率全數通過
- [x] 3.2 對完整 `git diff` 執行 `code-review-and-quality`（5-axes review）並解決所有 Critical / High 項目

## Workflow follow-up

- 審查通過後執行 OpenSpec `archive`，將 `route-composition` 正式寫入 `openspec/specs/route-composition/spec.md`。
- 執行 `openspec validate --all --strict` 驗證歸檔結果。
