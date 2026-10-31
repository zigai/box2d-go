# Samples

An interactive app that runs about 150 physics scenes on the Go port, ported from Box2D's C samples app. Use it to see how the engine behaves, try ideas, and compare against the C version by eye.

- **Scenes:** stacking, joints, shapes, collision, continuous collision, events, character movement, determinism, and benchmarks.
- **Live controls:** pause, single-step, restart, drag bodies with the mouse, and tune each scene from its panel.
- **Replays:** record a session and play it back with seeking.
- **Precision:** runs on the float build or the large-world double build.

## Usage

Run from the repository root:

```sh
just samples
```

Run the double-precision build:

```sh
just samples double
```

The app uses OpenGL 3.3 through GLFW and Dear ImGui, so it needs a C compiler with `CGO_ENABLED=1`.

Pass a recording file after the precision, relative to `samples/`, to open it in the replay sample:

```sh
just samples float path/to/recording.b2rec
```

## Controls

| Input | Action |
| --- | --- |
| Left drag | Pull a body |
| Right drag, arrow keys | Pan |
| Scroll, Z, X | Zoom |
| P | Pause or resume |
| . | Step once; Shift+. steps five times |
| R | Restart the scene |
| [ and ] | Previous or next scene |
| Ctrl+O | Pick a scene |
| Home | Reset the view |
| Tab | Show or hide the panels |
| M, I | Show metrics, show step timings |
| Ctrl+Q | Quit |
