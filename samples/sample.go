// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/ncruces/zenity"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const infoPanelWidth = 16.0

const fltMin = 1.17549435e-38

const settingsFileName = "settings.ini"

type SampleContext struct {
	window    *glfw.Window
	camera    Camera
	draw      *Draw
	sample    Sampler
	capacity  box2d.Capacity
	debugDraw box2d.DebugDraw

	uiScale                      float32
	hertz                        float32
	recycleDistance              float32
	subStepCount                 int32
	restitutionIterations        int32
	workerCount                  int32
	restart                      bool
	pause                        bool
	singleStep                   int
	enableWarmStarting           bool
	enableContinuous             bool
	enableRestitutionPropagation bool
	enableSleep                  bool
	showUI                       bool

	showMetrics bool

	showProfile bool

	showControls bool

	homeCenter box2d.Pos
	homeZoom   float32

	openSamplePicker bool

	recordingFile string

	replayFile string

	savedRecordingFile string

	sampleIndex int
	newUser     bool

	replayKeyframeBudgetMB    int32
	replayKeyframeMinInterval int32
}

func newSampleContext() *SampleContext {
	return &SampleContext{
		uiScale:                   1.0,
		hertz:                     60.0,
		recycleDistance:           0.05,
		subStepCount:              4,
		restitutionIterations:     2,
		workerCount:               1,
		enableWarmStarting:        true,
		enableContinuous:          true,
		enableSleep:               true,
		showUI:                    true,
		homeCenter:                box2d.Pos{X: 0.0, Y: 20.0},
		homeZoom:                  1.0,
		recordingFile:             "recording.b2rec",
		newUser:                   true,
		replayKeyframeBudgetMB:    512,
		replayKeyframeMinInterval: 16,
	}
}

func (context *SampleContext) Save() {
	file, err := os.Create(settingsFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", settingsFileName, err)
		return
	}
	fmt.Fprintf(file, "{\n")
	fmt.Fprintf(file, "  \"sampleIndex\": %d,\n", context.sampleIndex)
	fmt.Fprintf(file, "  \"newUser\": %d,\n", 0)
	fmt.Fprintf(file, "  \"drawShapes\": %t,\n", context.debugDraw.DrawShapes)
	fmt.Fprintf(file, "  \"showMetrics\": %t,\n", context.showMetrics)
	fmt.Fprintf(file, "  \"showProfile\": %t,\n", context.showProfile)
	fmt.Fprintf(file, "  \"replayKeyframeBudgetMB\": %d,\n", context.replayKeyframeBudgetMB)
	fmt.Fprintf(file, "  \"replayKeyframeMinInterval\": %d\n", context.replayKeyframeMinInterval)
	fmt.Fprintf(file, "}\n")
	if err := file.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", settingsFileName, err)
	}
}

type debugDrawer struct {
	context *SampleContext
}

func (d debugDrawer) DrawPolygon(transform box2d.WorldTransform, vertices []box2d.Vec2, color box2d.HexColor) {
	d.context.draw.DrawPolygon(transform, vertices, color)
}

func (d debugDrawer) DrawSolidPolygon(transform box2d.WorldTransform, vertices []box2d.Vec2, radius float32, color box2d.HexColor) {
	d.context.draw.DrawSolidPolygon(transform, vertices, radius, color)
}

func (d debugDrawer) DrawCircle(center box2d.Pos, radius float32, color box2d.HexColor) {
	d.context.draw.DrawCircle(center, radius, color)
}

func (d debugDrawer) DrawSolidCircle(transform box2d.WorldTransform, center box2d.Vec2, radius float32, color box2d.HexColor) {
	d.context.draw.DrawSolidCircle(transform, center, radius, color)
}

func (d debugDrawer) DrawSolidCapsule(p1, p2 box2d.Pos, radius float32, color box2d.HexColor) {
	d.context.draw.DrawCapsule(p1, p2, radius, color)
}

func (d debugDrawer) DrawLine(p1, p2 box2d.Pos, color box2d.HexColor) {
	d.context.draw.DrawLine(p1, p2, color)
}

func (d debugDrawer) DrawTransform(transform box2d.WorldTransform) {
	d.context.draw.DrawTransform(transform, 1.0)
}

func (d debugDrawer) DrawPoint(p box2d.Pos, size float32, color box2d.HexColor) {
	d.context.draw.DrawPoint(p, size, color)
}

func (d debugDrawer) DrawString(p box2d.Pos, s string, color box2d.HexColor) {
	d.context.draw.DrawString(&d.context.camera, p, color, "%s", s)
}

func (d debugDrawer) DrawAABB(aabb box2d.AABB, color box2d.HexColor) {
	d.context.draw.DrawBounds(aabb, color)
}

func (context *SampleContext) Load() {
	context.camera = GetDefaultCamera()
	context.debugDraw = box2d.DefaultDebugDraw()
	context.debugDraw.Drawer = debugDrawer{context}

	context.recycleDistance = box2d.ContactRecycleDistance()

	data, err := os.ReadFile(settingsFileName)
	if err != nil || len(data) == 0 {
		return
	}

	context.newUser = false

	var values map[string]json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse %s: %v\n", settingsFileName, err)
		return
	}

	readInt := func(key string, target *int) {
		if raw, ok := values[key]; ok {
			if v, err := strconv.Atoi(string(bytes.TrimSpace(raw))); err == nil {
				*target = v
			}
		}
	}
	readInt32 := func(key string, target *int32) {
		v := int(*target)
		readInt(key, &v)
		*target = int32(v)
	}
	readBool := func(key string, target *bool) {
		if raw, ok := values[key]; ok {
			switch string(bytes.TrimSpace(raw)) {
			case "true":
				*target = true
			case "false":
				*target = false
			}
		}
	}

	readInt("sampleIndex", &context.sampleIndex)
	readBool("drawShapes", &context.debugDraw.DrawShapes)
	readBool("showMetrics", &context.showMetrics)
	readBool("showProfile", &context.showProfile)
	readInt32("replayKeyframeBudgetMB", &context.replayKeyframeBudgetMB)
	readInt32("replayKeyframeMinInterval", &context.replayKeyframeMinInterval)
}

type Sampler interface {
	base() *Sample
	Step()
	DrawControls() bool
	HasSolverControls() bool
	HasProfile() bool
	InfoPanelWidthEm() float32
	DrawMetricsTab()
	Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey)
	FocusHome()
	MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey)
	MouseUp(p box2d.Pos, button glfw.MouseButton)
	MouseMove(p box2d.Pos)
	Destroy()
}

const (
	maxTasks        = 512
	maxThreads      = 64
	profileCapacity = 512
)

type Sample struct {
	self Sampler

	context *SampleContext
	camera  *Camera
	draw    *Draw

	mouseBodyID box2d.Body

	worldID         *box2d.World
	recording       *box2d.Recording
	recordStartStep int
	mouseJointID    box2d.Joint
	mousePoint      box2d.Pos
	mouseForceScale float32
	stepCount       int
	screenTextX     float32
	screenTextY     float32

	profiles            [profileCapacity]box2d.Profile
	currentProfileIndex int
	profileReadIndex    uint64
	profileWriteIndex   uint64

	didStep bool
}

func (s *Sample) init(context *SampleContext, self Sampler, createWorld bool) {
	s.self = self
	s.context = context
	s.camera = &context.camera
	s.draw = context.draw

	s.screenTextX = 5.0
	s.screenTextY = 0.0

	s.mouseForceScale = 100.0

	scenes.SetRandomSeed(scenes.RandSeed)

	if createWorld {
		s.CreateWorld()
	}
}

