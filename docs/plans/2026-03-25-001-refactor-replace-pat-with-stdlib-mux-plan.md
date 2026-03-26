---
title: "用 Go 标准库 http.ServeMux 替换 pat 路由库"
type: refactor
status: active
date: 2026-03-25
---

# 用 Go 标准库 http.ServeMux 替换 pat 路由库

## 概述

评估并实施用 Go 1.22+ 内置的 `http.ServeMux` 替换第三方路由库 `github.com/bmizerany/pat` 的可行性。

## 现状分析

### 依赖情况

| 文件 | 路由库 |
|------|--------|
| `web/handlers.go:19` | pat v0.0.0-20210406213842-e4b6760bdd6f |
| `apps/imsto-admin/main.go:40` | pat v0.0.0-20210406213842-e4b6760bdd6f |

### pat 使用模式

```go
// web/handlers.go:19-31
mux := pat.New()
mux.Get("/imsto/roofs", http.HandlerFunc(roofsHandler))
mux.Post("/imsto/ticket", CheckAPIKey(http.HandlerFunc(ticketHandlerPost)))
mux.Get("/imsto/ticket", CheckAPIKey(http.HandlerFunc(ticketHandlerGet)))
mux.Post("/imsto/token", CheckAPIKey(http.HandlerFunc(tokenHandler)))
mux.Post("/imsto/:roof", CheckAPIKey(secure(storedHandler)))
mux.Del("/imsto/:roof/:id", CheckAPIKey(secure(deleteHandler)))
mux.Get("/imsto/:roof/id", http.HandlerFunc(GetOrHeadHandler))
mux.Get("/imsto/:roof/metas/count", http.HandlerFunc(countHandler))
mux.Get("/imsto/:roof/metas", http.HandlerFunc(browseHandler))
```

### 参数提取方式

```go
// web/handlers.go - 当前使用 r.URL.Query().Get(":roof")
roof := r.URL.Query().Get(":roof")           // line 72
id, err := imid.ParseID(r.URL.Query().Get(":id"))  // line 170
```

## Go 标准库 mux 能力评估

### 新版 mux 特性 (Go 1.22+)

| 特性 | 支持情况 | 说明 |
|------|----------|------|
| 路径参数 `:name` | ✅ | 语法与 pat 相同 |
| 方法限定路由 | ✅ | GET/POST/DELETE 等 |
| 路径前缀匹配 | ✅ | `/static/` |
| 尾部通配符 `*name` | ✅ | 捕获剩余路径 |
| 路由优先级 | ✅ | 最长匹配优先 |
| 中间件模式 | ✅ | 包装 HandlerFunc |

### 关键差异

| 方面 | pat | 标准库 mux |
|------|-----|------------|
| 参数提取 | `r.URL.Query().Get(":name")` | `r.PathValue("name")` (无冒号) |
| 方法路由 | `mux.Get()` `mux.Post()` | `mux.HandleFunc("GET", pattern, fn)` |
| 依赖 | 需引入第三方 | 标准库内置 |

## 迁移方案

### 改动点清单

#### 1. web/handlers.go

| 行号 | 原代码 | 改后 |
|------|--------|------|
| 19 | `mux := pat.New()` | `mux := http.NewServeMux()` |
| 20 | `mux.Get("/imsto/roofs", ...)` | `mux.HandleFunc("GET", "/imsto/roofs", ...)` |
| 22-31 | 同上模式 | 同上模式 |

**注意**: 参数提取需从 `r.URL.Query().Get(":roof")` 改为 `r.PathValue("roof")`

#### 2. apps/imsto-admin/main.go

| 行号 | 原代码 | 改后 |
|------|--------|------|
| 40 | `mux := pat.New()` | `mux := http.NewServeMux()` |
| 45-53 | `mux.Get(...)` | `mux.HandleFunc("GET", ...)` |

#### 3. 参数提取改动

需修改以下位置，将 `:name` 改为 `name`:

```go
// web/handlers.go
roof := r.URL.Query().Get(":roof")     // → roof := r.PathValue("roof")
id, err := imid.ParseID(r.URL.Query().Get(":id"))  // → id, err := imid.ParseID(r.PathValue("id"))
```

### 中间件模式兼容性

中间件包装模式保持不变:

```go
// 不变
mux.Post("/imsto/:roof", CheckAPIKey(secure(storedHandler)))

// 标准库写法
mux.HandleFunc("POST", "/imsto/:roof", func(w http.ResponseWriter, r *http.Request) {
    CheckAPIKey(secure(http.HandlerFunc(storedHandler))).ServeHTTP(w, r)
})
```

**建议**: 封装辅助函数以保持代码简洁

```go
func withMiddleware(pattern, method string, handler http.HandlerFunc, mw ...func(http.Handler) http.Handler) {
    var h http.Handler = handler
    for i := len(mw) - 1; i >= 0; i-- {
        h = mw[i](h)
    }
    mux.HandleFunc(method, pattern, h.ServeHTTP)
}
```

## 依赖移除

```go
// go.mod 移除
require github.com/bmizerany/pat v0.0.0-20210406213842-e4b6760bdd6f

// 不再需要 import
// import "github.com/bmizerany/pat"
```

## 测试要点

### 功能测试

- [ ] 所有 GET 路由正常返回
- [ ] 所有 POST 路由正常处理
- [ ] 所有 DELETE 路由正常工作
- [ ] 路径参数正确传递
- [ ] 中间件 (CheckAPIKey, secure) 正常工作

### 边界情况

- [ ] 非法路径返回 404
- [ ] 不支持的方法返回 405
- [ ] 缺少参数时正确处理

## 风险评估

| 风险 | 等级 | 缓解措施 |
|------|------|----------|
| 参数提取 API 变更 | 低 | Go 1.22+ 已稳定 |
| 中间件执行顺序 | 低 | 现有测试覆盖 |
| 路由优先级差异 | 低 | 标准库按最长匹配 |

## 优势

1. **减少依赖**: 移除一个第三方依赖
2. **简化维护**: 使用标准库，减少版本管理
3. **提升性能**: 标准库实现通常经过高度优化
4. **未来兼容**: 随 Go 版本自动获得新特性

## 结论

**可行性评估: ✅ 可行**

- 项目使用 Go 1.25，完全支持新版 `http.ServeMux`
- pat 使用模式简单，迁移成本低
- 中间件架构兼容，无需重构
- 风险可控，建议实施

## 实施建议

1. 先修改 `web/handlers.go` 并测试
2. 再修改 `apps/imsto-admin/main.go`
3. 确认无测试失败后移除 go.mod 依赖
4. 运行完整测试套件验证
