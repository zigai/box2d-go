// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

// Names for the inspector readouts
func replayBodyTypeName(bodyType box2d.BodyType) string {
	switch bodyType {
	case box2d.StaticBody:
		return "static"
	case box2d.KinematicBody:
		return "kinematic"
	case box2d.DynamicBody:
		return "dynamic"
	default:
		return "?"
	}
}

func replayShapeTypeName(shapeType box2d.ShapeType) string {
	switch shapeType {
	case box2d.CircleShape:
		return "circle"
	case box2d.CapsuleShape:
		return "capsule"
	case box2d.SegmentShape:
		return "segment"
	case box2d.PolygonShape:
		return "polygon"
	case box2d.ChainSegmentShape:
		return "chain segment"
	default:
		return "?"
	}
}

func replayJointTypeName(jointType box2d.JointType) string {
	switch jointType {
	case box2d.DistanceJointType:
		return "distance"
	case box2d.FilterJointType:
		return "filter"
	case box2d.MotorJointType:
		return "motor"
	case box2d.PrismaticJointType:
		return "prismatic"
	case box2d.RevoluteJointType:
		return "revolute"
	case box2d.WeldJointType:
		return "weld"
	case box2d.WheelJointType:
		return "wheel"
	default:
		return "?"
	}
}

func replayQueryTypeName(queryType box2d.ReplayQueryType) string {
	switch queryType {
	case box2d.ReplayQueryOverlapAABB:
		return "overlap AABB"
	case box2d.ReplayQueryOverlapShape:
		return "overlap shape"
	case box2d.ReplayQueryCastRay:
		return "cast ray"
	case box2d.ReplayQueryCastShape:
		return "cast shape"
	case box2d.ReplayQueryCollideMover:
		return "collide mover"
	case box2d.ReplayQueryCastRayClosest:
		return "cast ray closest"
	case box2d.ReplayQueryCastMover:
		return "cast mover"
	case box2d.ReplayQueryShapeTestPoint:
		return "shape test point"
	case box2d.ReplayQueryShapeRayCast:
		return "shape ray cast"
	default:
		return "?"
	}
}

func shapeIndex(shape box2d.Shape) int {
	return int(int32(shape.ID() >> 32))
}

func bodyIndex(body box2d.Body) int {
	return int(int32(body.ID() >> 32))
}

func hsvColor(h, s, v float32) imgui.Vec4 {
	var r, g, b float32
	imgui.ColorConvertHSVtoRGB(h, s, v, &r, &g, &b)
	return imgui.Vec4{X: r, Y: g, Z: b, W: 1.0}
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// Inspector selection, keyed by stable creation ordinals so it survives a backward scrub. Resolved
// to live ids each frame from the player's body tracking; out of range means "not at this frame".
type replaySelKind int

const (
	selNone replaySelKind = iota
	selBody
	selShape
	selJoint
	selQuery
)

// Plays back a recording by re-running the engine one step at a time and drawing the
// replayed world. Stepping is driven by the recorded inputs, not by World.Step, so the motion
// reproduces the original session exactly. Pause, single step, and restart use the shared sample
// controls. Mouse picking is disabled because dragging a body would mutate the replayed world
// and diverge it from the recording.
type ReplayViewer struct {
	Sample

	player *box2d.Replay
	path   string
	status string

	info              box2d.ReplayInfo
	speed             float32
	frameAccumulator  float32
	loop              bool
	selectTimelineTab bool
	prevShowMetrics   bool

	// Load popup state. A fresh open configures the keyframe policy here, then the popup switches to
	// a progress bar while every keyframe is generated up front. Temporaries hold the in-popup edits
	// so Cancel leaves the persisted settings untouched.
	requestLoadPopup bool
	generating       bool
	popupBudgetMB    int32
	popupMinInterval int32

	selKind         replaySelKind
	selBodyOrdinal  int  // index into the player's tracked body list
	selSlot         int  // shape or joint slot within that body
	selQuery        int  // query index, only meaningful for the current frame
	revealSelection bool // one-shot request to expand and scroll the tree to a viewport pick
}

// The player owns the world we draw, so skip the base world. worldID stays nil until
// CreatePlayer adopts the player's world.
func CreateReplayViewer(context *SampleContext) Sampler {
	s := &ReplayViewer{
		speed:             1.0,
		selectTimelineTab: true,
		popupBudgetMB:     512,
		popupMinInterval:  16,
		selKind:           selNone,
		selBodyOrdinal:    -1,
		selSlot:           -1,
		selQuery:          -1,
	}
	s.init(context, s, false)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 7.5}
		s.context.camera.zoom = 10.0
	}

	// The timeline scrubber lives in the metrics drawer, so open it for the replay
	s.prevShowMetrics = s.context.showMetrics
	s.context.showMetrics = true
	s.context.pause = true
	s.selectTimelineTab = true

	s.path = s.context.replayFile
	s.status = ""
	s.player = nil

	// A fresh open gathers the keyframe policy through the Load popup, then pre-generates every
	// keyframe. A restart reuses the persisted policy and fills the ring lazily so R stays quick.
	if !s.context.restart {
		if len(s.path) > 0 {
			s.requestLoadPopup = true
		} else {
			s.status = "Open recording from Replay menu"
		}
	} else {
		s.CreatePlayer()
	}
	return s
}