func (s *Sample) base() *Sample { return s }

func (s *Sample) Destroy() {
	if s.worldID != nil {
		s.FinishRecording()
		s.worldID.Close()
		s.worldID = nil
	}
}

func (s *Sample) StartRecording() {
	if s.recording != nil {
		return
	}

	start := time.Now()

	recording := box2d.NewRecording(0)
	s.worldID.StartRecording(recording)

	if recording.Size() == 0 {
		return
	}

	s.recording = recording
	s.recordStartStep = s.stepCount

	ms := float32(time.Since(start).Seconds() * 1000.0)
	fmt.Printf("b2World_StartRecording took : %.6g ms", ms)
}

func (s *Sample) FinishRecording() {
	if s.recording == nil {
		return
	}

	s.worldID.StopRecording()
	if err := saveRecording(s.recording, s.context.recordingFile); err == nil {
		s.context.savedRecordingFile = s.context.recordingFile
	} else {
		fmt.Fprintf(os.Stderr, "failed to save recording: %v\n", err)
	}
	s.recording = nil
}

func saveRecording(recording *box2d.Recording, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := recording.WriteTo(file); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func (s *Sample) CreateWorld() {
	if s.worldID != nil {
		s.FinishRecording()
		s.worldID.Close()
		s.worldID = nil
	}

	worldDef := box2d.DefaultWorldDef()
	worldDef.WorkerCount = int(s.context.workerCount)
	worldDef.UserData = s.self
	worldDef.EnableSleep = s.context.enableSleep
	worldDef.Capacity = s.context.capacity
	s.worldID = box2d.NewWorld(&worldDef)

	s.worldID.SetContactRecycleDistance(s.context.recycleDistance)
}

func (s *Sample) ResetText() {
	fontSize := imgui.FontSize()
	if s.context.showUI {
		s.screenTextY = imgui.FrameHeight() + 1.5*fontSize
	} else {
		s.screenTextY = 3.0 * fontSize
	}

	s.screenTextX = 5.0
	if s.IsProfileVisible() {
		s.screenTextX += s.GetProfilePanelWidth() + 0.5*fontSize
	}
}

func (s *Sample) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if s.mouseJointID.IsValid() {
		return
	}

	if button == glfw.MouseButton1 {
		d := box2d.Vec2{X: 0.001, Y: 0.001}
		box := box2d.AABB{LowerBound: d.Neg(), UpperBound: d}

		s.mousePoint = p

		var queryBody box2d.Body
		s.worldID.OverlapAABB(p, box, box2d.DefaultQueryFilter(), func(shape box2d.Shape) bool {
			body := shape.Body()
			if body.Type() != box2d.DynamicBody {
				return true
			}

			if shape.TestPoint(p) {
				queryBody = body
				return false
			}

			return true
		})

		if queryBody.IsValid() {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.KinematicBody
			bodyDef.Position = s.mousePoint
			bodyDef.EnableSleep = false
			s.mouseBodyID = s.worldID.CreateBody(&bodyDef)

			jointDef := box2d.DefaultMotorJointDef()
			jointDef.Base.BodyA = s.mouseBodyID
			jointDef.Base.BodyB = queryBody
			jointDef.Base.LocalFrameB.P = queryBody.LocalPoint(p)
			jointDef.LinearHertz = 7.5
			jointDef.LinearDampingRatio = 1.0

			massData := queryBody.MassData()
			g := s.worldID.Gravity().Length()
			mg := massData.Mass * g

			jointDef.MaxSpringForce = s.mouseForceScale * mg

			if massData.Mass > 0.0 {
				lever := float32(math.Sqrt(float64(massData.RotationalInertia / massData.Mass)))
				jointDef.MaxVelocityTorque = 0.25 * lever * mg
			}

			s.mouseJointID = s.worldID.CreateMotorJoint(&jointDef).Joint
		}
	}
}

func (s *Sample) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if s.mouseJointID != (box2d.Joint{}) && button == glfw.MouseButton1 {
		s.mouseJointID.Destroy()
		s.mouseJointID = box2d.Joint{}

		s.mouseBodyID.Destroy()
		s.mouseBodyID = box2d.Body{}
	}
}

func (s *Sample) MouseMove(p box2d.Pos) {
	if !s.mouseJointID.IsValid() {
		s.mouseJointID = box2d.Joint{}
	}

	s.mousePoint = p
}

func (s *Sample) DrawScreenTextLine(format string, args ...any) {
	s.draw.DrawScreenString(s.screenTextX, s.screenTextY, box2d.ColorWhite, format, args...)
	s.screenTextY += 1.5 * imgui.FontSize()
}

func (s *Sample) InfoPanelWidthEm() float32 {
	return infoPanelWidth
}

func (s *Sample) FocusHome() {
	s.context.camera.center = s.context.homeCenter
	s.context.camera.zoom = s.context.homeZoom
}

func (s *Sample) Step() {
	s.didStep = false

	var timeStep float32
	if !s.context.pause || s.context.singleStep > 0 {
		if s.context.hertz > 0.0 {
			timeStep = 1.0 / s.context.hertz
		}
		s.context.singleStep = max(0, s.context.singleStep-1)
	}

	if s.mouseJointID != (box2d.Joint{}) && !s.mouseJointID.IsValid() {
		s.mouseJointID = box2d.Joint{}

		if s.mouseBodyID != (box2d.Body{}) {
			s.mouseBodyID.Destroy()
			s.mouseBodyID = box2d.Body{}
		}
	}

	if s.mouseBodyID != (box2d.Body{}) && timeStep > 0.0 {
		wake := true
		s.mouseBodyID.SetTargetTransform(box2d.WorldTransform{P: s.mousePoint, Q: box2d.RotIdentity}, timeStep, wake)
	}

	s.context.debugDraw.DrawingBounds = GetViewBounds(&s.context.camera)

	s.worldID.SetSleepingEnabled(s.context.enableSleep)
	s.worldID.SetWarmStartingEnabled(s.context.enableWarmStarting)
	s.worldID.SetContinuousEnabled(s.context.enableContinuous)
	s.worldID.SetRestitutionIterations(int(s.context.restitutionIterations))
	s.worldID.SetRestitutionPropagationEnabled(s.context.enableRestitutionPropagation)

	s.worldID.Step(timeStep, int(s.context.subStepCount))

	s.worldID.Draw(&s.context.debugDraw)

	if timeStep > 0.0 {
		s.stepCount += 1
		s.didStep = true

		if s.profileWriteIndex == profileCapacity+s.profileReadIndex {
			s.profileReadIndex += 1
		}

		s.currentProfileIndex = int(s.profileWriteIndex & (profileCapacity - 1))
		s.profiles[s.currentProfileIndex] = s.worldID.Profile()

		s.profileWriteIndex += 1
	}
}

func (s *Sample) DrawControls() bool { return false }

func (s *Sample) HasSolverControls() bool { return true }

func (s *Sample) HasProfile() bool { return true }

func (s *Sample) DrawMetricsTab() {}

func (s *Sample) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {}

func (s *Sample) ResetProfile() {
	s.recordStartStep -= s.stepCount
	s.stepCount = 0
	s.profiles = [profileCapacity]box2d.Profile{}
	s.currentProfileIndex = 0
	s.profileReadIndex = 0
	s.profileWriteIndex = 0
}

