// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"cmp"
	"fmt"
	"math"
	"os"
	"runtime"
	"slices"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	imglfw "github.com/AllenDang/cimgui-go/impl/glfw"
	"github.com/AllenDang/cimgui-go/impl/opengl3"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

func init() {
	runtime.LockOSThread()
}

var (
	sContext          = newSampleContext()
	sRightMouseDown   = false
	sClickPointWS     box2d.Pos
	sFramebufferScale float32 = 1.0
	sImguiWindow      *imglfw.GLFWwindow
)

func sortSamples() {
	slices.SortStableFunc(sampleEntries, func(a, b SampleEntry) int {
		return cmp.Or(cmp.Compare(a.category, b.category), cmp.Compare(a.name, b.name))
	})
	replayIndex = slices.IndexFunc(sampleEntries, func(e SampleEntry) bool { return e.replay })
}

func applyUIStyle() {
	style := imgui.CurrentStyle()

	style.SetWindowPadding(imgui.Vec2{X: 10.0, Y: 10.0})
	style.SetFramePadding(imgui.Vec2{X: 8.0, Y: 4.0})
	style.SetCellPadding(imgui.Vec2{X: 6.0, Y: 4.0})
	style.SetItemSpacing(imgui.Vec2{X: 8.0, Y: 7.0})
	style.SetItemInnerSpacing(imgui.Vec2{X: 7.0, Y: 4.0})
	style.SetIndentSpacing(18.0)
	style.SetScrollbarSize(12.0)
	style.SetGrabMinSize(10.0)

	style.SetWindowBorderSize(1.0)
	style.SetFrameBorderSize(0.0)
	style.SetPopupBorderSize(1.0)
	style.SetTabBorderSize(0.0)
	style.SetSeparatorTextBorderSize(1.0)

	style.SetWindowRounding(4.0)
	style.SetChildRounding(4.0)
	style.SetPopupRounding(4.0)
	style.SetFrameRounding(3.0)
	style.SetGrabRounding(3.0)
	style.SetScrollbarRounding(3.0)
	style.SetTabRounding(3.0)

	style.SetWindowTitleAlign(imgui.Vec2{X: 0.0, Y: 0.5})

	accent := imgui.Vec4{X: 0.28, Y: 0.48, Z: 0.66, W: 1.00}
	accentHi := imgui.Vec4{X: 0.38, Y: 0.60, Z: 0.80, W: 1.00}
	accentLo := imgui.Vec4{X: 0.22, Y: 0.36, Z: 0.50, W: 1.00}

	c := style.Colors()
	c[imgui.ColText] = imgui.Vec4{X: 0.90, Y: 0.91, Z: 0.93, W: 1.00}
	c[imgui.ColTextDisabled] = imgui.Vec4{X: 0.49, Y: 0.51, Z: 0.55, W: 1.00}
	c[imgui.ColWindowBg] = imgui.Vec4{X: 0.110, Y: 0.115, Z: 0.125, W: 0.97}
	c[imgui.ColChildBg] = imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 0.00}
	c[imgui.ColPopupBg] = imgui.Vec4{X: 0.100, Y: 0.105, Z: 0.115, W: 0.98}
	c[imgui.ColBorder] = imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 0.45}
	c[imgui.ColBorderShadow] = imgui.Vec4{X: 0.00, Y: 0.00, Z: 0.00, W: 0.00}
	c[imgui.ColFrameBg] = imgui.Vec4{X: 0.18, Y: 0.19, Z: 0.21, W: 1.00}
	c[imgui.ColFrameBgHovered] = imgui.Vec4{X: 0.24, Y: 0.26, Z: 0.29, W: 1.00}
	c[imgui.ColFrameBgActive] = imgui.Vec4{X: 0.29, Y: 0.32, Z: 0.36, W: 1.00}
	c[imgui.ColTitleBg] = imgui.Vec4{X: 0.090, Y: 0.095, Z: 0.105, W: 1.00}
	c[imgui.ColTitleBgActive] = imgui.Vec4{X: 0.14, Y: 0.16, Z: 0.19, W: 1.00}
	c[imgui.ColTitleBgCollapsed] = imgui.Vec4{X: 0.090, Y: 0.095, Z: 0.105, W: 0.75}
	c[imgui.ColMenuBarBg] = imgui.Vec4{X: 0.13, Y: 0.14, Z: 0.16, W: 1.00}
	c[imgui.ColScrollbarBg] = imgui.Vec4{X: 0.06, Y: 0.06, Z: 0.07, W: 0.55}
	c[imgui.ColScrollbarGrab] = imgui.Vec4{X: 0.28, Y: 0.30, Z: 0.33, W: 1.00}
	c[imgui.ColScrollbarGrabHovered] = imgui.Vec4{X: 0.36, Y: 0.39, Z: 0.43, W: 1.00}
	c[imgui.ColScrollbarGrabActive] = accent
	c[imgui.ColCheckMark] = accentHi
	c[imgui.ColSliderGrab] = accent
	c[imgui.ColSliderGrabActive] = accentHi
	c[imgui.ColButton] = imgui.Vec4{X: 0.22, Y: 0.24, Z: 0.27, W: 1.00}
	c[imgui.ColButtonHovered] = accentLo
	c[imgui.ColButtonActive] = accent
	c[imgui.ColHeader] = imgui.Vec4{X: 0.19, Y: 0.21, Z: 0.24, W: 1.00}
	c[imgui.ColHeaderHovered] = accentLo
	c[imgui.ColHeaderActive] = accent
	c[imgui.ColSeparator] = imgui.Vec4{X: 1.00, Y: 1.00, Z: 1.00, W: 0.09}
	c[imgui.ColSeparatorHovered] = accentLo
	c[imgui.ColSeparatorActive] = accent
	c[imgui.ColResizeGrip] = imgui.Vec4{X: 1.00, Y: 1.00, Z: 1.00, W: 0.06}
	c[imgui.ColResizeGripHovered] = accentLo
	c[imgui.ColResizeGripActive] = accent
	c[imgui.ColTab] = imgui.Vec4{X: 0.15, Y: 0.16, Z: 0.18, W: 1.00}
	c[imgui.ColTabHovered] = accentLo
	c[imgui.ColTabSelected] = accent
	c[imgui.ColTabSelectedOverline] = accentHi
	c[imgui.ColTabDimmed] = imgui.Vec4{X: 0.12, Y: 0.13, Z: 0.14, W: 1.00}
	c[imgui.ColTabDimmedSelected] = accentLo
	c[imgui.ColTextSelectedBg] = imgui.Vec4{X: accent.X, Y: accent.Y, Z: accent.Z, W: 0.40}
	c[imgui.ColDragDropTarget] = accentHi
	c[imgui.ColNavCursor] = accentHi
	c[imgui.ColPlotLines] = imgui.Vec4{X: 0.70, Y: 0.72, Z: 0.75, W: 1.00}
	c[imgui.ColPlotLinesHovered] = accentHi
	c[imgui.ColPlotHistogram] = accent
	c[imgui.ColPlotHistogramHovered] = accentHi
	style.SetColors(&c)
}

