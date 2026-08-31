# Contributing to seg

Keep changes focused and include tests for behavior changes. Before opening a pull request, run:

```bash
go test ./...
npm --prefix console ci
npm --prefix console run build
```

Do not commit credentials, private data, generated binaries, or local model outputs. By contributing, you agree
that your contribution may be distributed under the repository's MIT License.