type rowDef struct {
	name   string
	indent int
	color  uint32
}

func addSegment(dl *imgui.DrawList, availWidth, t, stepNow float32, col uint32, x float32, cursor imgui.Vec2, barHeight float32) float32 {
	w := availWidth * (t / stepNow)
	if w > 0.0 {
		dl.AddRectFilled(imgui.Vec2{X: x, Y: cursor.Y}, imgui.Vec2{X: x + w, Y: cursor.Y + barHeight}, col)
		x += w
	}

	return x
}

const (
	profileSectionWidth = 8.0
	profileIndentWidth  = 0.75
	profileValueWidth   = 3.0
	profileBarWidth     = 4.0
	profileColumnCount  = 5
)

var (
	colorStep            = imCol32(230, 230, 230, 255)
	colorPairs           = imCol32(102, 153, 255, 255)
	colorCollide         = imCol32(255, 140, 51, 255)
	colorSolve           = imCol32(102, 204, 102, 255)
	colorSolveChild      = imCol32(150, 190, 150, 255)
	colorConstraintChild = imCol32(110, 200, 190, 255)
	colorSensors         = imCol32(200, 120, 220, 255)
	colorOther           = imCol32(140, 140, 140, 255)

	colorTrack = imCol32(70, 70, 74, 255)
)

func (s *Sample) IsProfileVisible() bool {
	return s.context.showUI && s.context.showProfile && s.self.HasProfile()
}

func (s *Sample) GetProfilePanelWidth() float32 {
	if !s.IsProfileVisible() {
		return 0.0
	}

	style := imgui.CurrentStyle()
	fontSize := imgui.FontSize()
	width := float32(profileSectionWidth + 3.0*profileValueWidth + profileBarWidth)
	return width*fontSize + 2.0*profileColumnCount*style.CellPadding().X + 2.0*style.WindowPadding().X
}

func (s *Sample) DrawProfile() {
	if !s.IsProfileVisible() {
		return
	}

	fontSize := imgui.FontSize()
	menuBarHeight := imgui.FrameHeight()
	panelWidth := s.GetProfilePanelWidth()

	imgui.SetNextWindowPos(imgui.Vec2{X: 0.5 * fontSize, Y: menuBarHeight + 0.5*fontSize})
	imgui.SetNextWindowSize(imgui.Vec2{X: panelWidth, Y: s.camera.height - menuBarHeight - fontSize})

	imgui.BeginV("Profile", nil,
		imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoTitleBar)

	count := int(s.profileWriteIndex - s.profileReadIndex)

	const rowCount = 22
	var histories [rowCount][profileCapacity]float32
	var totals [rowCount]float32
	for i := range count {
		idx := int((s.profileReadIndex + uint64(i)) & (profileCapacity - 1))
		p := &s.profiles[idx]
		histories[0][i] = p.Step
		histories[1][i] = p.Pairs
		histories[2][i] = p.Collide
		histories[3][i] = p.Solve
		histories[4][i] = p.SolverSetup
		histories[5][i] = p.Constraints
		histories[6][i] = p.PrepareConstraints
		histories[7][i] = p.IntegrateVelocities
		histories[8][i] = p.WarmStart
		histories[9][i] = p.SolveImpulses
		histories[10][i] = p.IntegratePositions
		histories[11][i] = p.RelaxImpulses
		histories[12][i] = p.Restitution
		histories[13][i] = p.StoreImpulses
		histories[14][i] = p.SplitIslands
		histories[15][i] = p.Transforms
		histories[16][i] = p.JointEvents
		histories[17][i] = p.HitEvents
		histories[18][i] = p.Refit
		histories[19][i] = p.SleepIslands
		histories[20][i] = p.Bullets
		histories[21][i] = p.Sensors
		for j := range rowCount {
			totals[j] += histories[j][i]
		}
	}

	const nowWindow = 10
	var now [rowCount]float32
	{
		n := min(count, nowWindow)
		if n > 0 {
			inv := 1.0 / float32(n)
			for r := range rowCount {
				var sum float32
				for i := count - n; i < count; i++ {
					sum += histories[r][i]
				}
				now[r] = sum * inv
			}
		}
	}

	var avg [rowCount]float32
	if count > 0 {
		scale := 1.0 / float32(count)
		for i := range rowCount {
			avg[i] = scale * totals[i]
		}
	}

	var rowMax [rowCount]float32
	for r := range rowCount {
		for i := range count {
			if histories[r][i] > rowMax[r] {
				rowMax[r] = histories[r][i]
			}
		}
	}

	rows := [rowCount]rowDef{
		{"step", 0, colorStep},
		{"pairs", 0, colorPairs},
		{"collide", 0, colorCollide},
		{"solve", 0, colorSolve},
		{"setup", 1, colorSolveChild},
		{"constraints", 1, colorSolveChild},
		{"prepare", 2, colorConstraintChild},
		{"velocities", 2, colorConstraintChild},
		{"warm start", 2, colorConstraintChild},
		{"bias", 2, colorConstraintChild},
		{"positions", 2, colorConstraintChild},
		{"relax", 2, colorConstraintChild},
		{"restitution", 2, colorConstraintChild},
		{"store", 2, colorConstraintChild},
		{"split", 2, colorConstraintChild},
		{"transforms", 1, colorSolveChild},
		{"joint events", 1, colorSolveChild},
		{"hit events", 1, colorSolveChild},
		{"refit BVH", 1, colorSolveChild},
		{"sleep", 1, colorSolveChild},
		{"bullets", 1, colorSolveChild},
		{"sensors", 0, colorSensors},
	}

	stepNow := max(now[0], 0.001)

	if imgui.Button("Reset") {
		s.ResetProfile()
	}
	imgui.SameLine()
	imText(" step %.2f ms", now[0])

	{
		pairsT := now[1]
		collideT := now[2]
		solveT := now[3]
		sensorsT := now[21]
		otherT := max(stepNow-pairsT-collideT-solveT-sensorsT, 0.0)

		availWidth := imgui.ContentRegionAvail().X
		barHeight := 1.5 * fontSize
		dl := imgui.WindowDrawList()
		cursor := imgui.CursorScreenPos()
		x := cursor.X

		x = addSegment(dl, availWidth, pairsT, stepNow, colorPairs, x, cursor, barHeight)
		x = addSegment(dl, availWidth, collideT, stepNow, colorCollide, x, cursor, barHeight)
		x = addSegment(dl, availWidth, solveT, stepNow, colorSolve, x, cursor, barHeight)
		x = addSegment(dl, availWidth, sensorsT, stepNow, colorSensors, x, cursor, barHeight)
		addSegment(dl, availWidth, otherT, stepNow, colorOther, x, cursor, barHeight)

		imgui.Dummy(imgui.Vec2{X: availWidth, Y: barHeight})
	}

	tableFlags := imgui.TableFlagsBordersInnerV | imgui.TableFlagsRowBg | imgui.TableFlagsSizingFixedFit

	if imgui.BeginTableV("profile", profileColumnCount, tableFlags, imgui.Vec2{}, 0.0) {
		imgui.TableSetupColumnV("section", imgui.TableColumnFlagsWidthFixed, profileSectionWidth*fontSize, 0)
		imgui.TableSetupColumnV("now", imgui.TableColumnFlagsWidthFixed, profileValueWidth*fontSize, 0)
		imgui.TableSetupColumnV("avg", imgui.TableColumnFlagsWidthFixed, profileValueWidth*fontSize, 0)
		imgui.TableSetupColumnV("max", imgui.TableColumnFlagsWidthFixed, profileValueWidth*fontSize, 0)
		imgui.TableSetupColumnV("%", imgui.TableColumnFlagsWidthFixed, profileBarWidth*fontSize, 0)
		imgui.TableHeadersRow()

		for r := range rowCount {
			d := &rows[r]

			imgui.TableNextRow()

			imgui.TableNextColumn()
			indent := float32(d.indent) * profileIndentWidth * fontSize
			imgui.SetCursorPosX(imgui.CursorPosX() + indent)
			imgui.PushStyleColorU32(imgui.ColText, d.color)
			imgui.TextUnformatted(d.name)
			imgui.PopStyleColor()

			idle := avg[r] < 0.005
			if idle {
				imgui.PushStyleColorVec4(imgui.ColText, *imgui.StyleColorVec4(imgui.ColTextDisabled))
			}
			imgui.TableNextColumn()
			imText("%6.2f", now[r])
			imgui.TableNextColumn()
			imText("%6.2f", avg[r])
			imgui.TableNextColumn()
			imText("%6.2f", rowMax[r])
			if idle {
				imgui.PopStyleColor()
			}

			imgui.TableNextColumn()

			isBudget := r == 0 && s.context.hertz > 0.0
			reference := stepNow
			if isBudget {
				reference = 1000.0 / s.context.hertz
			}
			frac := clampFloat(now[r]/reference, 0.0, 1.0)
			barWidth := imgui.ContentRegionAvail().X
			barHeight := imgui.TextLineHeight()
			p := imgui.CursorScreenPos()
			drawList := imgui.WindowDrawList()
			drawList.AddRectFilled(p, imgui.Vec2{X: p.X + barWidth, Y: p.Y + barHeight}, colorTrack)
			drawList.AddRectFilled(p, imgui.Vec2{X: p.X + frac*barWidth, Y: p.Y + barHeight}, d.color)
			imgui.Dummy(imgui.Vec2{X: barWidth, Y: barHeight})
			if isBudget {
				imSetItemTooltip("%.0f%% of %.1f ms frame", 100.0*now[r]/reference, reference)
			} else {
				imSetItemTooltip("%.0f%% of step", 100.0*now[r]/reference)
			}
		}
		imgui.EndTable()
	}

	imgui.End()
}

