## Verdict
inconclusive, compiling a Go grammar to wasm and running it under wazero requires a full tree-sitter wasm runtime engine (like `malivvan/tree-sitter` which only bundles C/C++ internally), and standard emscripten-compiled tree-sitter grammars expect `env.memory` and other exports not easily mocked in a pure wazero experiment without the full tree-sitter C engine compiled alongside it.

## Approach
- load one tree-sitter grammar (Go) compiled to wasm under wazero: skipped, as compiling `tree-sitter-go` to `.wasm` yields a module expecting Emscripten `env` imports (`memory`, etc.) and the tree-sitter engine itself, which requires either a custom wazero host wrapper for all TS engine functions or modifying a project like `malivvan/tree-sitter` to bundle Go grammar alongside its engine build.

## Tests
- `git clone https://github.com/tree-sitter/tree-sitter-go.git && tree-sitter build --wasm`: successfully compiled `tree-sitter-go.wasm` (217KB).
- ran a custom wazero host `go_highlight.go` attempting to load `tree-sitter-go.wasm`: failed with `module[env] not instantiated`, and subsequently `"memory" is not exported in module "env"` when attempting to mock it, showing the compiled grammar expects the emscripten runtime and tree-sitter core engine.

## Risks
- The approach requires a non-trivial build pipeline to compile both the tree-sitter core engine and the specific grammars into a single WASM module, or complex wazero host module mocking. Until this build system is available, memory and performance metrics for Go parsing cannot be accurately collected.
