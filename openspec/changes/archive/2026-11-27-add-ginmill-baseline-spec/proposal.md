# Proposal

## Why

`ginmill` 是 `go-kittyd` 與後續 Go 服務共用的上游 Gin 路由組合基礎套件，但目前尚未建立 OpenSpec Living Spec，且既有測試輔助函式 `routes()` 存在未回傳路由清單的盲點，導致實際 HTTP 路由掛載與 `make check` 品質閘門尚未被真實覆蓋。在進行任何功能演進前，需先依現況建立單一、清晰的基線規格與完整行為測試，確保上游契約穩定。

## What Changes

- 依據 [ginmill.go](file:///home/antigravity/gjchentw/ginmill/ginmill.go) 現況建立 `ginmill` 套件唯一的基線能力規格（`route-composition`），完整定義 `Features` 封裝、不可變讀取（防禦性複本）、`Server.With` 鏈式掛載與實際 HTTP 分派行為，以及現有邊界限制（不修改公開 API 簽章、零破壞性變更）。
- 補齊實作驗證與品質基線：修正既有測試中 `routes()` 回傳空切片的盲點，新增以 `net/http/httptest` 驗證真實請求分派與多組 `Features` 鏈式組合的測試，並消除既有 3 個 lint 提示，使 `make check` 與 `GOWORK=off make check` 全綠通過。

## Capabilities

### New Capabilities
- `route-composition`: 將 HTTP 路由規格與業務實作解耦，使相容介面的路由契約成為可獨立發佈、跨實作重用且可鏈式組裝進 Gin 伺服器的共享資源，並提供不可變的路由清單檢視與透明衝突邊界。

### Modified Capabilities
<!-- 無既有能力變更 -->

## Impact

- **受影響檔案**：[ginmill.go](file:///home/antigravity/gjchentw/ginmill/ginmill.go)（僅格式整理，不變動公開 API 行為與簽章）、[ginmill_test.go](file:///home/antigravity/gjchentw/ginmill/ginmill_test.go)（修正既有測試盲點並補齊 `httptest` 真實請求驗證）。
- **公開 API 相容性**：完全維持現有 `Ginmill`、`Server`、`Features`、`NewFeatures`、`(*Features).GetRoutes`、`(*Server).With` 簽章與語意，對下游 `go-kittyd` 零破壞性影響。
- **依賴項目**：不新增任何第三方相依（維持僅依賴 `github.com/gin-gonic/gin`）。