func (s *Sample) DrawMetrics() {
	if !s.context.showMetrics {
		return
	}

	fontSize := imgui.FontSize()
	menuWidth := s.self.InfoPanelWidthEm() * fontSize
	drawerHeight := 16.0 * fontSize
	drawerX := 0.5 * fontSize
	if s.IsProfileVisible() {
		drawerX += s.GetProfilePanelWidth() + 0.5*fontSize
	}
	drawerWidth := s.camera.width - menuWidth - fontSize - drawerX

	imgui.SetNextWindowPos(imgui.Vec2{X: drawerX, Y: s.camera.height - drawerHeight - 0.5*fontSize})
	imgui.SetNextWindowSize(imgui.Vec2{X: drawerWidth, Y: drawerHeight})

	imgui.BeginV("Metrics", nil,
		imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoTitleBar)

	if imgui.BeginTabBarV("MetricsTabs", imgui.TabBarFlagsNone) {
		if s.self.HasProfile() && imgui.BeginTabItem("Frame Time") {
			var maxValue float32
			var times [profileCapacity]float32
			var stepTimes [profileCapacity]float32
			var pairsTimes [profileCapacity]float32
			var collideTimes [profileCapacity]float32
			var solveTimes [profileCapacity]float32
			var sensorsTimes [profileCapacity]float32
			count := int(s.profileWriteIndex - s.profileReadIndex)
			for i := range count {
				index := (s.profileReadIndex + uint64(i)) & (profileCapacity - 1)
				times[i] = float32(i) / 60.0
				stepTimes[i] = s.profiles[index].Step
				pairsTimes[i] = s.profiles[index].Pairs
				collideTimes[i] = s.profiles[index].Collide
				solveTimes[i] = s.profiles[index].Solve
				sensorsTimes[i] = s.profiles[index].Sensors
				maxValue = max(stepTimes[i], maxValue)
			}

			implot.PushStyleVarFloat(implot.StyleVarMinorAlpha, 1.0)
			implot.PushStyleColorU32(implot.ColAxisGrid, imCol32(255, 255, 255, 15))

			plotSize := imgui.ContentRegionAvail()
			if implot.BeginPlotV("Profile", plotSize, implot.FlagsNoTitle) {
				implot.SetupAxesV("", "ms", 0, 0)
				implot.SetupAxesLimitsV(0, profileCapacity/60.0, 0.0, float64(maxValue), implot.CondAlways)
				plotProfileLine("step", colorStep, times[:count], stepTimes[:count])
				plotProfileLine("pairs", colorPairs, times[:count], pairsTimes[:count])
				plotProfileLine("collide", colorCollide, times[:count], collideTimes[:count])
				plotProfileLine("solve", colorSolve, times[:count], solveTimes[:count])
				plotProfileLine("sensors", colorSensors, times[:count], sensorsTimes[:count])
				implot.EndPlot()
			}

			implot.PopStyleColor()
			implot.PopStyleVar()

			imgui.EndTabItem()
		}

		if imgui.BeginTabItem("Counters") {
			if !s.worldID.IsValid() {
				imgui.TextDisabled("No world")
			} else {
				c := s.worldID.Counters()
				capacity := s.worldID.MaxCapacity()
				colorCount := len(c.ColorCounts)
				overflowIndex := colorCount - 1

				if imgui.BeginTableV("counters_layout", 3, imgui.TableFlagsSizingFixedFit, imgui.Vec2{}, 0.0) {
					imgui.TableSetupColumnV("world", imgui.TableColumnFlagsWidthFixed, 14.0*fontSize, 0)
					imgui.TableSetupColumnV("memory", imgui.TableColumnFlagsWidthFixed, 11.0*fontSize, 0)
					imgui.TableSetupColumnV("solver", imgui.TableColumnFlagsWidthStretch, 0.0, 0)
					imgui.TableNextRow()
					imgui.TableNextColumn()

					imText("bodies   %d / %d", c.BodyCount, capacity.StaticBodyCount+capacity.DynamicBodyCount)
					imText("shapes   %d / %d", c.ShapeCount, capacity.StaticShapeCount+capacity.DynamicShapeCount)
					imText("contacts %d / %d", c.ContactCount, capacity.ContactCount)
					imText("joints   %d", c.JointCount)
					imText("islands  %d", c.IslandCount)
					imText("tasks    %d", c.TaskCount)

					imgui.TableNextColumn()

					imText("static tree  %d", c.StaticTreeHeight)
					imText("movable tree %d", c.TreeHeight)
					imText("alloc %d K", c.ByteCount/1024)
					imText("stack %d K", c.StackUsed/1024)

					imgui.TableNextColumn()

					totalCount := 0
					normalCount := 0
					for i := range colorCount {
						totalCount += c.ColorCounts[i]
						if i != overflowIndex {
							normalCount += c.ColorCounts[i]
						}
					}
					overflowCount := c.ColorCounts[overflowIndex]

					imText("%d constraints across %d colors", totalCount, colorCount-1)

					availWidth := imgui.ContentRegionAvail().X
					barHeight := 2.0 * fontSize
					dl := imgui.WindowDrawList()

					cursor := imgui.CursorScreenPos()
					dl.AddRectFilled(cursor, imgui.Vec2{X: cursor.X + availWidth, Y: cursor.Y + barHeight}, imCol32(40, 40, 40, 255))
					if normalCount > 0 {
						x := cursor.X
						invTotal := 1.0 / float32(normalCount)
						for i := range overflowIndex {
							cnt := c.ColorCounts[i]
							if cnt == 0 {
								continue
							}
							segW := availWidth * float32(cnt) * invTotal
							hex := box2d.GraphColor(i)
							col := imCol32(uint8(hex>>16), uint8(hex>>8), uint8(hex), 255)
							dl.AddRectFilled(imgui.Vec2{X: x, Y: cursor.Y}, imgui.Vec2{X: x + segW, Y: cursor.Y + barHeight}, col)
							x += segW
						}
					}
					imgui.Dummy(imgui.Vec2{X: availWidth, Y: barHeight})

					imgui.Spacing()
					var overflowFrac float32
					if totalCount > 0 {
						overflowFrac = float32(overflowCount) / float32(totalCount)
					}
					imgui.PushStyleColorU32(imgui.ColPlotHistogram, imCol32(220, 60, 60, 255))
					imgui.ProgressBarV(overflowFrac, imgui.Vec2{X: -fltMin, Y: 0.0}, fmt.Sprintf("overflow %d", overflowCount))
					imgui.PopStyleColor()

					var recycledFrac float32
					if c.AwakeContactCount > 0 {
						recycledFrac = clampFloat(float32(c.RecycledContactCount)/float32(c.AwakeContactCount), 0.0, 1.0)
					}
					recycledOverlay := fmt.Sprintf("recycled %d / %d", c.RecycledContactCount, c.AwakeContactCount)
					imgui.ProgressBarV(recycledFrac, imgui.Vec2{X: -fltMin, Y: 0.0}, recycledOverlay)

					imgui.EndTable()
				}
			}

			imgui.EndTabItem()
		}

		s.self.DrawMetricsTab()

		imgui.EndTabBar()
	}

	imgui.End()
}

