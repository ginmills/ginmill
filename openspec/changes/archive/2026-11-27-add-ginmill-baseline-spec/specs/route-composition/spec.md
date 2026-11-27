# Spec Delta

## Purpose

將 HTTP 路由規格與業務實作解耦，使路由契約成為可獨立發佈、跨實作重用且可自由組裝的共享資源（Feature），讓任何實作相容介面的 Gin 服務皆能直接掛載使用，同時提供不可變的路由清單檢視與透明的 Gin 原生衝突邊界。

## ADDED Requirements

### Requirement: 路由規格與業務實作解耦（共享資源封裝）
系統 SHALL 允許共享模組在不綁定目標伺服器實例與具體業務實作的前提下，將由相容處理介面組成的路由集合（含空集合或 `nil`、群組前綴與路徑參數）封裝為獨立且可傳遞的共享路由資源（`Features`）。

#### Scenario: 封裝由相容介面綁定的共享路由契約
- **WHEN** 共享模組定義處理介面並透過暫時 `gin.Engine` 宣告路由（含群組前綴或路徑參數），再以 `NewFeatures(r.Routes())` 封裝
- **THEN** 回傳非 nil 的 `*Features`，完整保存對應的 HTTP Method、Path 與 HandlerFunc，且不依賴任何目標伺服器實例

#### Scenario: 封裝空集合或 nil 路由集合
- **WHEN** 呼叫端以空的 `gin.RoutesInfo{}` 或 `nil` 封裝為 `Features`
- **THEN** 回傳非 nil 的 `*Features`，且其包含的路由數量為 0

### Requirement: 同一路由契約可跨不同業務實作重用
系統 SHALL 支援將同一份共享路由契約分別綁定至不同業務實作，並各自掛載於獨立的伺服器實例；各伺服器在處理相同路徑的 HTTP 請求時 MUST 各自執行對應實作的邏輯且互不污染。

#### Scenario: 同一路由契約套用兩套不同業務實作
- **WHEN** 業務實作 A 與業務實作 B 皆實作同一個共享路由契約介面，分別透過同一份 Feature 定義函式產生 `Features` 並掛載至兩個獨立的 Gin 伺服器
- **THEN** 對兩台伺服器送出相同路徑的 HTTP 請求時，伺服器 A 回傳實作 A 的結果、伺服器 B 回傳實作 B 的結果，彼此互不干擾

### Requirement: 主程式可組裝單一或多組獨立共享路由資源
系統 SHALL 支援以鏈式呼叫將單一或多組來自不同共享模組且路徑不衝突的路由資源掛載至同一個 Gin 伺服器，掛載後所有路由 MUST 共存並能正確分派真實 HTTP 請求（含群組前綴與路徑參數）；掛載空資源 MUST NOT 影響既有路由。

#### Scenario: 掛載含群組前綴與路徑參數的共享資源
- **WHEN** 主程式將含 `r.Group("/api/v1")` 與 `GET /statuses/:id` 的共享資源掛載至 `Server.Engine`，並送出 `GET /api/v1/statuses/42` 請求
- **THEN** `With` 回傳原本的 `*Server` 指標，且對應處理函式接收到參數 `id="42"` 並回傳預期回應

#### Scenario: 鏈式組裝多組來自不同模組的共享資源
- **WHEN** 主程式透過 `s.With(f1).With(f2)` 連續掛載兩組路徑不衝突的共享資源（例如 `GET /alpha` 與 `POST /beta`）
- **THEN** 對 `Server.Engine` 送出 `GET /alpha` 與 `POST /beta` 皆由各自對應的處理函式成功回應

#### Scenario: 掛載空共享資源不影響既有路由
- **WHEN** `Server.Engine` 已有路由，再以空 `Features` 呼叫 `s.With(empty)`
- **THEN** 不發生錯誤或 panic，既有路由維持正常運作

### Requirement: 共享路由資源可安全檢視且不被外部修改污染
系統 SHALL 提供共享路由資源的路由清單檢視能力（`GetRoutes`），並回傳新配置的防禦性複本；呼叫端修改回傳切片的元素或欄位 MUST NOT 污染原共享資源內部保存的路由定義。

#### Scenario: 修改檢視回傳的路由切片不污染原共享資源
- **WHEN** 呼叫端取得 `got := f.GetRoutes()` 後修改 `got[0].Path` 與 `got[0].Method`
- **THEN** 再次呼叫 `f.GetRoutes()` 取得的路由仍維持原始定義不變

#### Scenario: 檢視空共享資源回傳空切片
- **WHEN** 對不含任何路由的 `Features` 呼叫 `GetRoutes()`
- **THEN** 回傳非 nil 且長度為 0 的 `gin.RoutesInfo`

### Requirement: 路由衝突與處理鏈邊界透明遵循 Gin 原生語意
共享路由資源掛載時僅註冊各路由攜帶的單一處理函式（不搬運定義階段暫時引擎上的 group middleware）；當同一 HTTP Method 與 Path 重複掛載至同一伺服器時，系統 SHALL 直接遵循底層 `gin.Engine.Handle` 的路由衝突語意（觸發 panic），不得靜默覆蓋。

#### Scenario: 僅執行路由處理函式而不搬運暫時引擎的 group middleware
- **WHEN** 定義階段於暫時 `gin.Engine` 的 `Group` 上附加 middleware 並註冊路由，掛載至目標伺服器後送出請求
- **THEN** 目標伺服器僅執行路由本身的 `HandlerFunc`，暫時引擎的 group middleware 不會被觸發

#### Scenario: 重複掛載相同 Method 與 Path 觸發路由衝突
- **WHEN** 呼叫端對同一個 `Server` 連續兩次掛載包含相同 `GET /dup` 路由的 `Features`
- **THEN** 第二次掛載遵循 Gin 路由樹規則觸發 panic，不會靜默覆蓋既有路由

### Requirement: 維持上游共享契約穩定與極簡依賴
作為多專案共用的上游基礎庫，系統 SHALL 維持既有公開契約（`Ginmill`、`Server`、`Features`、`NewFeatures`、`(*Features).GetRoutes`、`(*Server).With`）之向後相容性，且直接第三方依賴 MUST 僅限於 `github.com/gin-gonic/gin`，不得引入業務框架或無關套件。

#### Scenario: 下游既有共享模組無須修改簽章即可整合
- **WHEN** 下游模組以標準共享契約模式（介面定義 → `NewFeatures(r.Routes())` → `Server.With`）整合 `ginmill`
- **THEN** 無須修改既有呼叫端簽章即可編譯並正確路由請求
