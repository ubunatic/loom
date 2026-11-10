## Item tml-declarative (research + experiment)

### Verdict
Loom should adopt a small Go builder API or minimal JSON format rather than a full XML dialect, because TUIOS requires a complex parsing layer (`tml`, ~5k loc) while the zero-dependency JSON prototype proves tree-building is trivial in pure Go.

### TUIOS findings
TUIOS provides a declarative XML-based layer named `tml` which generates a UI tree rendered by `loom` components (e.g. `<split>`, `<tabs>`).
- Layout nodes are defined explicitly, for instance `layout.tml` which nests `<grid>`, `<stack>`, `<split>`, `<row>`, `<col>` (permalink: `https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/tml/testdata/examples/layout.tml#L3-L9`).
- Attributes handle static configuration (e.g., `ratio="0.4"` on `<split>`) and dynamic bindings using JSON state like `{"passed": 412}` passed into variables (permalink: `https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/tml/testdata/examples/dashboard.tml#L2-L13`).
- TUIOS translates XML strings into `*tml.Node` ASTs (`https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/tml/xml.go#L49`) which are passed to the `loom` Surface builder (`https://github.com/Gaurav-Gosain/tuios/blob/6fe2c0df9922b2eec28f392fee4e196efd86be36/loom/surface.go#L154`).
- Loom currently constructs widgets imperatively, passing child widgets to constructors like `func NewSplit(first, second Widget) *Split` (`split.go:43`) and `func NewTabs(tabs ...Tab) *Tabs` (`tabs.go:109`).

### Prototype
- **Format:** JSON declarative object tree matching widget constructor arguments.
- **Files:** `experiments/declarative/prototype.go` (defines `Node` struct and `Build` recursive function), `experiments/declarative/prototype_test.go` (tests).
- **What works:** Recursive rendering of a nested Split (from `examples/split`) and a multi-tab view (from `examples/tabs`) matching the imperative loom builders exactly, verified by checking terminal output equality on a fixed `Canvas`.

### Tests
- `go test ./experiments/declarative/... -v`
- Result:
```
=== RUN   TestDummy
--- PASS: TestDummy (0.00s)
=== RUN   TestDeclarativeSplit
--- PASS: TestDeclarativeSplit (0.00s)
=== RUN   TestDeclarativeTabs
--- PASS: TestDeclarativeTabs (0.00s)
PASS
ok      ubunatic.com/loom/experiments/declarative       0.026s
```

### Recommendation
Loom should avoid adopting a full-blown custom XML parser due to complexity and footprint size, but could adopt a minimal declarative layer. The smallest first step is to implement a Go builder API that chains nested structures (e.g. `loom.BuildSplit(loom.BuildView(...), loom.BuildView(...))`) to streamline UI definition without text parsing overhead.
