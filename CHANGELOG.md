# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Add AnchorPos struct and functions #252
- Add Clipboard Support
- Fix SIGSEGV on text.NewAtlas if glyph absent 
- Use slice for range in Drawer.Dirty(), to improve performance
- GLTriangle's fragment shader is used when rendered by the Canvas.
- Add MSAA support

## [v0.10.0] 2020-08-22
- Add AnchorPos struct and functions
- Gamepad API added
- Support setting an initial window position
- Support hiding the window initially
- Support creating maximized windows
- Support waiting for events to reduce CPU load
- Adding clipping rectangle support in GLTriangles

## [v0.10.0-beta] 2020-05-10
- Add `WindowConfig.TransparentFramebuffer` option to support window transparency onto the background
- Fixed Line intersects failing on lines passing through (0, 0)

## [v0.10.0-alpha] 2020-05-08
- Upgrade to GLFW 3.3! :tada:
  - Closes https://github.com/faiface/pixel/issues/137
- Add support for glfw's DisableCursor
  - Closes https://github.com/faiface/pixel/issues/213

## [v0.9.0] - 2020-05-02
- Added feature from https://github.com/faiface/pixel/pull/219
  - Exposing Window.SwapBuffers so buffers can be swapped without polling input
- Add more examples
- Add position as out variable from vertex shader
- Add experimental joystick support
- Add mouse cursor operations
- Add `Vec.Floor(…)` function
- Add circle geometry
- Fix `Matrix.Unproject(…)` for rotated matrix
- Add 2D Line geometry
- Add floating point round error correction
- Performance improvements
- Fix race condition in `NewGLTriangles(…)`
- Add `TriangleData` benchmarks and improvements
- Add zero rectangle variable for utility and consistency
- Add support for Go Modules
- Add `NoIconify` and `AlwaysOnTop` window hints


## [v0.8.0] - 2018-10-10
Changelog for this and older versions can be found on the corresponding [GitHub
releases](https://github.com/faiface/pixel/releases).

[Unreleased]: https://github.com/faiface/pixel/compare/v0.10.0...HEAD
[v0.10.0]: https://github.com/faiface/pixel/compare/v0.10.0-beta...v0.10.0
[v0.10.0-beta]: https://github.com/faiface/pixel/compare/v0.10.0-alpha...v0.10.0-beta
[v0.10.0-alpha]: https://github.com/faiface/pixel/compare/v0.9.0...v0.10.0-alpha
[v0.9.0]: https://github.com/faiface/pixel/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/faiface/pixel/releases/tag/v0.8.0