func createUI(window *glfw.Window) {
	imgui.CreateContext()
	implot.CreateContext()

	sImguiWindow = imglfw.NewGLFWwindowFromC((*struct{})(unsafe.Pointer(window.Handle())))
	if !imglfw.InitForOpenGL(sImguiWindow, false) {
		panic("ImGui_ImplGlfw_InitForOpenGL failed")
	}

	if !opengl3.Init() {
		panic("ImGui_ImplOpenGL3_Init failed")
	}

	io := imgui.CurrentIO()
	applyUIStyle()

	style := imgui.CurrentStyle()
	style.ScaleAllSizes(sContext.uiScale)
	style.SetFontSizeBase(float32(math.Floor(float64(13.0 * sContext.uiScale))))

	if sContext.uiScale == 1.0 && sFramebufferScale == 1.0 {
		io.Fonts().AddFontDefaultBitmap()
	} else {
		io.Fonts().AddFontDefaultVector()
	}
}

func destroyUI() {
	opengl3.Shutdown()
	imglfw.Shutdown()
	implot.DestroyContext()
	imgui.DestroyContext()
}

func resizeWindowCallback(_ *glfw.Window, width, height int) {
	sContext.camera.width = float32(width)
	sContext.camera.height = float32(height)
}

func keyCallback(
	window *glfw.Window,
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	imglfw.KeyCallback(sImguiWindow, int32(key), int32(scancode), int32(action), int32(mods))
	if imgui.CurrentIO().WantCaptureKeyboard() {
		return
	}

	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape:
			if sContext.showControls {
				sContext.showControls = false
			} else if sContext.sample != nil {
				sContext.sample.Keyboard(key, action, mods)
			}

		case glfw.KeyQ:
			if mods == glfw.ModControl {
				sContext.window.SetShouldClose(true)
			} else if sContext.sample != nil {
				sContext.sample.Keyboard(key, action, mods)
			}

		case glfw.KeyLeft, glfw.KeyRight, glfw.KeyDown, glfw.KeyUp, glfw.KeyZ, glfw.KeyX:

		case glfw.KeyHome:
			if sContext.sample != nil {
				sContext.sample.FocusHome()
			}

		case glfw.KeyR:
			SelectSample(sContext, sContext.sampleIndex, true)

		case glfw.KeyO:
			if mods == glfw.ModControl {
				sContext.showUI = true
				sContext.openSamplePicker = true
			} else if sContext.sample != nil {
				sContext.sample.Keyboard(key, action, mods)
			}

		case glfw.KeyPeriod:
			if mods&glfw.ModShift != 0 {
				sContext.singleStep += 5
			} else {
				sContext.singleStep += 1
			}

		case glfw.KeyP, glfw.KeyPause:
			sContext.pause = !sContext.pause

		case glfw.KeyLeftBracket:
			selection := sContext.sampleIndex - 1
			if selection < 0 {
				selection = len(sampleEntries) - 1
			}
			SelectSample(sContext, selection, false)

		case glfw.KeyRightBracket:
			selection := sContext.sampleIndex + 1
			if selection == len(sampleEntries) {
				selection = 0
			}
			SelectSample(sContext, selection, false)

		case glfw.KeyTab:
			sContext.showUI = !sContext.showUI

		case glfw.KeyM:
			sContext.showMetrics = !sContext.showMetrics

		case glfw.KeyI:
			sContext.showProfile = !sContext.showProfile

		default:
			if sContext.sample != nil {
				sContext.sample.Keyboard(key, action, mods)
			}
		}
	}
}