func (s *ReplayViewer) Destroy() {
	s.ClosePlayer()

	s.context.showMetrics = s.prevShowMetrics
	s.worldID = nil
}

func (s *ReplayViewer) ClosePlayer() {
	if s.player != nil {
		s.player.Close()
		s.player = nil
	}
	s.worldID = nil
	s.selKind = selNone
	s.selBodyOrdinal = -1
	s.selSlot = -1
	s.selQuery = -1
}

func loadRecordingFromFile(path string) (*box2d.Recording, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return box2d.ReadRecording(file)
}

func (s *ReplayViewer) CreatePlayer() {
	s.ClosePlayer()

	if len(s.path) == 0 {
		s.status = "Open recording from Replay menu"
		return
	}

	// Load the file into a recording buffer, then hand its bytes to the player. Replay workers of 0
	// uses the serial fallback, otherwise force a different count to spot-check cross-thread determinism.
	recording, err := loadRecordingFromFile(s.path)
	if err == nil {
		// Use a large worker count so key frame generation is fast
		s.context.workerCount = int32(min(8, scenes.GetNumberOfCores()/2))
		s.player, err = box2d.NewReplay(recording.Data(), int(s.context.workerCount))
		if err != nil {
			s.player = nil
		}
	} else {
		s.player = nil
	}

	s.frameAccumulator = 0.0
	if s.player != nil {
		s.worldID = s.player.World()
		s.info = s.player.Info()

		// Apply the persisted keyframe policy before any stepping captures keyframes. A freshly
		// created player starts at the engine defaults, so the ring rebuilds under our spacing.
		bytes := uint(s.context.replayKeyframeBudgetMB) * 1024 * 1024
		s.player.SetKeyframePolicy(bytes, int(s.context.replayKeyframeMinInterval))

		s.status = "loaded"

		if !s.context.restart {
			s.FrameRecording()
		}
	} else {
		s.info = box2d.ReplayInfo{}
		s.status = "failed to open file"
	}
}

// Modal shown after the Replay menu picks a file: choose the keyframe budget and min interval,
// then Load creates the player and pre-generates the whole ring behind a progress bar. Drawn at
// the root ID stack from Step, like the sample picker.
func (s *ReplayViewer) DrawLoadPopup() {
	popupID := "Load Replay"

	if s.requestLoadPopup {
		s.requestLoadPopup = false
		s.popupBudgetMB = s.context.replayKeyframeBudgetMB
		s.popupMinInterval = s.context.replayKeyframeMinInterval
		imgui.OpenPopupStr(popupID)
	}

	fontSize := imgui.FontSize()
	imgui.SetNextWindowPosV(
		imgui.Vec2{X: s.context.camera.width * 0.5, Y: s.context.camera.height * 0.35},
		imgui.CondAppearing,
		imgui.Vec2{X: 0.5, Y: 0.5},
	)
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 26.0 * fontSize, Y: 0.0}, imgui.CondAppearing)

	if !imgui.BeginPopupModalV(popupID, nil, imgui.WindowFlagsAlwaysAutoResize) {
		return
	}

	// Show just the file name, paths run long
	name := s.path
	if slash := strings.LastIndex(s.path, "\\"); slash >= 0 {
		name = s.path[slash+1:]
	}
	imTextDisabled("File:")
	imgui.SameLine()
	imgui.TextUnformatted(name)
	imgui.Separator()

	if s.generating {
		// Step forward in wall-clock slices so the bar animates. Forward stepping captures
		// keyframes at the interval; a restart then returns to frame 0 with the ring kept.
		ticks := time.Now()
		for !s.player.IsAtEnd() && millisecondsSince(ticks) < 12.0 {
			s.player.StepFrame()
		}

		frame := s.player.Frame()
		total := 1
		if s.info.FrameCount > 0 {
			total = s.info.FrameCount
		}
		frac := float32(1.0)
		if frame < total {
			frac = float32(frame) / float32(total)
		}
		overlay := fmt.Sprintf("%d / %d", frame, s.info.FrameCount)
		imgui.TextUnformatted("Generating keyframes")
		imgui.ProgressBarV(frac, imgui.Vec2{X: -fltMin, Y: 0.0}, overlay)

		if s.player.IsAtEnd() {
			s.player.Restart()
			s.worldID = s.player.World()
			s.generating = false
			s.context.pause = true
			imgui.CloseCurrentPopup()
		}

		imgui.EndPopup()
		return
	}

	imgui.PushItemWidth(10.0 * fontSize)
	imgui.SliderInt("Memory budget (MB)", &s.popupBudgetMB, 128, 4096)
	imgui.SliderInt("Min sample interval", &s.popupMinInterval, 8, 60)
	imgui.PopItemWidth()

	// Surface a failed open inline so Load can be retried
	if s.status != "" && s.status != "loaded" {
		imTextColored(imgui.Vec4{X: 0.85, Y: 0.30, Z: 0.30, W: 1.0}, "%s", s.status)
	}

	imgui.Separator()
	if imgui.Button("Load") {
		// Commit the choices so they persist, build the player under that policy, then start
		// pre-generation. An empty recording has nothing to generate.
		s.context.replayKeyframeBudgetMB = s.popupBudgetMB
		s.context.replayKeyframeMinInterval = s.popupMinInterval
		s.CreatePlayer()
		if s.player != nil {
			s.generating = s.info.FrameCount > 0
			if !s.generating {
				imgui.CloseCurrentPopup()
			}
		}
	}
	imgui.SameLine()
	if imgui.Button("Cancel") {
		imgui.CloseCurrentPopup()
	}

	imgui.EndPopup()
}

