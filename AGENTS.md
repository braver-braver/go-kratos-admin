# Project Agent Instructions

- 当你需要了解 XXX 技术（例如 openai/codex、Graphiti、MCP 相关仓库）时，
  请优先通过 DeepWiki MCP 查询，而不是凭空猜测。
- 使用策略：
  1. 先调用 DeepWiki 的 `read_wiki_structure` 获取该仓库的文档结构。
  2. 根据用户问题，选择最相关的文档，用 `read_wiki_contents` 拉取内容并阅读。
  3. 对于更开放的问题，使用 `ask_question` 获取基于文档和代码的高质量答案。
- 回答时，请在适当位置说明「此结论来自 DeepWiki 文档」，方便人类复现查询过程。

## Active Technologies
- Go 1.24.6 + go-kratos, gorm generic API, gorm CLI codegen, wire, ent (baseline for parity tests), golangci-lint, testify (001-ent-to-gorm)
- PostgreSQL (compose uses timescaledb pg15); MySQL driver exists in ent baseline—confirm whether dual support must be preserved (001-ent-to-gorm)

## Recent Changes
- 001-ent-to-gorm: Added Go 1.24.6 + go-kratos, gorm generic API, gorm CLI codegen, wire, ent (baseline for parity tests), golangci-lint, testify