func plotProfileLine(label string, color uint32, xs, ys []float32) {
	spec := implot.NewSpec()
	defer spec.Destroy()
	spec.SetLineColor(imgui.ColorConvertU32ToFloat4(color))
	var zero float32
	xp, yp := &zero, &zero
	if len(xs) > 0 {
		xp, yp = &xs[0], &ys[0]
	}
	implot.PlotLineFloatPtrFloatPtrV(label, xp, yp, int32(len(xs)), *spec)
}

func (s *Sample) DrawHud(frameTime float32) {
	entry := &sampleEntries[s.context.sampleIndex]
	fontSize := imgui.FontSize()

	s.context.draw.DrawScreenString(5.0, 1.5*fontSize, box2d.ColorYellow, "%s : %s", entry.category, entry.name)
	s.context.draw.DrawScreenString(5.0, 3.0*fontSize, box2d.ColorForestGreen, "Press TAB to show UI")
	s.screenTextY += 1.5 * fontSize

	if s.context.pause {
		s.context.draw.DrawScreenString(5.0, 4.5*fontSize, box2d.ColorRed, "****PAUSED****")
		s.screenTextY += 1.5 * fontSize
	}

	s.context.draw.DrawScreenString(5.0, s.camera.height-0.5*fontSize, box2d.ColorSeaGreen, "%.1f ms  step %d",
		1000.0*frameTime, s.stepCount)
}

func ParsePath(svgPath string, offset box2d.Vec2, points []box2d.Vec2, scale float32, reverseOrder bool) int {
	pointCount := 0
	var currentPoint box2d.Vec2
	ptr := 0
	at := func(i int) byte {
		if i < len(svgPath) {
			return svgPath[i]
		}
		return 0
	}
	isDigit := func(c byte) bool { return c >= '0' && c <= '9' }
	isSpace := func(c byte) bool { return c != 0 && unicode.IsSpace(rune(c)) }
	scanFloat := func(text string) float32 {
		v, err := strconv.ParseFloat(text, 32)
		if err != nil {
			panic(fmt.Sprintf("ParsePath: invalid number %q", text))
		}
		return float32(v)
	}
	scanPair := func(text string) (float32, float32) {
		x, y, ok := strings.Cut(text, ",")
		if !ok {
			panic(fmt.Sprintf("ParsePath: invalid pair %q", text))
		}
		return scanFloat(x), scanFloat(y)
	}
	command := at(0)

	for at(ptr) != 0 {
		if !isDigit(at(ptr)) && at(ptr) != '-' {
			command = at(ptr)

			switch command {
			case 'M', 'L', 'H', 'V', 'm', 'l', 'h', 'v':
				ptr += 2
			}

			if command == 'z' {
				break
			}
		}

		if !isDigit(at(ptr)) && at(ptr) != '-' {
			panic("ParsePath: expected number")
		}

		end := ptr
		for at(end) != 0 && !isSpace(at(end)) {
			end++
		}
		token := svgPath[ptr:end]

		switch command {
		case 'M', 'L':
			x, y := scanPair(token)
			currentPoint.X = x
			currentPoint.Y = y
		case 'H':
			currentPoint.X = scanFloat(token)
		case 'V':
			currentPoint.Y = scanFloat(token)
		case 'm', 'l':
			x, y := scanPair(token)
			currentPoint.X += x
			currentPoint.Y += y
		case 'h':
			currentPoint.X += scanFloat(token)
		case 'v':
			currentPoint.Y += scanFloat(token)
		default:
			panic("ParsePath: unknown command")
		}

		points[pointCount] = box2d.Vec2{X: scale * (currentPoint.X + offset.X), Y: -scale * (currentPoint.Y + offset.Y)}
		pointCount += 1
		if pointCount == len(points) {
			break
		}

		ptr = end

		for isSpace(at(ptr)) {
			ptr++
		}
	}

	if pointCount == 0 {
		return 0
	}

	if pointCount > 2 && points[0].Distance(points[pointCount-1]) <= box2d.LinearSlop() {
		pointCount -= 1
	}

	return pointCount
}

func fuzzyScore(needle, haystack string) int {
	if needle == "" {
		return 0
	}

	score := 0
	hi := 0
	prevMatchHi := -2

	lower := func(c byte) byte {
		if c >= 'A' && c <= 'Z' {
			return c + ('a' - 'A')
		}
		return c
	}
	isAlnum := func(c byte) bool {
		return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
	}

	for ni := 0; ni < len(needle); ni++ {
		nc := lower(needle[ni])
		for hi < len(haystack) && lower(haystack[hi]) != nc {
			hi++
		}
		if hi == len(haystack) {
			return -1
		}

		bonus := 1
		if hi == 0 {
			bonus += 10
		} else if !isAlnum(haystack[hi-1]) {
			bonus += 5
		}
		if hi == prevMatchHi+1 {
			bonus += 3
		}

		score += bonus
		prevMatchHi = hi
		hi++
	}

	return score
}

type SampleCreateFcn func(context *SampleContext) Sampler

type SampleCapacityFcn func() box2d.Capacity

type SampleEntry struct {
	category    string
	name        string
	createFcn   SampleCreateFcn
	capacityFcn SampleCapacityFcn
	replay      bool
}

var sampleEntries []SampleEntry

var replayIndex = -1