// Advance one recorded step and keep the world pointer current
func (s *ReplayViewer) AdvanceOne() {
	s.player.StepFrame()
	s.worldID = s.player.World()
}

// Frame the whole recorded motion. Older recordings lack stored bounds, so fall back to the
// live frame-0 bounds when the recorded extents are empty.
func (s *ReplayViewer) FrameRecording() {
	bounds := s.info.Bounds
	extents := bounds.Extents()
	if extents.X <= 0.0 && extents.Y <= 0.0 {
		bounds = s.worldID.Bounds()
	}
	FocusOnBounds(&s.context.camera, bounds)
	s.context.camera.zoom *= 1.5
}

// The recording loads after construction, so home is the fit made on load, not the view the
// constructor set up
func (s *ReplayViewer) FocusHome() {
	if s.player != nil {
		s.FrameRecording()
	} else {
		s.Sample.FocusHome()
	}
}

// , steps backward. Forward is the global single step on . so it works in every sample.
// Shift moves five frames, matching that key. Esc drops the selection.
func (s *ReplayViewer) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	if s.generating || s.player == nil || action != glfw.Press {
		return
	}

	if mods&(glfw.ModControl|glfw.ModAlt) != 0 {
		return
	}

	if key == glfw.KeyEscape {
		s.selKind = selNone
	} else if key == glfw.KeyComma {
		back := 1
		if mods&glfw.ModShift != 0 {
			back = 5
		}
		s.player.SeekFrame(max(0, s.player.Frame()-back))
		s.worldID = s.player.World()
		s.frameAccumulator = 0.0
		s.context.pause = true
	}
}

func (s *ReplayViewer) Step() {
	s.DrawLoadPopup()

	// While the ring builds the world is mid-fast-forward, so hold off drawing it. The popup
	// owns the screen and shows the progress bar until generation finishes.
	if s.generating {
		s.stepCount = s.player.Frame()
		s.context.singleStep = 0
		return
	}

	if s.player == nil {
		s.DrawScreenTextLine("%s", s.status)
		s.context.singleStep = 0
		return
	}

	if s.context.singleStep > 0 {
		// Stepping takes over from playback, like the transport buttons
		s.context.singleStep = max(0, s.context.singleStep-1)
		s.context.pause = true
		if !s.player.IsAtEnd() {
			s.AdvanceOne()
		}
		s.frameAccumulator = 0.0
	} else if !s.context.pause {
		// Speed scales how many recorded steps pass per display frame. Below 1 advances
		// only every few frames, above 1 advances several.
		s.frameAccumulator += s.speed
		for s.frameAccumulator >= 1.0 {
			s.frameAccumulator -= 1.0
			if s.player.IsAtEnd() {
				if s.loop {
					s.player.Restart()
					s.worldID = s.player.World()
				} else {
					s.frameAccumulator = 0.0
					break
				}
			}
			s.AdvanceOne()
		}
	}

	// Keep the base panel "step N" line tracking the replay frame
	s.stepCount = s.player.Frame()

	s.context.debugDraw.DrawingBounds = GetViewBounds(&s.context.camera)

	if s.worldID != nil {
		s.worldID.Draw(&s.context.debugDraw)
		if s.selKind == selQuery {
			s.player.DrawFrameQueries(&s.context.debugDraw, s.selQuery)
		}
		s.DrawSelectionHighlight()
	}

	if s.context.showUI {
		s.DrawOutlinePanel()
	}
}

// Shared transport row used by both the right panel and the timeline tab
func (s *ReplayViewer) DrawTransport() {
	if s.player == nil {
		return
	}

	frame := s.player.Frame()

	if imgui.Button("|<") {
		s.player.SeekFrame(0)
		s.worldID = s.player.World()
		s.frameAccumulator = 0.0
	}
	imgui.SameLine()
	if imgui.Button("<") {
		s.player.SeekFrame(frame - 1)
		s.worldID = s.player.World()
		s.frameAccumulator = 0.0
		s.context.pause = true
	}
	imgui.SameLine()

	if s.context.pause {
		imgui.PushStyleColorVec4(imgui.ColButton, hsvColor(2.0/7.0, 0.6, 0.6))
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, hsvColor(2.0/7.0, 0.7, 0.7))
		imgui.PushStyleColorVec4(imgui.ColButtonActive, hsvColor(2.0/7.0, 0.8, 0.8))
		if imgui.Button("Play ") {
			s.context.pause = false
		}
		imgui.PopStyleColorV(3)
	} else {
		imgui.PushStyleColorVec4(imgui.ColButton, hsvColor(1.0/7.0, 0.6, 0.6))
		imgui.PushStyleColorVec4(imgui.ColButtonHovered, hsvColor(1.0/7.0, 0.7, 0.7))
		imgui.PushStyleColorVec4(imgui.ColButtonActive, hsvColor(1.0/7.0, 0.8, 0.8))
		if imgui.Button("Pause") {
			s.context.pause = true
		}
		imgui.PopStyleColorV(3)
	}

	imgui.SameLine()
	if imgui.Button(">") {
		s.player.SeekFrame(frame + 1)
		s.worldID = s.player.World()
		s.frameAccumulator = 0.0
		s.context.pause = true
	}
	imgui.SameLine()
	if imgui.Button(">|") {
		s.player.SeekFrame(s.info.FrameCount)
		s.worldID = s.player.World()
		s.frameAccumulator = 0.0
	}
}