func updateCameraKeys(frameTime float32) {
	if imgui.CurrentIO().WantCaptureKeyboard() {
		return
	}

	window := sContext.window
	camera := &sContext.camera

	dt := min(frameTime, 0.1)

	zoomRate := float32(0.3)
	if window.GetKey(glfw.KeyZ) == glfw.Press {
		camera.zoom = min(
			float32(math.Exp(float64(zoomRate*dt)))*camera.zoom,
			max(camera.zoom, 100.0),
		)
	} else if window.GetKey(glfw.KeyX) == glfw.Press {
		camera.zoom = max(float32(math.Exp(float64(-zoomRate*dt)))*camera.zoom, 0.5)
	}

	panStep := camera.zoom * dt
	var pan box2d.Vec2
	if window.GetKey(glfw.KeyLeft) == glfw.Press {
		pan.X -= panStep
	}
	if window.GetKey(glfw.KeyRight) == glfw.Press {
		pan.X += panStep
	}
	if window.GetKey(glfw.KeyDown) == glfw.Press {
		pan.Y -= panStep
	}
	if window.GetKey(glfw.KeyUp) == glfw.Press {
		pan.Y += panStep
	}
	camera.center = camera.center.Add(pan)
}

func charCallback(window *glfw.Window, c rune) {
	imglfw.CharCallback(sImguiWindow, uint32(c))
	if imgui.CurrentIO().WantCaptureKeyboard() {
		return
	}

	if c == '?' {
		sContext.showControls = !sContext.showControls
	}
}

func mouseButtonCallback(
	window *glfw.Window,
	button glfw.MouseButton,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	imglfw.MouseButtonCallback(sImguiWindow, int32(button), int32(action), int32(mods))

	if imgui.CurrentIO().WantCaptureMouse() {
		return
	}

	xd, yd := window.GetCursorPos()
	ps := box2d.Vec2{X: float32(xd), Y: float32(yd)}

	if button == glfw.MouseButton1 {
		pw := ConvertScreenToWorld(&sContext.camera, ps)
		if action == glfw.Press {
			sContext.sample.MouseDown(pw, button, mods)
		}

		if action == glfw.Release {
			sContext.sample.MouseUp(pw, button)
		}
	} else if button == glfw.MouseButton2 {
		if action == glfw.Press {
			sClickPointWS = ConvertScreenToWorld(&sContext.camera, ps)
			sRightMouseDown = true
		}

		if action == glfw.Release {
			sRightMouseDown = false
		}
	}
}

func mouseMotionCallback(window *glfw.Window, xd, yd float64) {
	ps := box2d.Vec2{X: float32(xd), Y: float32(yd)}

	imglfw.CursorPosCallback(sImguiWindow, float64(ps.X), float64(ps.Y))

	pw := ConvertScreenToWorld(&sContext.camera, ps)
	sContext.sample.MouseMove(pw)

	if sRightMouseDown {
		diff := pw.Sub(sClickPointWS)
		sContext.camera.center = sContext.camera.center.Add(diff.Neg())
		sClickPointWS = ConvertScreenToWorld(&sContext.camera, ps)
	}
}

