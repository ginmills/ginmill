# Design

## Context

`ginmill` 核心實作位於單一檔案 [ginmill.go](file:///home/antigravity/gjchentw/ginmill/ginmill.go)（48 行），提供極簡的 Gin 路由契約封裝與掛載機制（詳見 `proposal.md` — Why）。本次變更旨在依據現況建立唯一的 baseline capability spec（`route-composition`），並在零破壞性變更前提下補齊真實 HTTP 行為測試與通過 `make check`。

經盤點既有程式碼與測試，確認三個關鍵現況事實：
1. **現有測試輔助函式盲點**：[ginmill_test.go](file:///home/antigravity/gjchentw/ginmill/ginmill_test.go#L10-L16) 的 `routes()` 宣告了 named return `(r gin.RoutesInfo)`，卻未將 `e.Routes()` 指派給 `r`，導致過去所有呼叫 `routes()` 的測試實際都在測試空切片（也因此 `TestServer_With_MultipleCalls` 重複呼叫 `s.With(f)` 才未觸發 Gin 重複註冊 panic）。
2. **`Ginmill` 介面與 `(*Server).With` 的簽章差異**：`Ginmill` 介面定義 `With(f *Features) (err error)`，而 `(*Server).With(f *Features) *Server` 回傳 `*Server` 以支援鏈式呼叫（且 `Server` 內嵌 `Ginmill`）。下游（如 `go-kittyd`）皆直接使用 `&ginmill.Server{Engine: r}` 搭配 `s.With(f)`，因此基線階段必須完整保留此簽章以維持 SemVer 向後相容。
3. **既有 Lint 項目**：[ginmill.go](file:///home/antigravity/gjchentw/ginmill/ginmill.go#L37) 有多餘空行（`gofumpt`），[ginmill_test.go](file:///home/antigravity/gjchentw/ginmill/ginmill_test.go#L13) 有未使用參數 `c`（`revive`）與第 129 行 `if initClause`（`gocritic`）。

## Goals / Non-Goals

**Goals:**
- 將 `ginmill` 現有公開行為忠實記錄為單一 capability（`route-composition`）的 Living Spec 基線。
- 修正測試輔助函式盲點，並以 `net/http/httptest` 建立涵蓋單組掛載、群組前綴與路徑參數、多組鏈式掛載、空集合、防禦性複本與重複路由衝突的完整單元與行為測試。
- 消除既有 3 個 lint 提示，使 `make check` 與 `GOWORK=off make check` 皆 100% 通過且覆蓋率維持高水準（目標 ≥ 95%）。

**Non-Goals:**
- 不修改任何公開 API 簽章（例如不在此階段變更 `Ginmill` 介面或讓 `With` 改回傳 `error`）。
- 不新增 middleware 鏈搬運、prefix mount 或 route 去重等新功能（留待後續獨立 OpenSpec change 評估）。
- 不引入任何額外第三方套件。

## Decisions

### Decision 1: 單一 Capability 命名為 `route-composition`
- **選擇**：在 `openspec/specs/route-composition/spec.md` 下維護整個套件唯一的能力規格。
- **理由**：`ginmill` 的單一職責即為「路由定義封裝與組合掛載」，使用領域行為名詞 `route-composition` 符合 OpenSpec capability 命名規範，避免過細拆分造成維護成本。
- **替代方案**：命名為 `ginmill`（與套件同名但較不具行為描述力）或拆分為 `features` 與 `server` 兩個 capability（對 48 行核心而言過度拆分）。

### Decision 2: 基線測試以真實 `httptest` 驗證並區分「多組鏈式」與「重複衝突」
- **選擇**：修正 `routes()` 回傳 `e.Routes()`，將原本因空切片誤過的 `TestServer_With_MultipleCalls` 改為驗證「多組不衝突 `Features` 鏈式掛載 + `httptest` 實際回應」，另新增獨立測試明確驗證「重複註冊相同 `Method + Path` 觸發 Gin panic」。
- **理由**：`ginmill` 沒有包裝或吞掉 Gin 的路由樹規則，明確記錄並測試此邊界可避免未來維護者誤以為 `With` 具有冪等覆寫語意。
- **替代方案**：修改 `With` 實作使其跳過重複路由——這會改變既有執行期行為，超出 baseline spec 範圍，故不採用。

### Decision 3: `ginmill.go` 僅做格式修正，零行為改動
- **選擇**：僅移除 [ginmill.go](file:///home/antigravity/gjchentw/ginmill/ginmill.go#L37) 的多餘空行以滿足 `gofumpt`，不改動任何結構、簽章或邏輯。
- **理由**：確保下游 `go-kittyd` 與現有使用者零風險無痛銜接。

## Risks / Trade-offs

- **[Risk] 既有 `Ginmill` 介面簽章（回傳 `error`）與 `(*Server).With`（回傳 `*Server`）不一致，可能讓新閱讀者困惑** → **Mitigation**：在本設計文件明確記錄此歷史現況為向後相容刻意保留，日後若計畫推出 `v2` 再透過獨立 breaking-change proposal 統一處理。
- **[Risk] `NewFeatures` 未對輸入切片做防禦性複本，若呼叫端在傳入後原地修改傳入的 slice 可能影響內部資料** → **Mitigation**：目前主流用法均直接傳入 `r.Routes()` 新產出的切片；先如實記錄現況，若未來有防禦輸入需求再透過獨立規格微調。