// A replay re-runs recorded inputs, so the live solver sliders would do nothing
func (s *ReplayViewer) HasSolverControls() bool {
	return false
}

func (s *ReplayViewer) HasProfile() bool {
	return false
}

// Wider than the default so the detail pane, hosted in the info panel, has room for ids and vectors.
func (s *ReplayViewer) InfoPanelWidthEm() float32 {
	return 22.0
}

// Right info panel: a compact summary with the selection detail below it. The scene tree lives in
// the Outline window and the transport in the Timeline tab.
func (s *ReplayViewer) DrawControls() bool {
	if imgui.Button("Show Timeline") {
		s.context.showMetrics = true
		s.selectTimelineTab = true
	}

	if s.player == nil {
		return false
	}

	if s.player.HasDiverged() {
		imTextColored(MakeColor(box2d.ColorRed), "****DIVERGED****")
	}

	end := ""
	if s.player.IsAtEnd() {
		end = "  (end)"
	}
	imTextDisabled("Frame %d / %d%s", s.player.Frame(), s.info.FrameCount, end)

	// The child takes the remaining panel height and scrolls a long detail. Return false so the
	// panel adds no trailing separator below this full height child.
	imgui.Separator()
	imTextColored(imgui.Vec4{X: 0.9, Y: 0.6, Z: 0.2, W: 1.0}, "Detail")
	imgui.BeginChildStr("detail")
	s.DrawDetail()
	imgui.EndChild()

	return false
}

func bodyShapes(body box2d.Body, capacity int) []box2d.Shape {
	shapes := make([]box2d.Shape, 0, capacity)
	for shape := range body.Shapes() {
		if len(shapes) == capacity {
			break
		}
		shapes = append(shapes, shape)
	}
	return shapes
}

func bodyJoints(body box2d.Body, capacity int) []box2d.Joint {
	joints := make([]box2d.Joint, 0, capacity)
	for joint := range body.Joints() {
		if len(joints) == capacity {
			break
		}
		joints = append(joints, joint)
	}
	return joints
}

func bodyContacts(body box2d.Body) []box2d.ContactData {
	capacity := min(body.ContactCapacity(), 64)
	contacts := make([]box2d.ContactData, 0, capacity)
	for contact := range body.ContactData() {
		if len(contacts) == capacity {
			break
		}
		contacts = append(contacts, contact)
	}
	return contacts
}

// Selection resolution. The selection is stored as creation ordinals so it survives a backward
// scrub that rebuilds the world. Each frame the ordinal is mapped back to a live id, or to null
// when that object does not exist at the current frame.
func (s *ReplayViewer) SelectedBody() box2d.Body {
	if s.selBodyOrdinal < 0 {
		return box2d.Body{}
	}
	return s.player.Body(s.selBodyOrdinal)
}

func (s *ReplayViewer) SelectedShape() box2d.Shape {
	body := s.SelectedBody()
	if s.selKind != selShape || !body.IsValid() {
		return box2d.Shape{}
	}
	shapes := bodyShapes(body, 32)
	if s.selSlot >= 0 && s.selSlot < len(shapes) {
		return shapes[s.selSlot]
	}
	return box2d.Shape{}
}

func (s *ReplayViewer) SelectedJoint() box2d.Joint {
	body := s.SelectedBody()
	if s.selKind != selJoint || !body.IsValid() {
		return box2d.Joint{}
	}
	joints := bodyJoints(body, 16)
	if s.selSlot >= 0 && s.selSlot < len(joints) {
		return joints[s.selSlot]
	}
	return box2d.Joint{}
}

func (s *ReplayViewer) FindBodyOrdinal(body box2d.Body) int {
	count := s.player.BodyCount()
	for i := range count {
		if s.player.Body(i) == body {
			return i
		}
	}
	return -1
}

// Map a picked shape back to its body ordinal and shape slot. A null shape clears the selection.
func (s *ReplayViewer) SelectShape(shape box2d.Shape) {
	if shape == (box2d.Shape{}) {
		s.selKind = selNone
		return
	}
	body := shape.Body()
	ordinal := s.FindBodyOrdinal(body)
	if ordinal < 0 {
		s.selKind = selNone
		return
	}
	shapes := bodyShapes(body, 32)
	slot := -1
	for i := range shapes {
		if shapes[i] == shape {
			slot = i
			break
		}
	}
	s.selKind = selShape
	s.selBodyOrdinal = ordinal
	s.selSlot = slot
	s.revealSelection = true // expand and scroll the tree to the picked shape next draw
}