func scrollCallback(window *glfw.Window, dx, dy float64) {
	imglfw.ScrollCallback(sImguiWindow, dx, dy)
	if imgui.CurrentIO().WantCaptureMouse() {
		return
	}

	xd, yd := window.GetCursorPos()
	ps := box2d.Vec2{X: float32(xd), Y: float32(yd)}
	pw1 := ConvertScreenToWorld(&sContext.camera, ps)

	if dy > 0 {
		sContext.camera.zoom /= 1.1
	} else {
		sContext.camera.zoom *= 1.1
	}

	pw2 := ConvertScreenToWorld(&sContext.camera, ps)

	pan := pw2.Sub(pw1)
	sContext.camera.center = sContext.camera.center.Add(pan.Neg())
}

func main() {
	os.Exit(run())
}

func run() int {
	sortSamples()

	sContext.Load()

	sContext.showControls = sContext.newUser

	sContext.workerCount = int32(min(8, scenes.GetNumberOfCores()/2))

	if len(os.Args) > 1 && replayIndex >= 0 {
		sContext.replayFile = os.Args[1]
		sContext.sampleIndex = replayIndex
	}

	if err := glfw.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize GLFW: %v\n", err)
		return -1
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	glfw.WindowHint(glfw.Samples, 4)

	version := box2d.CurrentVersion()
	precision := "single"
	if box2d.IsDoublePrecision() {
		precision = "double"
	}
	title := fmt.Sprintf(
		"Box2D Version %d.%d.%d - %s precision",
		version.Major,
		version.Minor,
		version.Revision,
		precision,
	)

	if primaryMonitor := glfw.GetPrimaryMonitor(); primaryMonitor != nil {
		contentScale, _ := primaryMonitor.GetContentScale()

		if runtime.GOOS == "darwin" {
			sContext.uiScale = 1.0
			sFramebufferScale = contentScale
		} else {
			sContext.uiScale = contentScale
			sFramebufferScale = 1.0
		}
	}

	window, err := glfw.CreateWindow(
		int(sContext.camera.width),
		int(sContext.camera.height),
		title,
		nil,
		nil,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open GLFW window: %v\n", err)
		glfw.Terminate()
		return -1
	}
	sContext.window = window

	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize OpenGL: %v\n", err)
		glfw.Terminate()
		return -1
	}

	{
		glVersionString := gl.GoStr(gl.GetString(gl.VERSION))
		glslVersionString := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
		fmt.Printf("OpenGL %s, GLSL %s\n", glVersionString, glslVersionString)
	}

	window.SetSizeCallback(resizeWindowCallback)
	window.SetKeyCallback(keyCallback)
	window.SetCharCallback(charCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)
	window.SetCursorPosCallback(mouseMotionCallback)
	window.SetScrollCallback(scrollCallback)

	createUI(window)
	sContext.draw = CreateDraw()

	sContext.sampleIndex = min(max(sContext.sampleIndex, 0), len(sampleEntries)-1)
	SelectSample(sContext, sContext.sampleIndex, false)

	gl.ClearColor(0.2, 0.2, 0.2, 1.0)

	var frameTime float32

	for !window.ShouldClose() {
		time1 := glfw.GetTime()

		updateCameraKeys(frameTime)

		width, height := window.GetSize()
		sContext.camera.width = float32(width)
		sContext.camera.height = float32(height)

		bufferWidth, bufferHeight := window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(bufferWidth), int32(bufferHeight))

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		opengl3.NewFrame()
		imglfw.NewFrame()

		io := imgui.CurrentIO()
		io.SetDisplaySize(imgui.Vec2{X: sContext.camera.width, Y: sContext.camera.height})

		if sContext.camera.width > 0.0 && sContext.camera.height > 0.0 {
			io.SetDisplayFramebufferScale(imgui.Vec2{
				X: float32(bufferWidth) / sContext.camera.width,
				Y: float32(bufferHeight) / sContext.camera.height,
			})
		}

		imgui.NewFrame()

		sContext.sample.base().ResetText()

		if !sContext.showUI {
			sContext.sample.base().DrawHud(frameTime)
		}

		sContext.draw.SetOrigin(sContext.camera.center)

		sContext.sample.Step()

		sContext.draw.Flush(&sContext.camera)

		if sContext.showUI {
			DrawUI(sContext, frameTime)
		}

		imgui.Render()
		opengl3.RenderDrawData(imgui.CurrentDrawData())

		window.SwapBuffers()

		glfw.PollEvents()

		time2 := glfw.GetTime()
		targetTime := time1 + 1.0/60.0
		for time2 < targetTime {
			runtime.Gosched()
			time2 = glfw.GetTime()
		}

		frameTime = float32(time2 - time1)
	}

	sContext.sample.Destroy()

	sContext.draw.Destroy()

	destroyUI()
	glfw.Terminate()

	sContext.Save()

	return 0
}