func RegisterSample(category, name string, fcn SampleCreateFcn) {
	sampleEntries = append(sampleEntries, SampleEntry{category, name, fcn, nil, false})
}

func RegisterSampleWithCapacity(category, name string, fcn SampleCreateFcn, capacityFcn SampleCapacityFcn) {
	sampleEntries = append(sampleEntries, SampleEntry{category, name, fcn, capacityFcn, false})
}

func RegisterReplay(category, name string, fcn SampleCreateFcn) {
	sampleEntries = append(sampleEntries, SampleEntry{category, name, fcn, nil, true})
}

func SelectSample(context *SampleContext, selection int, restart bool) {
	if !restart {
		ResetView(&context.camera)
		context.sampleIndex = selection
		context.pause = false
		context.subStepCount = 4
		context.restitutionIterations = 2
		context.enableRestitutionPropagation = false
		context.debugDraw.DrawJoints = true
	}

	context.singleStep = 0

	if context.sample != nil {
		context.sample.Destroy()
	}
	context.sample = nil
	if sampleEntries[context.sampleIndex].capacityFcn != nil {
		context.capacity = sampleEntries[context.sampleIndex].capacityFcn()
	} else {
		context.capacity = box2d.DefaultWorldDef().Capacity
	}
	context.restart = restart
	context.sample = sampleEntries[context.sampleIndex].createFcn(context)
	context.restart = false

	if !restart {
		context.homeCenter = context.camera.center
		context.homeZoom = context.camera.zoom
	}
}

func drawRow(key, desc string) {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.TextUnformatted(key)
	imgui.TableSetColumnIndex(1)
	imgui.TextUnformatted(desc)
}

var showAbout bool

func drawMenuBar(context *SampleContext) {
	fontSize := imgui.FontSize()

	if imgui.BeginMainMenuBar() {
		if imgui.BeginMenu("Sim") {
			imgui.MenuItemBoolPtr("Pause", "P", &context.pause)
			if imgui.MenuItemBoolV("Single Step", ".", false, true) {
				context.singleStep += 1
			}
			if imgui.MenuItemBoolV("Restart", "R", false, true) {
				SelectSample(context, context.sampleIndex, true)
			}
			imgui.Separator()
			if imgui.MenuItemBoolV("Previous Sample", "[", false, true) {
				selection := context.sampleIndex - 1
				if selection < 0 {
					selection = len(sampleEntries) - 1
				}
				SelectSample(context, selection, false)
			}
			if imgui.MenuItemBoolV("Next Sample", "]", false, true) {
				selection := context.sampleIndex + 1
				if selection == len(sampleEntries) {
					selection = 0
				}
				SelectSample(context, selection, false)
			}
			imgui.Separator()
			if imgui.MenuItemBool("Reset Profile") {
				context.sample.base().ResetProfile()
			}
			imgui.Separator()
			if imgui.MenuItemBoolV("Quit", "Ctrl+Q", false, true) {
				context.window.SetShouldClose(true)
			}
			imgui.EndMenu()
		}

		if imgui.BeginMenu("View") {
			if imgui.MenuItemBoolV("Hide UI", "Tab", false, true) {
				context.showUI = false
			}
			if imgui.MenuItemBoolV("Reset Camera", "Home", false, true) {
				context.sample.FocusHome()
			}
			imgui.Separator()
			imgui.MenuItemBoolPtr("Shapes", "", &context.debugDraw.DrawShapes)
			imgui.MenuItemBoolPtr("Chain Normals", "", &context.debugDraw.DrawChainNormals)
			imgui.MenuItemBoolPtr("Joints", "", &context.debugDraw.DrawJoints)
			imgui.MenuItemBoolPtr("Joint Extras", "", &context.debugDraw.DrawJointExtras)
			imgui.MenuItemBoolPtr("Bounds", "", &context.debugDraw.DrawBounds)
			imgui.MenuItemBoolPtr("Mass", "", &context.debugDraw.DrawMass)
			imgui.MenuItemBoolPtr("Body Names", "", &context.debugDraw.DrawBodyNames)
			imgui.MenuItemBoolPtr("Graph Colors", "", &context.debugDraw.DrawGraphColors)
			imgui.MenuItemBoolPtr("Islands", "", &context.debugDraw.DrawIslands)
			imgui.Separator()
			imgui.MenuItemBoolPtr("Contact Points", "", &context.debugDraw.DrawContacts)
			imgui.MenuItemBoolPtr("Contact Normals", "", &context.debugDraw.DrawContactNormals)
			imgui.MenuItemBoolPtr("Contact Features", "", &context.debugDraw.DrawContactFeatures)
			imgui.MenuItemBoolPtr("Contact Forces", "", &context.debugDraw.DrawContactForces)
			imgui.MenuItemBoolPtr("Friction Forces", "", &context.debugDraw.DrawFrictionForces)
			if imgui.BeginMenu("Anchor") {
				if imgui.MenuItemBoolV("Anchor A", "", context.debugDraw.DrawAnchorA, true) {
					context.debugDraw.DrawAnchorA = true
				}
				if imgui.MenuItemBoolV("Anchor B", "", !context.debugDraw.DrawAnchorA, true) {
					context.debugDraw.DrawAnchorA = false
				}
				imgui.EndMenu()
			}
			imgui.Separator()
			imgui.MenuItemBoolPtr("Metrics", "M", &context.showMetrics)
			imgui.MenuItemBoolPtrV("Profile", "I", &context.showProfile, context.sample.HasProfile())
			imgui.Separator()
			if imgui.BeginMenu("Scale") {
				imgui.PushItemWidth(6.0 * fontSize)
				imgui.InputFloat("Joint", &context.debugDraw.JointScale)
				imgui.InputFloat("Force", &context.debugDraw.ForceScale)
				imgui.PopItemWidth()
				imgui.EndMenu()
			}
			imgui.EndMenu()
		}

		if imgui.BeginMenu("Samples") {
			i := 0
			for i < len(sampleEntries) {
				category := sampleEntries[i].category
				if imgui.BeginMenu(category) {
					for i < len(sampleEntries) && category == sampleEntries[i].category {
						selected := i == context.sampleIndex
						if imgui.MenuItemBoolV(sampleEntries[i].name, "", selected, true) {
							SelectSample(context, i, false)
						}
						i++
					}
					imgui.EndMenu()
				} else {
					for i < len(sampleEntries) && category == sampleEntries[i].category {
						i++
					}
				}
			}
			imgui.EndMenu()
		}

		if replayIndex >= 0 && imgui.BeginMenu("Replay") {
			if imgui.MenuItemBool("Open...") {
				outPath, err := zenity.SelectFile(zenity.FileFilter{Name: "Box2D recording", Patterns: []string{"*.b2rec"}})
				if err == nil {
					context.replayFile = outPath
					SelectSample(context, replayIndex, false)
				} else if err != zenity.ErrCanceled {
					fmt.Fprintf(os.Stderr, "open dialog failed: %v\n", err)
				}
			}
			imgui.EndMenu()
		}

		if imgui.BeginMenu("Help") {
			imgui.MenuItemBoolPtr("Controls", "?", &context.showControls)
			imgui.MenuItemBoolPtr("About", "", &showAbout)
			imgui.EndMenu()
		}

		imgui.EndMainMenuBar()

		{
			menuBarBottom := imgui.FrameHeight()
			borderColor := imgui.ColorU32Col(imgui.ColBorder)
			displaySize := imgui.CurrentIO().DisplaySize()
			imgui.ForegroundDrawListViewportPtr().AddLineArgs(imgui.Vec2{X: 0.0, Y: menuBarBottom}, imgui.Vec2{X: displaySize.X, Y: menuBarBottom},
				borderColor, 1.0)
		}

		if context.showControls {
			imgui.SetNextWindowPosV(imgui.Vec2{X: context.camera.width * 0.5, Y: context.camera.height * 0.5}, imgui.CondAppearing,
				imgui.Vec2{X: 0.5, Y: 0.5})
			imgui.SetNextWindowSizeV(imgui.Vec2{X: 24.0 * fontSize, Y: 0.0}, imgui.CondAppearing)

			if imgui.BeginV("Controls", &context.showControls, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsAlwaysAutoResize) {
				imgui.SeparatorText("Keyboard")
				if imgui.BeginTableV("keys", 2, imgui.TableFlagsSizingFixedFit, imgui.Vec2{}, 0.0) {
					drawRow("Tab", "Show / hide UI")
					drawRow("M", "Show / hide metrics")
					drawRow("I", "Show / hide profile")
					drawRow("P", "Pause / resume")
					drawRow(".", "Single step (Shift: 5)")
					drawRow(",", "Step back, replay only (Shift: 5)")
					drawRow("R", "Restart sample")
					drawRow("[  ]", "Previous / next sample")
					drawRow("Ctrl+O", "Open sample picker")
					drawRow("Arrows", "Pan camera")
					drawRow("Z  X", "Zoom out / in")
					drawRow("Home", "Reset camera")
					drawRow("?", "Show / hide controls")
					drawRow("Esc", "Cancel / close")
					drawRow("Ctrl+Q", "Quit")
					imgui.EndTable()
				}

				imgui.SeparatorText("Mouse")
				if imgui.BeginTableV("mouse", 2, imgui.TableFlagsSizingFixedFit, imgui.Vec2{}, 0.0) {
					drawRow("Left drag", "Move bodies (mouse joint)")
					drawRow("Right drag", "Pan camera")
					drawRow("Scroll wheel", "Zoom")
					imgui.EndTable()
				}
			}
			imgui.End()
		}

		if showAbout {
			imgui.SetNextWindowPosV(imgui.Vec2{X: context.camera.width * 0.5, Y: context.camera.height * 0.5}, imgui.CondAppearing,
				imgui.Vec2{X: 0.5, Y: 0.5})
			imgui.SetNextWindowSizeV(imgui.Vec2{X: 22.0 * fontSize, Y: 0.0}, imgui.CondAppearing)

			if imgui.BeginV("About", &showAbout, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsAlwaysAutoResize) {
				version := box2d.CurrentVersion()
				imText("Box2D %d.%d.%d", version.Major, version.Minor, version.Revision)
				imgui.Spacing()
				imgui.TextLinkOpenURLV("box2d.org", "https://box2d.org/")
				imgui.TextLinkOpenURLV("github.com/erincatto/box2d", "https://github.com/erincatto/box2d")
			}
			imgui.End()
		}
	}
}