// Draw a body's live contact points and normals, the most useful solver readout
func (s *ReplayViewer) DrawBodyContacts(body box2d.Body) {
	draw := s.context.draw
	for _, contact := range bodyContacts(body) {
		originA := contact.ShapeA.Body().Position()
		m := &contact.Manifold
		for j := range m.PointCount {
			point := originA.Add(m.Points[j].AnchorA)
			draw.DrawPoint(point, 6.0, box2d.ColorOrange)
			draw.DrawLine(point, point.Add(m.Normal.Scale(0.3)), box2d.ColorOrange)
		}
	}
}

// Highlight the current selection without touching the world. Queries are already drawn by
// Replay.DrawFrameQueries, so they need nothing here.
func (s *ReplayViewer) DrawSelectionHighlight() {
	draw := s.context.draw

	if s.selKind == selShape {
		shape := s.SelectedShape()
		if !shape.IsValid() {
			return
		}
		body := shape.Body()
		draw.DrawBounds(shape.AABB(), box2d.ColorYellow)
		draw.DrawTransform(body.Transform(), 0.5)
		draw.DrawPoint(body.WorldCenter(), 8.0, box2d.ColorYellow)
		s.DrawBodyContacts(body)
	} else if s.selKind == selBody {
		body := s.SelectedBody()
		if !body.IsValid() {
			return
		}
		draw.DrawBounds(body.ComputeAABB(), box2d.ColorYellow)
		draw.DrawTransform(body.Transform(), 0.5)
		draw.DrawPoint(body.WorldCenter(), 8.0, box2d.ColorYellow)
		s.DrawBodyContacts(body)
	} else if s.selKind == selJoint {
		joint := s.SelectedJoint()
		if !joint.IsValid() {
			return
		}
		a := joint.BodyA()
		b := joint.BodyB()
		if !a.IsValid() || !b.IsValid() {
			return
		}

		// Anchor points in world space, a line joins them for joints with separated frames
		pA := a.Transform().Apply(joint.LocalFrameA().P)
		pB := b.Transform().Apply(joint.LocalFrameB().P)
		draw.DrawLine(pA, pB, box2d.ColorMagenta)
		draw.DrawPoint(pA, 8.0, box2d.ColorMagenta)
		draw.DrawPoint(pB, 8.0, box2d.ColorMagenta)
	}
}

// Left Outline window holding the scene tree. The selection detail lives in the right info panel,
// so the tree owns the whole column. Opened from Step, which runs inside the imgui frame.
func (s *ReplayViewer) DrawOutlinePanel() {
	if s.player == nil {
		return
	}

	fontSize := imgui.FontSize()
	menuBarHeight := imgui.FrameHeight()
	drawerHeight := 16.0 * fontSize // matches the metrics drawer in sample.go
	top := menuBarHeight + 0.5*fontSize
	// Stop above the timeline drawer, which this sample keeps open
	bottom := s.context.camera.height - 0.5*fontSize
	if s.context.showMetrics {
		bottom = s.context.camera.height - drawerHeight - fontSize
	}

	imgui.SetNextWindowPos(imgui.Vec2{X: 0.5 * fontSize, Y: top})
	imgui.SetNextWindowSize(imgui.Vec2{X: 22.0 * fontSize, Y: bottom - top})
	imgui.BeginV(
		"Outline",
		nil,
		imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoTitleBar,
	)

	imTextColored(imgui.Vec4{X: 0.9, Y: 0.6, Z: 0.2, W: 1.0}, "Outline")
	imgui.BeginChildStr("tree")
	s.DrawOutlineTree()
	imgui.EndChild()

	imgui.End()
}

