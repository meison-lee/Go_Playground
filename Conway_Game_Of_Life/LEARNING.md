# Conway's Game of Life

Conway's Game of Life（康威生命遊戲）是一個由數學家 John Horton Conway 在 1970 年提出的零玩家遊戲。遊戲規則極為簡單，但能產生出極其豐富、不可預測的圖案與行為，被譽為「數學與電腦科學的經典模擬系統」。

## 遊戲玩法簡介

- 整個遊戲在一個二維網格上進行，每個格子可以是「活的」或「死的」。
- 每一輪（generation），每個格子的生死由以下規則決定：
    1. **存活**：如果一個活細胞周圍有 2 或 3 個活鄰居，則下一輪仍然存活。
    2. **死亡**：如果活細胞周圍活鄰居少於 2 個（孤獨）或多於 3 個（擁擠），則死亡。
    3. **誕生**：如果一個死細胞周圍有恰好 3 個活鄰居，則下一輪誕生為活細胞。
- 整個系統根據初始狀態自動演化，無需玩家介入。

你可以任意設計初始狀態，觀察格子如何隨時間變化，形成各種結構、週期或消亡。

## 延伸閱讀

維基百科介紹：  
[https://zh.wikipedia.org/zh-tw/康威生命遊戲](https://zh.wikipedia.org/zh-tw/%E5%BA%B7%E5%A8%81%E7%94%9F%E5%91%BD%E6%B8%B8%E6%88%8F)

---
在這次實作 Conway’s Game of Life 時，我針對 Go 語言的並發（concurrency）特性做了多種嘗試與效能比較，具體心得如下：

1. 初版設計：每格一個 goroutine
最初嘗試將每一個 cell 的狀態計算都交給獨立的 goroutine，200x400 的網格等於一次開 8 萬個 goroutine。
實際效能遠遜於單純 for-loop，主因在於 goroutine 雖然輕量，但數量過多時，Go runtime 需要大量 context switching，產生排程和管理上的顯著開銷，反而成為系統瓶頸。這個實驗驗證了「goroutine 並不是越多越好」，開啟過多 goroutine 會造成資源爭用和效能衰退。

2. 行級並行：每行一個 goroutine
接著將併發單位改成「每一行」開一個 goroutine，減少 goroutine 數量（例如 400 個）。
這樣可以在中小型網格下有效分散計算壓力，效能有明顯提升。然而，當 height、width 持續上升（如 20000 x 10000），goroutine 數仍舊可能大於 CPU 核心數百倍，調度成本依然存在。

3. Worker Pool：依 CPU 核心動態分配 goroutine
最終採用 worker pool 模式：
依照 runtime.NumCPU() 決定要開多少 goroutine（通常等於或略多於實體核心數），讓每個 worker 從 channel 中依序領取 row 進行計算。
這種模式能最大化 CPU 利用率、避免 goroutine 過量造成效能瓶頸，同時動態平衡每個 worker 的工作量。這也是實際大型運算、伺服器應用最常見的併發模式。


