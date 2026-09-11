# Splash Screen Example

This example demonstrates the startup splash screen and transition lifecycle documented in [`HarnezSplashTarget.md`](../../docs/HarnezSplashTarget.md).

From the repository root:

```sh
go run ./examples/splash
go run ./examples/splash --watch
go run ./examples/splash --help
go run ./examples/splash --width 60 --height 15
```

## Features

- **Centered Viewport Layout**: Automatically centers the splash container horizontally and vertically across arbitrary terminal sizes.
- **Braille Spinner & Progress**: Animated braille glyph spinner (`⠋⠙⠹...`) and bracketed braille/pattern progress bar.
- **Provider Status Pills**: Unicode service badges (`● mic  ✳ claude  ֍ codex  Λ agy`) with lifecycle state color styling (pending, fetching, done, failed).
- **Asynchronous Lifecycle**: Coordinates background provider initialization while keeping the animation loop smooth.
- **Interactive Skip**: Press `Esc`, `q`, or `Enter` to dismiss the splash screen immediately.