// The scene tree: bodies (creation order), each expandable to its shapes and joints, plus the
// current frame's queries. Clicking a row selects it; clicking a body arrow expands it.
func (s *ReplayViewer) DrawOutlineTree() {
	// A viewport pick asks the tree to reveal its target once: expand the owning body and scroll to
	// the row. Consumed at the end so it never fights the user's own expand/collapse.
	reveal := s.revealSelection

	count := s.player.BodyCount()
	for ord := range count {
		body := s.player.Body(ord)
		if body == (box2d.Body{}) || !body.IsValid() {
			continue
		}

		ownsSelection := s.selBodyOrdinal == ord &&
			(s.selKind == selBody || s.selKind == selShape || s.selKind == selJoint)

		name := body.Name()
		if name == "" {
			name = replayBodyTypeName(body.Type())
		}
		label := fmt.Sprintf("Body %d  %s###b%d", ord, name, ord)

		flags := imgui.TreeNodeFlagsOpenOnArrow | imgui.TreeNodeFlagsSpanAvailWidth
		if s.selKind == selBody && s.selBodyOrdinal == ord {
			flags |= imgui.TreeNodeFlagsSelected
		}
		// Reveal a picked shape or joint by expanding its body
		if reveal && ownsSelection && s.selKind != selBody {
			imgui.SetNextItemOpen(true)
		}
		open := imgui.TreeNodeExStrV(label, flags)
		if reveal && ownsSelection && s.selKind == selBody {
			imgui.SetScrollHereYV(0.5)
		}
		if imgui.IsItemClicked() && !imgui.IsItemToggledOpen() {
			s.selKind = selBody
			s.selBodyOrdinal = ord
			s.selSlot = -1
		}
		if !open {
			continue
		}

		shapes := bodyShapes(body, 32)
		for i, shape := range shapes {
			sl := fmt.Sprintf("Shape %d  %s###b%ds%d", i, replayShapeTypeName(shape.Type()), ord, i)
			lf := imgui.TreeNodeFlagsLeaf | imgui.TreeNodeFlagsSpanAvailWidth | imgui.TreeNodeFlagsNoTreePushOnOpen
			if s.selKind == selShape && s.selBodyOrdinal == ord && s.selSlot == i {
				lf |= imgui.TreeNodeFlagsSelected
			}
			imgui.TreeNodeExStrV(sl, lf)
			if reveal && s.selKind == selShape && s.selBodyOrdinal == ord && s.selSlot == i {
				imgui.SetScrollHereYV(0.5)
			}
			if imgui.IsItemClicked() {
				s.selKind = selShape
				s.selBodyOrdinal = ord
				s.selSlot = i
			}
		}

		joints := bodyJoints(body, 16)
		for j, joint := range joints {
			jl := fmt.Sprintf("%s joint###b%dj%d", replayJointTypeName(joint.Type()), ord, j)
			lf := imgui.TreeNodeFlagsLeaf | imgui.TreeNodeFlagsSpanAvailWidth | imgui.TreeNodeFlagsNoTreePushOnOpen
			if s.selKind == selJoint && s.selBodyOrdinal == ord && s.selSlot == j {
				lf |= imgui.TreeNodeFlagsSelected
			}
			imgui.TreeNodeExStrV(jl, lf)
			if imgui.IsItemClicked() {
				s.selKind = selJoint
				s.selBodyOrdinal = ord
				s.selSlot = j
			}
		}

		imgui.TreePop()
	}

	qn := s.player.FrameQueryCount()
	ql := fmt.Sprintf("Queries (%d)###queries", qn)
	if imgui.TreeNodeExStrV(ql, imgui.TreeNodeFlagsSpanAvailWidth) {
		for i := range qn {
			q := s.player.FrameQuery(i)
			qi := fmt.Sprintf("%s  (%d)###q%d", replayQueryTypeName(q.Type), q.HitCount, i)
			lf := imgui.TreeNodeFlagsLeaf | imgui.TreeNodeFlagsSpanAvailWidth | imgui.TreeNodeFlagsNoTreePushOnOpen
			if s.selKind == selQuery && s.selQuery == i {
				lf |= imgui.TreeNodeFlagsSelected
			}
			imgui.TreeNodeExStrV(qi, lf)
			if imgui.IsItemClicked() {
				s.selKind = selQuery
				s.selQuery = i
			}
		}
		imgui.TreePop()
	}

	s.revealSelection = false
}

// Detail pane for the current selection. Full width, so 64-bit hex fits without clipping.
func (s *ReplayViewer) DrawDetail() {
	if s.selKind == selNone {
		imTextWrapped("Click a node, or a shape in the view.")
		if s.worldID != nil {
			g := s.worldID.Gravity()
			c := s.worldID.Counters()
			imText("gravity (%.2f, %.2f)", g.X, g.Y)
			imText("bodies %d  shapes %d", c.BodyCount, c.ShapeCount)
			imText("contacts %d  joints %d", c.ContactCount, c.JointCount)
		}
		return
	}

	if s.selKind == selQuery {
		s.DrawQueryDetail()
		return
	}

	body := s.SelectedBody()
	if !body.IsValid() {
		imTextDisabled("Not present at this frame.")
		return
	}

	s.DrawBodyDetail(body)
	if s.selKind == selShape {
		shape := s.SelectedShape()
		if shape.IsValid() {
			s.DrawShapeDetail(shape)
		}
	} else if s.selKind == selJoint {
		joint := s.SelectedJoint()
		if joint.IsValid() {
			s.DrawJointDetail(joint)
		}
	}
	s.DrawContactDetail(body)
}

func (s *ReplayViewer) DrawBodyDetail(body box2d.Body) {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("Body", imgui.TreeNodeFlagsDefaultOpen) {
		return
	}

	name := body.Name()
	xf := body.Transform()
	v := body.LinearVelocity()

	if name == "" {
		name = "(none)"
	}

	imText("id      %d", bodyIndex(body))
	imText("name    %s", name)
	imText("type    %s", replayBodyTypeName(body.Type()))
	imText("pos     (%.3f, %.3f)", xf.P.X, xf.P.Y)
	imText("angle   %.1f deg", xf.Q.Angle()*57.2957795)
	imText("vel     (%.3f, %.3f)", v.X, v.Y)
	imText("omega   %.3f rad/s", body.AngularVelocity())
	imText("mass    %.4g kg", body.Mass())
	imText("inertia %.4g", body.RotationalInertia())
	imText("awake   %s", yesNo(body.Awake()))
	imText("enabled %s", yesNo(body.Enabled()))
	imText("bullet  %s", yesNo(body.IsBullet()))
	imText("gravity scale %.2f", body.GravityScale())
	imText("shapes %d  joints %d", body.ShapeCount(), body.JointCount())
}