type scored struct {
	idx   int
	score int
}

func rebuild(q string) []int {
	var list []scored
	for i := range sampleEntries {
		nameScore := fuzzyScore(q, sampleEntries[i].name)
		catScore := fuzzyScore(q, sampleEntries[i].category)
		best := -1
		if nameScore >= 0 {
			best = nameScore * 2
		}
		if catScore >= 0 && catScore > best {
			best = catScore
		}
		if best < 0 {
			continue
		}
		list = append(list, scored{i, best})
	}

	for i := 1; i < len(list); i++ {
		tmp := list[i]
		j := i - 1
		for j >= 0 && list[j].score < tmp.score {
			list[j+1] = list[j]
			j--
		}
		list[j+1] = tmp
	}

	filtered := make([]int, len(list))
	for i := range list {
		filtered[i] = list[i].idx
	}
	return filtered
}

var picker struct {
	query         string
	prevQuery     string
	highlight     int
	prevHighlight int
	filtered      []int
	justOpened    bool
	forceScroll   bool
}

func drawSamplePicker(context *SampleContext) {
	fontSize := imgui.FontSize()

	if context.openSamplePicker {
		imgui.OpenPopupStr("##sample_picker")
		context.openSamplePicker = false
		picker.query = ""
		picker.prevQuery = ""
		picker.highlight = 0
		picker.prevHighlight = 0
		picker.filtered = rebuild(picker.query)
		picker.justOpened = true
		picker.forceScroll = true
	}

	imgui.SetNextWindowPosV(imgui.Vec2{X: context.camera.width * 0.5, Y: context.camera.height * 0.35}, imgui.CondAppearing,
		imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 32.0 * fontSize, Y: 0.0}, imgui.CondAppearing)

	if imgui.BeginPopupV("##sample_picker",
		imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoSavedSettings) {
		if picker.justOpened {
			imgui.SetKeyboardFocusHere()
			picker.justOpened = false
		}

		imgui.PushItemWidth(-1.0)
		imgui.InputTextWithHint("##q", "Search by name or category...", &picker.query, 0, nil)
		imgui.PopItemWidth()
		if len(picker.query) > 63 {
			picker.query = picker.query[:63]
		}

		if picker.query != picker.prevQuery {
			picker.filtered = rebuild(picker.query)
			picker.prevQuery = picker.query
			picker.highlight = 0
			picker.forceScroll = true
		}

		filteredCount := len(picker.filtered)
		if filteredCount > 0 {
			if imgui.IsKeyPressedBoolV(imgui.KeyDownArrow, true) {
				picker.highlight = (picker.highlight + 1) % filteredCount
			}
			if imgui.IsKeyPressedBoolV(imgui.KeyUpArrow, true) {
				picker.highlight = (picker.highlight + filteredCount - 1) % filteredCount
			}
		}
		commit := imgui.IsKeyPressedBoolV(imgui.KeyEnter, false) || imgui.IsKeyPressedBoolV(imgui.KeyKeypadEnter, false)

		imgui.BeginChildStrV("##results", imgui.Vec2{X: 0.0, Y: 14.0 * fontSize}, imgui.ChildFlagsBorders, 0)
		for row := range filteredCount {
			i := picker.filtered[row]
			label := fmt.Sprintf("%s  >  %s", sampleEntries[i].category, sampleEntries[i].name)
			sel := row == picker.highlight
			if imgui.SelectableBoolV(label, sel, 0, imgui.Vec2{}) {
				picker.highlight = row
				commit = true
			}
			if sel && (picker.forceScroll || picker.highlight != picker.prevHighlight) {
				imgui.SetScrollHereY()
			}
		}
		imgui.EndChild()
		picker.prevHighlight = picker.highlight
		picker.forceScroll = false

		if commit && filteredCount > 0 {
			SelectSample(context, picker.filtered[picker.highlight], false)
			imgui.CloseCurrentPopup()
		}

		if imgui.IsKeyPressedBoolV(imgui.KeyEscape, false) {
			imgui.CloseCurrentPopup()
		}

		imgui.EndPopup()
	}
}