func (s *ReplayViewer) DrawShapeDetail(shape box2d.Shape) {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("Shape", imgui.TreeNodeFlagsDefaultOpen) {
		return
	}

	imText("id      %d", shapeIndex(shape))
	imText("type     %s", replayShapeTypeName(shape.Type()))
	f := shape.Filter()
	imText("category 0x%016x", f.CategoryBits)
	imText("mask     0x%016x", f.MaskBits)
	imText("group    %d", f.GroupIndex)
	imText("density  %.3g", shape.Density())
	imText("friction %.3g", shape.Friction())
	imText("restitution %.3g", shape.Restitution())
	imText("sensor   %s", yesNo(shape.IsSensor()))
	mat := shape.SurfaceMaterial()
	imText("custom color 0x%06x", mat.CustomColor)
	aabb := shape.AABB()
	imText(
		"aabb (%.2f, %.2f)-(%.2f, %.2f)",
		aabb.LowerBound.X,
		aabb.LowerBound.Y,
		aabb.UpperBound.X,
		aabb.UpperBound.Y,
	)
}

func (s *ReplayViewer) DrawContactDetail(body box2d.Body) {
	contacts := bodyContacts(body)

	header := fmt.Sprintf("Contacts (%d)###contacts", len(contacts))
	if !imgui.CollapsingHeaderTreeNodeFlags(header) {
		return
	}

	for i := range contacts {
		m := &contacts[i].Manifold
		imText("shapes %d / %d", shapeIndex(contacts[i].ShapeA), shapeIndex(contacts[i].ShapeB))
		imText("normal (%.2f, %.2f)", m.Normal.X, m.Normal.Y)
		imText("points %d", m.PointCount)
		for j := range m.PointCount {
			mp := &m.Points[j]
			imText("  sep %.3f  Pn %.2g", mp.Separation, mp.NormalImpulse)
		}

		imgui.Separator()
	}
}

func (s *ReplayViewer) DrawJointDetail(joint box2d.Joint) {
	if !imgui.CollapsingHeaderTreeNodeFlagsV("Joint", imgui.TreeNodeFlagsDefaultOpen) {
		return
	}

	jointType := joint.Type()
	imText("type     %s", replayJointTypeName(jointType))
	imText("body A   %d", bodyIndex(joint.BodyA()))
	imText("body B   %d", bodyIndex(joint.BodyB()))
	imText("collide  %s", yesNo(joint.CollideConnected()))
	imText("force    %.3g", joint.ConstraintForce().Length())
	imText("torque   %.3g", joint.ConstraintTorque())

	switch jointType {
	case box2d.RevoluteJointType:
		revolute, _ := joint.Revolute()
		imText("angle    %.1f deg", revolute.Angle()*57.2957795)
	case box2d.PrismaticJointType:
		prismatic, _ := joint.Prismatic()
		imText("translation %.3f", prismatic.Translation())
	case box2d.DistanceJointType:
		distance, _ := joint.Distance()
		imText("length   %.3f", distance.CurrentLength())
	}
}

func (s *ReplayViewer) DrawQueryDetail() {
	count := s.player.FrameQueryCount()
	if s.selQuery < 0 || s.selQuery >= count {
		imTextDisabled("Query not present at this frame.")
		return
	}

	q := s.player.FrameQuery(s.selQuery)
	if !imgui.CollapsingHeaderTreeNodeFlagsV("Query", imgui.TreeNodeFlagsDefaultOpen) {
		return
	}

	imText("type     %s", replayQueryTypeName(q.Type))
	shapeLocal := q.Type == box2d.ReplayQueryShapeTestPoint ||
		q.Type == box2d.ReplayQueryShapeRayCast
	if !shapeLocal {
		imText("category 0x%016x", q.Filter.CategoryBits)
		imText("mask     0x%016x", q.Filter.MaskBits)
	} else {
		imText("shape    %d", shapeIndex(q.Shape))
	}
	imText("hits     %d", q.HitCount)

	// Hits as one wrapped id list, so a 50-hit query stays compact
	var line strings.Builder
	for h := 0; h < q.HitCount && line.Len() < 256-12; h++ {
		hit := s.player.FrameQueryHit(s.selQuery, h)
		fmt.Fprintf(&line, "%d ", shapeIndex(hit.Shape))
	}
	if q.HitCount > 0 {
		imTextWrapped("hit shapes: %s", line.String())
	}
}

// All replay controls live in the metrics drawer tab.
func (s *ReplayViewer) DrawMetricsTab() {
	tabFlags := imgui.TabItemFlagsNone
	if s.selectTimelineTab {
		tabFlags |= imgui.TabItemFlagsSetSelected
		s.selectTimelineTab = false
	}

	if !imgui.BeginTabItemV("Timeline", nil, tabFlags) {
		return
	}

	fontSize := imgui.FontSize()

	// File row, always available so a recording can be loaded even when none is open
	imText("File: %s", s.path)
	imgui.SameLine()
	imgui.TextUnformatted(s.status)

	if s.player == nil {
		imgui.EndTabItem()
		return
	}

	// Transport row: buttons, speed, loop, replay worker count
	s.DrawTransport()
	imgui.SameLine()

	speedNames := []string{"0.25x", "0.5x", "1x", "2x", "4x"}
	speedValues := []float32{0.25, 0.5, 1.0, 2.0, 4.0}
	speedIndex := int32(2)
	for i := range 5 {
		if s.speed == speedValues[i] {
			speedIndex = int32(i)
		}
	}
	imgui.PushItemWidth(5.0 * fontSize)
	if imgui.ComboStrarr("Speed", &speedIndex, speedNames, 5) {
		s.speed = speedValues[speedIndex]
	}
	imgui.PopItemWidth()
	imgui.SameLine()
	imgui.Checkbox("Loop", &s.loop)
	imgui.SameLine()

	// Replaying at a different worker count is a visual cross-thread determinism check.
	// 0 means use the recorded count. Re-open on release so the player is not rebuilt mid-drag.
	imgui.PushItemWidth(6.0 * fontSize)
	if imgui.SliderInt("Workers", &s.context.workerCount, 0, box2d.MaxWorkers) {
		if s.worldID != nil {
			s.worldID.SetWorkerCount(int(s.context.workerCount))
		}
	}
	imgui.PopItemWidth()

	// Live ring state: the effective spacing widens as the ring evicts under the budget, and the
	// memory held grows as keyframes accumulate. Budget and min interval are chosen in the Load
	// popup and persisted, so there is no live slider here.
	imTextDisabled("keyframe spacing %d frames, %.1f MB", s.player.KeyframeInterval(),
		float64(s.player.KeyframeBytes())/(1024.0*1024.0))

	// Scrubber: full width, seeks both directions
	scrub := int32(s.player.Frame())
	imgui.PushItemWidth(-1.0)
	if imgui.SliderInt("##frame", &scrub, 0, int32(s.info.FrameCount)) {
		s.player.SeekFrame(int(scrub))
		s.worldID = s.player.World()
		s.frameAccumulator = 0.0
		s.context.pause = true
	}
	imgui.PopItemWidth()

	// Mark where the replay first diverged on the scrubber track
	divergeFrame := s.player.DivergeFrame()
	if divergeFrame >= 0 && s.info.FrameCount > 0 {
		lo := imgui.ItemRectMin()
		hi := imgui.ItemRectMax()
		t := float32(divergeFrame) / float32(s.info.FrameCount)
		x := lo.X + t*(hi.X-lo.X)
		imgui.WindowDrawList().
			AddLineArgs(imgui.Vec2{X: x, Y: lo.Y}, imgui.Vec2{X: x, Y: hi.Y}, imCol32(220, 60, 60, 255), 2.0)
	}

	// Info row: recording metadata, live counts, divergence
	imText("frames %d", s.info.FrameCount)
	if s.info.TimeStep > 0.0 {
		imgui.SameLine()
		imText("   %.0f hz, %d sub-steps", 1.0/s.info.TimeStep, s.info.SubStepCount)
	}
	if s.info.LengthScale > 0.0 {
		imgui.SameLine()
		imText("   %.6g units/m", s.info.LengthScale)
	}
	if s.worldID != nil {
		c := s.worldID.Counters()
		imgui.SameLine()
		imText(
			"   bodies %d  shapes %d  contacts %d  joints %d",
			c.BodyCount,
			c.ShapeCount,
			c.ContactCount,
			c.JointCount,
		)
	}
	if divergeFrame >= 0 {
		imgui.SameLine()
		imTextColored(
			imgui.Vec4{X: 0.85, Y: 0.30, Z: 0.30, W: 1.0},
			"   diverged at frame %d",
			divergeFrame,
		)
	}

	imgui.EndTabItem()
}

// Left click selects a shape to inspect. Picking only reads the world, it never creates the drag
// joint the base sample does, so the replay is not mutated. Dragging stays disabled.
func (s *ReplayViewer) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button != glfw.MouseButton1 || s.worldID == nil {
		return
	}

	// A tiny box around the click point, exact at any distance with the click as the origin
	d := box2d.Vec2{X: 0.001, Y: 0.001}
	box := box2d.AABB{LowerBound: d.Neg(), UpperBound: d}
	var picked box2d.Shape
	s.worldID.OverlapAABB(p, box, box2d.DefaultQueryFilter(), func(shape box2d.Shape) bool {
		if shape.TestPoint(p) {
			picked = shape
			return false
		}
		return true
	})

	// A miss clears the selection
	s.SelectShape(picked)
}

func (s *ReplayViewer) MouseUp(p box2d.Pos, button glfw.MouseButton) {
}

func (s *ReplayViewer) MouseMove(p box2d.Pos) {
}

func init() {
	RegisterReplay("Replay", "Viewer", CreateReplayViewer)
}