func helpMarker(text string) {
	imgui.SameLineV(0.0, imgui.CurrentStyle().ItemInnerSpacing().X)
	width := imgui.FontSize()
	height := imgui.FrameHeight()
	pos := imgui.CursorScreenPos()
	imgui.Dummy(imgui.Vec2{X: width, Y: height})
	hovered := imgui.IsItemHovered()
	alpha := float32(0.6)
	if hovered {
		alpha = 1.0
	}
	color := imgui.ColorU32ColV(imgui.ColCheckMark, alpha)
	center := imgui.Vec2{X: pos.X + 0.5*width, Y: pos.Y + 0.5*height}
	imgui.WindowDrawList().AddCircleFilled(center, 0.2*width, color)
	imSetItemTooltip("%s", text)
}

func drawInfoPanel(context *SampleContext, frameTime float32) {
	entry := &sampleEntries[context.sampleIndex]
	sample := context.sample.base()
	fontSize := imgui.FontSize()
	menuWidth := context.sample.InfoPanelWidthEm() * fontSize
	menuBarHeight := imgui.FrameHeight()

	imgui.SetNextWindowPos(imgui.Vec2{X: context.camera.width - menuWidth - 0.5*fontSize, Y: menuBarHeight + 0.5*fontSize})
	imgui.SetNextWindowSize(imgui.Vec2{X: menuWidth, Y: context.camera.height - menuBarHeight - fontSize})

	imgui.BeginV("Info", nil,
		imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoTitleBar)

	imTextColored(MakeColor(box2d.ColorGoldenRod), "%s", entry.name)
	imTextColored(MakeColor(box2d.ColorLightGray), "%s", entry.category)
	imgui.Separator()

	if context.pause {
		imTextColored(MakeColor(box2d.ColorRed), "PAUSED")
		imgui.SameLine()
		imgui.TextDisabled("(P)")
		imgui.Separator()
	}

	imTextColored(MakeColor(box2d.ColorSeaGreen), "frame %.1f ms", 1000.0*frameTime)
	imTextColored(MakeColor(box2d.ColorSeaGreen), "step %d", sample.stepCount)
	imgui.Separator()
	imTextColored(MakeColor(box2d.ColorSeaGreen), "cam (%.1f, %.1f)", context.camera.center.X, context.camera.center.Y)
	imTextColored(MakeColor(box2d.ColorSeaGreen), "zoom %.2f", context.camera.zoom)

	imgui.Separator()

	imgui.PushItemWidth(6.0 * fontSize)
	if context.sample.DrawControls() {
		imgui.Separator()
	}
	imgui.PopItemWidth()

	if context.sample.HasSolverControls() && imgui.CollapsingHeaderTreeNodeFlagsV("Solver", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.PushItemWidth(6.0 * fontSize)

		imgui.SliderInt("Sub-steps##Solver", &context.subStepCount, 1, 32)
		helpMarker("The solver breaks the full step into several sub-steps.\nMore sub-steps usually lead to more accurate results.")

		imgui.SliderInt("Bounce Iters##Solver", &context.restitutionIterations, 0, 8)
		helpMarker("Iterations for the restitution solver.")

		imgui.SliderFloatV("Hertz##Solver", &context.hertz, 5.0, 240.0, "%.0f Hz", 0)
		helpMarker("The number of world steps per second.")

		if imgui.SliderInt("Workers##Solver", &context.workerCount, 1, box2d.MaxWorkers) {
			context.workerCount = min(max(context.workerCount, 1), box2d.MaxWorkers)
			SelectSample(context, context.sampleIndex, true)
			sample = context.sample.base()
		}
		helpMarker("The number worker threads used by the world step.")

		recyclingCentimeters := 100.0 * context.recycleDistance
		if imgui.SliderFloatV("Recycle##Solver", &recyclingCentimeters, 0.0, 10.0, "%.1f cm", 0) {
			context.recycleDistance = 0.01 * recyclingCentimeters
			sample.worldID.SetContactRecycleDistance(context.recycleDistance)
		}
		helpMarker("The contact recycling distance tolerance.\nSet to zero to disable recycling.")

		imgui.PopItemWidth()

		imgui.Checkbox("Sleep##Solver", &context.enableSleep)
		helpMarker("Allow bodies to sleep, reducing simulation CPU cost.")

		imgui.Checkbox("Warm Starting##Solver", &context.enableWarmStarting)
		helpMarker("Enable solver warm starting which usually improves stacking stability.")

		imgui.Checkbox("Continuous##Solver", &context.enableContinuous)
		helpMarker("Enable continuous collision detection.")

		imgui.Checkbox("Bounce Propagation##Solver", &context.enableRestitutionPropagation)
		helpMarker("Enable restitution solver propagation across all touching contacts points")
	}

	if context.sample.HasSolverControls() && imgui.CollapsingHeaderTreeNodeFlagsV("Recording", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.PushItemWidth(9.0 * fontSize)
		imgui.InputTextWithHint("File##Recording", "", &context.recordingFile, 0, nil)
		imgui.PopItemWidth()
		if len(context.recordingFile) > 255 {
			context.recordingFile = context.recordingFile[:255]
		}

		if sample.recording == nil {
			if imgui.Button("Record (restart)##Recording") {
				SelectSample(context, context.sampleIndex, true)
				sample = context.sample.base()
				sample.StartRecording()
			}

			if imgui.Button("Record Now##Recording") {
				sample.StartRecording()
			}

			if replayIndex >= 0 && context.savedRecordingFile != "" {
				if imgui.Button("Play##Recording") {
					context.replayFile = context.savedRecordingFile
					SelectSample(context, replayIndex, false)
					sample = context.sample.base()
				}
				imSetItemTooltip("Open %s in the replay viewer", context.savedRecordingFile)
			}
		} else {
			if imgui.Button("Stop##Recording") {
				sample.FinishRecording()
			}
		}

		if sample.recording != nil {
			kilobytes := float32(sample.recording.Size()) / 1024.0
			steps := sample.stepCount - sample.recordStartStep
			if kilobytes < 1024.0 {
				imTextColored(MakeColor(box2d.ColorRed), "recording %d steps, %.0f KB", steps, kilobytes)
			} else {
				imTextColored(MakeColor(box2d.ColorRed), "recording %d steps, %.1f MB", steps, kilobytes/1024.0)
			}
		}
	}

	imgui.End()
}

func DrawUI(context *SampleContext, frameTime float32) {
	drawMenuBar(context)
	drawSamplePicker(context)
	drawInfoPanel(context, frameTime)
	context.sample.base().DrawProfile()
	context.sample.base().DrawMetrics()
}

func imCol32(r, g, b, a uint8) uint32 {
	return uint32(a)<<24 | uint32(b)<<16 | uint32(g)<<8 | uint32(r)
}

func MakeColor(hexColor box2d.HexColor) imgui.Vec4 {
	return imgui.ColorConvertU32ToFloat4(imCol32(uint8(hexColor>>16), uint8(hexColor>>8), uint8(hexColor), 255))
}

func clampFloat(a, lower, upper float32) float32 {
	return min(max(a, lower), upper)
}

func escapeFormat(format string, args []any) string {
	return strings.ReplaceAll(fmt.Sprintf(format, args...), "%", "%%")
}

func imText(format string, args ...any) {
	imgui.Text(escapeFormat(format, args))
}

func imTextColored(color imgui.Vec4, format string, args ...any) {
	imgui.TextColored(color, escapeFormat(format, args))
}

func imTextDisabled(format string, args ...any) {
	imgui.TextDisabled(escapeFormat(format, args))
}

func imTextWrapped(format string, args ...any) {
	imgui.TextWrapped(escapeFormat(format, args))
}

func imSetItemTooltip(format string, args ...any) {
	imgui.SetItemTooltip(escapeFormat(format, args))
}
