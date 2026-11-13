// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

const (
	initialScreenTextX     = 5.0
	initialMouseForceScale = 100.0
	screenTextLineHeight   = 1.5
	panelMarginEm          = 0.5
	millisecondsPerSecond  = 1000.0
	percentScale           = 100.0
	infoPanelWidth         = 16.0
	fltMin                 = 1.17549435e-38
	settingsFileName       = "settings.ini"
	maxTasks               = 512
	maxThreads             = 64
	profileCapacity        = 512

	profileSectionWidth         = 8.0
	profileIndentWidth          = 0.75
	profileValueWidth           = 3.0
	profileBarWidth             = 4.0
	profileColumnCount          = 5
	profileRowCount             = 22
	colorStep            uint32 = 0xffe6e6e6
	colorPairs           uint32 = 0xffff9966
	colorCollide         uint32 = 0xff338cff
	colorSolve           uint32 = 0xff66cc66
	colorSolveChild      uint32 = 0xff96be96
	colorConstraintChild uint32 = 0xffbec86e
	colorSensors         uint32 = 0xffdc78c8
	colorOther           uint32 = 0xff8c8c8c
	colorTrack           uint32 = 0xff4a4646
)

var (
	sampleEntries []SampleEntry
	replayIndex   = -1
	showAbout     bool
	picker        struct {
		query         string
		prevQuery     string
		highlight     int
		prevHighlight int
		filtered      []int
		justOpened    bool
		forceScroll   bool
	}
)

type SampleContext struct {
	window    *glfw.Window    //exhaustruct:optional
	camera    Camera          //exhaustruct:optional
	draw      *Draw           //exhaustruct:optional
	sample    Sampler         //exhaustruct:optional
	capacity  box2d.Capacity  //exhaustruct:optional
	debugDraw box2d.DebugDraw //exhaustruct:optional

	uiScale                      float32
	hertz                        float32
	recycleDistance              float32
	subStepCount                 int32
	restitutionIterations        int32
	workerCount                  int32
	restart                      bool //exhaustruct:optional
	pause                        bool //exhaustruct:optional
	singleStep                   int  //exhaustruct:optional
	enableWarmStarting           bool
	enableContinuous             bool
	enableRestitutionPropagation bool //exhaustruct:optional
	enableSleep                  bool
	showUI                       bool

	showMetrics bool //exhaustruct:optional

	showProfile bool //exhaustruct:optional

	showControls bool //exhaustruct:optional

	homeCenter box2d.Pos
	homeZoom   float32

	openSamplePicker bool //exhaustruct:optional

	recordingFile string

	replayFile string //exhaustruct:optional

	savedRecordingFile string //exhaustruct:optional

	sampleIndex int //exhaustruct:optional
	newUser     bool

	replayKeyframeBudgetMB    int32
	replayKeyframeMinInterval int32
}
type debugDrawer struct {
	context *SampleContext
}
type samplePresentation interface {
	HasProfile() bool
	InfoPanelWidthEm() float32
	DrawMetricsTab()
}
type sampleInput interface {
	Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey)
	FocusHome()
	MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey)
	MouseUp(p box2d.Pos, button glfw.MouseButton)
	MouseMove(p box2d.Pos)
}
type Sampler interface {
	samplePresentation
	sampleInput
	base() *Sample
	Step()
	DrawControls() bool
	HasSolverControls() bool
	Destroy()
}
type Sample struct {
	self samplePresentation //exhaustruct:optional

	context *SampleContext
	camera  *Camera
	draw    *Draw

	mouseBodyID box2d.Body //exhaustruct:optional

	worldID         *box2d.World     //exhaustruct:optional
	recording       *box2d.Recording //exhaustruct:optional
	recordStartStep int              //exhaustruct:optional
	mouseJointID    box2d.Joint      //exhaustruct:optional
	mousePoint      box2d.Pos        //exhaustruct:optional
	mouseForceScale float32
	stepCount       int //exhaustruct:optional
	screenTextX     float32
	screenTextY     float32 //exhaustruct:optional

	profiles            [profileCapacity]box2d.Profile //exhaustruct:optional
	currentProfileIndex int                            //exhaustruct:optional
	profileReadIndex    uint64                         //exhaustruct:optional
	profileWriteIndex   uint64                         //exhaustruct:optional

	didStep bool //exhaustruct:optional
}
type rowDef struct {
	name   string
	indent int
	color  uint32
}
type (
	SampleCreateFcn   func(context *SampleContext) Sampler
	SampleCapacityFcn func() box2d.Capacity
	SampleEntry       struct {
		category    string
		name        string
		createFcn   SampleCreateFcn
		capacityFcn SampleCapacityFcn
		replay      bool
	}
)

type scored struct {
	idx   int
	score int
}

func ParsePath(
	svgPath string,
	offset box2d.Vec2,
	points []box2d.Vec2,
	scale float32,
) int {
	pointCount := 0

	var currentPoint box2d.Vec2

	ptr := 0
	at := func(i int) byte {
		if i < len(svgPath) {
			return svgPath[i]
		}

		return 0
	}
	command := at(0)

	for at(ptr) != 0 {
		if !isPathNumberStart(at(ptr)) {
			command = at(ptr)

			switch command {
			case 'M', 'L', 'H', 'V', 'm', 'l', 'h', 'v':
				ptr += 2
			}

			if command == 'z' {
				break
			}
		}

		if !isPathNumberStart(at(ptr)) {
			panic("ParsePath: expected number")
		}

		end := ptr
		for at(end) != 0 && !isSpace(at(end)) {
			end++
		}

		token := svgPath[ptr:end]

		applyPathCommand(&currentPoint, command, token)

		points[pointCount] = box2d.Vec2{
			X: scale * (currentPoint.X + offset.X),
			Y: -scale * (currentPoint.Y + offset.Y),
		}

		pointCount += 1
		if pointCount == len(points) {
			break
		}

		ptr = end

		for isSpace(at(ptr)) {
			ptr++
		}
	}

	return pathPointCount(points, pointCount)
}

func RegisterSample(category string, name string, fcn SampleCreateFcn) {
	sampleEntries = append(sampleEntries, SampleEntry{category, name, fcn, nil, false})
}

func RegisterSampleWithCapacity(
	category string, name string,
	fcn SampleCreateFcn,
	capacityFcn SampleCapacityFcn,
) {
	sampleEntries = append(sampleEntries, SampleEntry{category, name, fcn, capacityFcn, false})
}

func RegisterReplay(category string, name string, fcn SampleCreateFcn) {
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

func DrawUI(context *SampleContext, frameTime float32) {
	drawMenuBar(context)
	drawSamplePicker(context)
	drawInfoPanel(context, frameTime)
	context.sample.base().DrawProfile()
	context.sample.base().DrawMetrics()
}

func MakeColor(hexColor box2d.HexColor) imgui.Vec4 {
	const (
		redShift   = 16
		greenShift = 8
	)

	return imgui.ColorConvertU32ToFloat4(
		imCol32(
			uint8(hexColor>>redShift),
			uint8(hexColor>>greenShift),
			uint8(hexColor),
		),
	)
}

func pathPointCount(points []box2d.Vec2, count int) int {
	if count > 2 && points[0].Distance(points[count-1]) <= box2d.LinearSlop() {
		return count - 1
	}

	return count
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isPathNumberStart(c byte) bool { return isDigit(c) || c == '-' }

func isSpace(c byte) bool { return c != 0 && unicode.IsSpace(rune(c)) }

func parsePathFloat(text string) float32 {
	const floatBits = 32

	v, err := strconv.ParseFloat(text, floatBits)
	if err != nil {
		panic(fmt.Sprintf("ParsePath: invalid number %q", text))
	}

	return float32(v)
}

func parsePathPair(text string) (float32, float32) {
	x, y, ok := strings.Cut(text, ",")
	if !ok {
		panic(fmt.Sprintf("ParsePath: invalid pair %q", text))
	}

	return parsePathFloat(x), parsePathFloat(y)
}

func applyPathCommand(point *box2d.Vec2, command byte, token string) {
	switch command {
	case 'M', 'L':
		x, y := parsePathPair(token)
		point.X = x
		point.Y = y
	case 'H':
		point.X = parsePathFloat(token)
	case 'V':
		point.Y = parsePathFloat(token)
	case 'm', 'l':
		x, y := parsePathPair(token)
		point.X += x
		point.Y += y
	case 'h':
		point.X += parsePathFloat(token)
	case 'v':
		point.Y += parsePathFloat(token)
	default:
		panic("ParsePath: unknown command")
	}
}

func newSampleContext() *SampleContext {
	const (
		defaultHertz                  = 60.0
		defaultRecycleDistance        = 0.05
		defaultSubSteps               = 4
		defaultRestitutionIterations  = 2
		defaultReplayBudgetMB         = 512
		defaultReplayKeyframeInterval = 16
	)

	return &SampleContext{
		uiScale:                   1.0,
		hertz:                     defaultHertz,
		recycleDistance:           defaultRecycleDistance,
		subStepCount:              defaultSubSteps,
		restitutionIterations:     defaultRestitutionIterations,
		workerCount:               1,
		enableWarmStarting:        true,
		enableContinuous:          true,
		enableSleep:               true,
		showUI:                    true,
		homeCenter:                box2d.Pos{X: 0.0, Y: defaultCameraCenterY},
		homeZoom:                  1.0,
		recordingFile:             "recording.b2rec",
		newUser:                   true,
		replayKeyframeBudgetMB:    defaultReplayBudgetMB,
		replayKeyframeMinInterval: defaultReplayKeyframeInterval,
	}
}

func (context *SampleContext) Save() {
	file, err := os.Create(settingsFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", settingsFileName, err)
		return
	}

	_, writeErr := fmt.Fprintf(file,
		"{\n  \"sampleIndex\": %d,\n  \"newUser\": %d,\n"+
			"  \"drawShapes\": %t,\n  \"showMetrics\": %t,\n  \"showProfile\": %t,\n"+
			"  \"replayKeyframeBudgetMB\": %d,\n  \"replayKeyframeMinInterval\": %d\n}\n",
		context.sampleIndex, 0, context.debugDraw.DrawShapes, context.showMetrics,
		context.showProfile, context.replayKeyframeBudgetMB, context.replayKeyframeMinInterval,
	)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", settingsFileName, err)
	}
}

func (d debugDrawer) DrawPolygon(
	transform box2d.WorldTransform,
	vertices []box2d.Vec2,
	color box2d.HexColor,
) {
	d.context.draw.DrawPolygon(transform, vertices, color)
}

func (d debugDrawer) DrawSolidPolygon(
	transform box2d.WorldTransform,
	vertices []box2d.Vec2,
	radius float32,
	color box2d.HexColor,
) {
	d.context.draw.DrawSolidPolygon(transform, vertices, radius, color)
}

func (d debugDrawer) DrawCircle(center box2d.Pos, radius float32, color box2d.HexColor) {
	d.context.draw.DrawCircle(center, radius, color)
}

func (d debugDrawer) DrawSolidCircle(
	transform box2d.WorldTransform,
	center box2d.Vec2,
	radius float32,
	color box2d.HexColor,
) {
	d.context.draw.DrawSolidCircle(transform, center, radius, color)
}

func (d debugDrawer) DrawSolidCapsule(
	p1 box2d.Pos,
	p2 box2d.Pos,
	radius float32,
	color box2d.HexColor,
) {
	d.context.draw.DrawCapsule(p1, p2, radius, color)
}

func (d debugDrawer) DrawLine(p1 box2d.Pos, p2 box2d.Pos, color box2d.HexColor) {
	d.context.draw.DrawLine(p1, p2, color)
}

func (d debugDrawer) DrawTransform(transform box2d.WorldTransform) {
	d.context.draw.DrawTransform(transform, 1.0)
}

func (d debugDrawer) DrawPoint(p box2d.Pos, size float32, color box2d.HexColor) {
	d.context.draw.DrawPoint(p, size, color)
}

func (d debugDrawer) DrawString(p box2d.Pos, s string, color box2d.HexColor) {
	d.context.draw.DrawStringf(&d.context.camera, p, color, "%s", s)
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

	ms := float32(time.Since(start).Seconds() * millisecondsPerSecond)
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
		return fmt.Errorf("create recording: %w", err)
	}

	_, writeErr := recording.WriteTo(file)

	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("write recording: %w", err)
	}

	return nil
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
	const (
		hudTopMarginLines = 3.0
		profileTextGapEm  = 0.5
	)

	fontSize := imgui.FontSize()
	if s.context.showUI {
		s.screenTextY = imgui.FrameHeight() + screenTextLineHeight*fontSize
	} else {
		s.screenTextY = hudTopMarginLines * fontSize
	}

	s.screenTextX = initialScreenTextX
	if s.IsProfileVisible() {
		s.screenTextX += s.GetProfilePanelWidth() + profileTextGapEm*fontSize
	}
}

//nolint:unparam // Overrides receive modifier keys through the shared input callback contract.
func (s *Sample) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if s.mouseJointID.IsValid() || button != glfw.MouseButton1 {
		return
	}

	const queryHalfExtent = 0.001

	d := box2d.Vec2{X: queryHalfExtent, Y: queryHalfExtent}
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

			const leverTorqueScale = 0.25

			jointDef.MaxVelocityTorque = leverTorqueScale * lever * mg
		}

		s.mouseJointID = s.worldID.CreateMotorJoint(&jointDef).Joint
	}
}

//nolint:unparam // Overrides receive the release position through the shared input callback contract.
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

func (s *Sample) DrawScreenTextLinef(format string, args ...any) {
	s.draw.DrawScreenStringf(s.screenTextX, s.screenTextY, box2d.ColorWhite, format, args...)
	s.screenTextY += screenTextLineHeight * imgui.FontSize()
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
		s.mouseBodyID.SetTargetTransform(
			box2d.WorldTransform{P: s.mousePoint, Q: box2d.RotIdentity},
			timeStep,
			wake,
		)
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

func addSegment(
	dl *imgui.DrawList,
	availWidth float32, t float32, stepNow float32,
	col uint32,
	x float32,
	cursor imgui.Vec2,
	barHeight float32,
) float32 {
	w := availWidth * (t / stepNow)
	if w > 0.0 {
		dl.AddRectFilled(
			imgui.Vec2{X: x, Y: cursor.Y},
			imgui.Vec2{X: x + w, Y: cursor.Y + barHeight},
			col,
		)
		x += w
	}

	return x
}

func (s *Sample) IsProfileVisible() bool {
	return s.context.showUI && s.context.showProfile && s.self.HasProfile()
}

func (s *Sample) GetProfilePanelWidth() float32 {
	if !s.IsProfileVisible() {
		return 0.0
	}

	style := imgui.CurrentStyle()
	fontSize := imgui.FontSize()

	const (
		profileValueColumns    = 3.0
		horizontalPaddingSides = 2.0
	)

	width := float32(profileSectionWidth + profileValueColumns*profileValueWidth + profileBarWidth)

	return width*fontSize + horizontalPaddingSides*profileColumnCount*style.CellPadding().X +
		horizontalPaddingSides*style.WindowPadding().X
}

func (s *Sample) DrawProfile() {
	if !s.IsProfileVisible() {
		return
	}

	fontSize := imgui.FontSize()
	menuBarHeight := imgui.FrameHeight()
	panelWidth := s.GetProfilePanelWidth()

	imgui.SetNextWindowPos(
		imgui.Vec2{X: panelMarginEm * fontSize, Y: menuBarHeight + panelMarginEm*fontSize},
	)
	imgui.SetNextWindowSize(
		imgui.Vec2{X: panelWidth, Y: s.camera.height - menuBarHeight - fontSize},
	)

	imgui.BeginV(
		"Profile",
		nil,
		imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoTitleBar,
	)

	count := int(s.profileWriteIndex - s.profileReadIndex)

	const rowCount = profileRowCount

	var (
		histories [rowCount][profileCapacity]float32
		totals    [rowCount]float32
	)

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

	var now, avg, rowMax [rowCount]float32
	summarizeProfile(&histories, &totals, count, &now, &avg, &rowMax)

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

	const minimumDisplayedStepMS = 0.001

	stepNow := max(now[0], minimumDisplayedStepMS)

	if imgui.Button("Reset") {
		s.ResetProfile()
	}

	imgui.SameLine()
	imTextf(" step %.2f ms", now[0])

	{
		pairsT := now[1]
		collideT := now[2]
		solveT := now[3]
		sensorsT := now[21]
		otherT := max(stepNow-pairsT-collideT-solveT-sensorsT, 0.0)

		availWidth := imgui.ContentRegionAvail().X

		const profileBarHeightEm = 1.5

		barHeight := profileBarHeightEm * fontSize
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

	s.drawProfileTable(rows[:], now[:], avg[:], rowMax[:], stepNow, fontSize)

	imgui.End()
}

func (s *Sample) DrawMetrics() {
	if !s.context.showMetrics {
		return
	}

	fontSize := imgui.FontSize()
	menuWidth := s.self.InfoPanelWidthEm() * fontSize

	const metricsHeightEm = 16.0

	drawerHeight := metricsHeightEm * fontSize

	drawerX := panelMarginEm * fontSize
	if s.IsProfileVisible() {
		drawerX += s.GetProfilePanelWidth() + panelMarginEm*fontSize
	}

	drawerWidth := s.camera.width - menuWidth - fontSize - drawerX

	imgui.SetNextWindowPos(
		imgui.Vec2{X: drawerX, Y: s.camera.height - drawerHeight - panelMarginEm*fontSize},
	)
	imgui.SetNextWindowSize(imgui.Vec2{X: drawerWidth, Y: drawerHeight})

	imgui.BeginV(
		"Metrics",
		nil,
		imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoTitleBar,
	)

	if imgui.BeginTabBarV("MetricsTabs", imgui.TabBarFlagsNone) {
		s.drawFrameTimeTab()
		s.drawCountersTab(fontSize)

		s.self.DrawMetricsTab()

		imgui.EndTabBar()
	}

	imgui.End()
}

func plotProfileLine(label string, color uint32, xs []float32, ys []float32) {
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
	const (
		hudHelpLine       = 3.0
		hudPausedLine     = 4.5
		hudBottomMarginEm = 0.5
	)

	entry := &sampleEntries[s.context.sampleIndex]
	fontSize := imgui.FontSize()

	s.context.draw.DrawScreenStringf(
		initialScreenTextX,
		screenTextLineHeight*fontSize,
		box2d.ColorYellow,
		"%s : %s",
		entry.category,
		entry.name,
	)
	s.context.draw.DrawScreenStringf(
		initialScreenTextX,
		hudHelpLine*fontSize,
		box2d.ColorForestGreen,
		"Press TAB to show UI",
	)
	s.screenTextY += screenTextLineHeight * fontSize

	if s.context.pause {
		s.context.draw.DrawScreenStringf(
			initialScreenTextX,
			hudPausedLine*fontSize,
			box2d.ColorRed,
			"****PAUSED****",
		)
		s.screenTextY += screenTextLineHeight * fontSize
	}

	s.context.draw.DrawScreenStringf(
		initialScreenTextX,
		s.camera.height-hudBottomMarginEm*fontSize,
		box2d.ColorSeaGreen,
		"%.1f ms  step %d",
		millisecondsPerSecond*frameTime,
		s.stepCount,
	)
}

func (s *Sample) drawFrameTimeTab() {
	const (
		profileSamplesPerSecond        = 60.0
		profileGridColor        uint32 = 0x0fffffff
	)

	if s.self.HasProfile() && imgui.BeginTabItem("Frame Time") {
		var (
			maxValue     float32
			times        [profileCapacity]float32
			stepTimes    [profileCapacity]float32
			pairsTimes   [profileCapacity]float32
			collideTimes [profileCapacity]float32
			solveTimes   [profileCapacity]float32
			sensorsTimes [profileCapacity]float32
		)

		count := int(s.profileWriteIndex - s.profileReadIndex)
		for i := range count {
			index := (s.profileReadIndex + uint64(i)) & (profileCapacity - 1)
			times[i] = float32(i) / profileSamplesPerSecond
			stepTimes[i] = s.profiles[index].Step
			pairsTimes[i] = s.profiles[index].Pairs
			collideTimes[i] = s.profiles[index].Collide
			solveTimes[i] = s.profiles[index].Solve
			sensorsTimes[i] = s.profiles[index].Sensors
			maxValue = max(stepTimes[i], maxValue)
		}

		implot.PushStyleVarFloat(implot.StyleVarMinorAlpha, 1.0)
		implot.PushStyleColorU32(implot.ColAxisGrid, profileGridColor)

		plotSize := imgui.ContentRegionAvail()
		if implot.BeginPlotV("Profile", plotSize, implot.FlagsNoTitle) {
			implot.SetupAxesV("", "ms", 0, 0)
			implot.SetupAxesLimitsV(
				0,
				profileCapacity/profileSamplesPerSecond,
				0.0,
				float64(maxValue),
				implot.CondAlways,
			)
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
}

func (s *Sample) drawCountersTab(fontSize float32) {
	const (
		worldColumnWidthEm         = 14.0
		memoryColumnWidthEm        = 11.0
		bytesPerKiB                = 1024
		solverBarHeightEm          = 2.0
		solverTrackColor    uint32 = 0xff282828
		overflowColor       uint32 = 0xff3c3cdc
		redShift                   = 16
		greenShift                 = 8
	)

	if !imgui.BeginTabItem("Counters") {
		return
	}

	if !s.worldID.IsValid() {
		imgui.TextDisabled("No world")
		imgui.EndTabItem()

		return
	}

	c := s.worldID.Counters()
	capacity := s.worldID.MaxCapacity()
	colorCount := len(c.ColorCounts)
	overflowIndex := colorCount - 1

	var automaticTableSize imgui.Vec2

	const countersColumnCount = 3
	if !imgui.BeginTableV(
		"counters_layout",
		countersColumnCount,
		imgui.TableFlagsSizingFixedFit,
		automaticTableSize,
		0.0,
	) {
		imgui.EndTabItem()
		return
	}

	imgui.TableSetupColumnV(
		"world",
		imgui.TableColumnFlagsWidthFixed,
		worldColumnWidthEm*fontSize,
		0,
	)
	imgui.TableSetupColumnV(
		"memory",
		imgui.TableColumnFlagsWidthFixed,
		memoryColumnWidthEm*fontSize,
		0,
	)
	imgui.TableSetupColumnV("solver", imgui.TableColumnFlagsWidthStretch, 0.0, 0)
	imgui.TableNextRow()
	imgui.TableNextColumn()

	imTextf(
		"bodies   %d / %d",
		c.BodyCount,
		capacity.StaticBodyCount+capacity.DynamicBodyCount,
	)
	imTextf(
		"shapes   %d / %d",
		c.ShapeCount,
		capacity.StaticShapeCount+capacity.DynamicShapeCount,
	)
	imTextf("contacts %d / %d", c.ContactCount, capacity.ContactCount)
	imTextf("joints   %d", c.JointCount)
	imTextf("islands  %d", c.IslandCount)
	imTextf("tasks    %d", c.TaskCount)

	imgui.TableNextColumn()

	imTextf("static tree  %d", c.StaticTreeHeight)
	imTextf("movable tree %d", c.TreeHeight)
	imTextf("alloc %d K", c.ByteCount/bytesPerKiB)
	imTextf("stack %d K", c.StackUsed/bytesPerKiB)

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

	imTextf("%d constraints across %d colors", totalCount, colorCount-1)

	availWidth := imgui.ContentRegionAvail().X
	barHeight := solverBarHeightEm * fontSize
	dl := imgui.WindowDrawList()

	cursor := imgui.CursorScreenPos()
	dl.AddRectFilled(
		cursor,
		imgui.Vec2{X: cursor.X + availWidth, Y: cursor.Y + barHeight},
		solverTrackColor,
	)

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
			col := imCol32(uint8(hex>>redShift), uint8(hex>>greenShift), uint8(hex))
			dl.AddRectFilled(
				imgui.Vec2{X: x, Y: cursor.Y},
				imgui.Vec2{X: x + segW, Y: cursor.Y + barHeight},
				col,
			)
			x += segW
		}
	}

	imgui.Dummy(imgui.Vec2{X: availWidth, Y: barHeight})

	imgui.Spacing()

	var overflowFrac float32
	if totalCount > 0 {
		overflowFrac = float32(overflowCount) / float32(totalCount)
	}

	imgui.PushStyleColorU32(imgui.ColPlotHistogram, overflowColor)
	imgui.ProgressBarV(
		overflowFrac,
		imgui.Vec2{X: -fltMin, Y: 0.0},
		fmt.Sprintf("overflow %d", overflowCount),
	)
	imgui.PopStyleColor()

	var recycledFrac float32
	if c.AwakeContactCount > 0 {
		recycledFrac = clampFloat(
			float32(c.RecycledContactCount)/float32(c.AwakeContactCount),
			0.0,
			1.0,
		)
	}

	recycledOverlay := fmt.Sprintf(
		"recycled %d / %d",
		c.RecycledContactCount,
		c.AwakeContactCount,
	)
	imgui.ProgressBarV(
		recycledFrac,
		imgui.Vec2{X: -fltMin, Y: 0.0},
		recycledOverlay,
	)

	imgui.EndTable()

	imgui.EndTabItem()
}

func (s *Sample) init(context *SampleContext, self Sampler, createWorld bool) {
	s.self = self
	s.context = context
	s.camera = &context.camera
	s.draw = context.draw

	s.screenTextX = initialScreenTextX
	s.screenTextY = 0.0

	s.mouseForceScale = initialMouseForceScale

	scenes.SetRandomSeed(scenes.RandSeed)

	if createWorld {
		s.CreateWorld()
	}
}
func (s *Sample) base() *Sample { return s }

func summarizeProfile(
	histories *[profileRowCount][profileCapacity]float32,
	totals *[profileRowCount]float32,
	count int,
	now *[profileRowCount]float32,
	avg *[profileRowCount]float32,
	rowMax *[profileRowCount]float32,
) {
	const nowWindow = 10

	n := min(count, nowWindow)
	if n > 0 {
		inv := 1.0 / float32(n)

		for r := range profileRowCount {
			var sum float32
			for i := count - n; i < count; i++ {
				sum += histories[r][i]
			}

			now[r] = sum * inv
		}
	}

	if count > 0 {
		scale := 1.0 / float32(count)
		for i := range profileRowCount {
			avg[i] = scale * totals[i]
		}
	}

	for r := range profileRowCount {
		for i := range count {
			if histories[r][i] > rowMax[r] {
				rowMax[r] = histories[r][i]
			}
		}
	}
}

func (s *Sample) drawProfileTable(
	rows []rowDef,
	now []float32,
	avg []float32,
	rowMax []float32,
	stepNow float32,
	fontSize float32,
) {
	tableFlags := imgui.TableFlagsBordersInnerV | imgui.TableFlagsRowBg | imgui.TableFlagsSizingFixedFit

	var automaticSize imgui.Vec2
	if !imgui.BeginTableV("profile", profileColumnCount, tableFlags, automaticSize, 0.0) {
		return
	}

	imgui.TableSetupColumnV(
		"section",
		imgui.TableColumnFlagsWidthFixed,
		profileSectionWidth*fontSize,
		0,
	)
	imgui.TableSetupColumnV(
		"now",
		imgui.TableColumnFlagsWidthFixed,
		profileValueWidth*fontSize,
		0,
	)
	imgui.TableSetupColumnV(
		"avg",
		imgui.TableColumnFlagsWidthFixed,
		profileValueWidth*fontSize,
		0,
	)
	imgui.TableSetupColumnV(
		"max",
		imgui.TableColumnFlagsWidthFixed,
		profileValueWidth*fontSize,
		0,
	)
	imgui.TableSetupColumnV("%", imgui.TableColumnFlagsWidthFixed, profileBarWidth*fontSize, 0)
	imgui.TableHeadersRow()

	for r := range rows {
		d := &rows[r]

		imgui.TableNextRow()

		imgui.TableNextColumn()

		indent := float32(d.indent) * profileIndentWidth * fontSize
		imgui.SetCursorPosX(imgui.CursorPosX() + indent)
		imgui.PushStyleColorU32(imgui.ColText, d.color)
		imgui.TextUnformatted(d.name)
		imgui.PopStyleColor()

		const idleThresholdMS = 0.005

		idle := avg[r] < idleThresholdMS
		if idle {
			imgui.PushStyleColorVec4(
				imgui.ColText,
				*imgui.StyleColorVec4(imgui.ColTextDisabled),
			)
		}

		imgui.TableNextColumn()
		imTextf("%6.2f", now[r])
		imgui.TableNextColumn()
		imTextf("%6.2f", avg[r])
		imgui.TableNextColumn()
		imTextf("%6.2f", rowMax[r])

		if idle {
			imgui.PopStyleColor()
		}

		imgui.TableNextColumn()

		isBudget := r == 0 && s.context.hertz > 0.0

		reference := stepNow
		if isBudget {
			reference = millisecondsPerSecond / s.context.hertz
		}

		frac := clampFloat(now[r]/reference, 0.0, 1.0)
		barWidth := imgui.ContentRegionAvail().X
		barHeight := imgui.TextLineHeight()
		p := imgui.CursorScreenPos()
		drawList := imgui.WindowDrawList()
		drawList.AddRectFilled(p, imgui.Vec2{X: p.X + barWidth, Y: p.Y + barHeight}, colorTrack)
		drawList.AddRectFilled(
			p,
			imgui.Vec2{X: p.X + frac*barWidth, Y: p.Y + barHeight},
			d.color,
		)
		imgui.Dummy(imgui.Vec2{X: barWidth, Y: barHeight})

		if isBudget {
			imSetItemTooltipf("%.0f%% of %.1f ms frame", percentScale*now[r]/reference, reference)
		} else {
			imSetItemTooltipf("%.0f%% of step", percentScale*now[r]/reference)
		}
	}

	imgui.EndTable()
}

func fuzzyScore(needle string, haystack string) int {
	if needle == "" {
		return 0
	}

	score := 0
	hi := 0

	const noPreviousMatch = -2

	prevMatchHi := noPreviousMatch

	for ni := range len(needle) {
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

func lower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}

	return c
}

func isAlnum(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func drawRow(key string, desc string) {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.TextUnformatted(key)
	imgui.TableSetColumnIndex(1)
	imgui.TextUnformatted(desc)
}

func drawMenuBar(context *SampleContext) {
	fontSize := imgui.FontSize()

	if !imgui.BeginMainMenuBar() {
		return
	}

	drawSimulationMenu(context)
	drawViewMenu(context, fontSize)
	drawSamplesMenu(context)

	if replayIndex >= 0 && imgui.BeginMenu("Replay") {
		if imgui.MenuItemBool("Open...") {
			outPath, err := zenity.SelectFile(
				zenity.FileFilter{Name: "Box2D recording", Patterns: []string{"*.b2rec"}},
			)
			if err == nil {
				context.replayFile = outPath
				SelectSample(context, replayIndex, false)
			} else if !errors.Is(err, zenity.ErrCanceled) {
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
		imgui.ForegroundDrawListViewportPtr().
			AddLineArgs(imgui.Vec2{X: 0.0, Y: menuBarBottom}, imgui.Vec2{X: displaySize.X, Y: menuBarBottom},
				borderColor, 1.0)
	}

	drawHelpWindows(context, fontSize)
}

func drawHelpWindows(context *SampleContext, fontSize float32) {
	const (
		dialogCenter    = 0.5
		controlsWidthEm = 24.0
		aboutWidthEm    = 22.0
	)

	if context.showControls {
		imgui.SetNextWindowPosV(
			imgui.Vec2{
				X: context.camera.width * dialogCenter,
				Y: context.camera.height * dialogCenter,
			},
			imgui.CondAppearing,
			imgui.Vec2{X: dialogCenter, Y: dialogCenter},
		)
		imgui.SetNextWindowSizeV(
			imgui.Vec2{X: controlsWidthEm * fontSize, Y: 0.0},
			imgui.CondAppearing,
		)

		if imgui.BeginV(
			"Controls",
			&context.showControls,
			imgui.WindowFlagsNoCollapse|imgui.WindowFlagsAlwaysAutoResize,
		) {
			drawControlHelpTables()
		}

		imgui.End()
	}

	if showAbout {
		imgui.SetNextWindowPosV(
			imgui.Vec2{
				X: context.camera.width * dialogCenter,
				Y: context.camera.height * dialogCenter,
			},
			imgui.CondAppearing,
			imgui.Vec2{X: dialogCenter, Y: dialogCenter},
		)
		imgui.SetNextWindowSizeV(
			imgui.Vec2{X: aboutWidthEm * fontSize, Y: 0.0},
			imgui.CondAppearing,
		)

		if imgui.BeginV(
			"About",
			&showAbout,
			imgui.WindowFlagsNoCollapse|imgui.WindowFlagsAlwaysAutoResize,
		) {
			version := box2d.CurrentVersion()
			imTextf("Box2D %d.%d.%d", version.Major, version.Minor, version.Revision)
			imgui.Spacing()
			imgui.TextLinkOpenURLV("box2d.org", "https://box2d.org/")
			imgui.TextLinkOpenURLV(
				"github.com/erincatto/box2d",
				"https://github.com/erincatto/box2d",
			)
		}

		imgui.End()
	}
}

func drawControlHelpTables() {
	var automaticTableSize imgui.Vec2

	imgui.SeparatorText("Keyboard")

	const helpTableColumns = 2
	if imgui.BeginTableV(
		"keys",
		helpTableColumns,
		imgui.TableFlagsSizingFixedFit,
		automaticTableSize,
		0.0,
	) {
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

	if imgui.BeginTableV(
		"mouse",
		helpTableColumns,
		imgui.TableFlagsSizingFixedFit,
		automaticTableSize,
		0.0,
	) {
		drawRow("Left drag", "Move bodies (mouse joint)")
		drawRow("Right drag", "Pan camera")
		drawRow("Scroll wheel", "Zoom")
		imgui.EndTable()
	}
}

func drawSimulationMenu(context *SampleContext) {
	if !imgui.BeginMenu("Sim") {
		return
	}

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

func drawViewMenu(context *SampleContext, fontSize float32) {
	if !imgui.BeginMenu("View") {
		return
	}

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
	imgui.MenuItemBoolPtrV(
		"Profile",
		"I",
		&context.showProfile,
		context.sample.HasProfile(),
	)
	imgui.Separator()

	if imgui.BeginMenu("Scale") {
		const scaleInputWidthEm = 6.0
		imgui.PushItemWidth(scaleInputWidthEm * fontSize)
		imgui.InputFloat("Joint", &context.debugDraw.JointScale)
		imgui.InputFloat("Force", &context.debugDraw.ForceScale)
		imgui.PopItemWidth()
		imgui.EndMenu()
	}

	imgui.EndMenu()
}

func drawSamplesMenu(context *SampleContext) {
	if !imgui.BeginMenu("Samples") {
		return
	}

	for i := 0; i < len(sampleEntries); {
		i = drawSampleCategory(context, sampleEntries[i].category, i)
	}

	imgui.EndMenu()
}

func drawSampleCategory(context *SampleContext, category string, start int) int {
	open := imgui.BeginMenu(category)

	i := start
	for i < len(sampleEntries) && category == sampleEntries[i].category {
		if open {
			selected := i == context.sampleIndex
			if imgui.MenuItemBoolV(sampleEntries[i].name, "", selected, true) {
				SelectSample(context, i, false)
			}
		}

		i++
	}

	if open {
		imgui.EndMenu()
	}

	return i
}

func rebuild(q string) []int {
	var list []scored

	for i := range sampleEntries {
		nameScore := fuzzyScore(q, sampleEntries[i].name)
		catScore := fuzzyScore(q, sampleEntries[i].category)

		best := -1

		if nameScore >= 0 {
			const sampleNameScoreWeight = 2

			best = nameScore * sampleNameScoreWeight
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

func drawSamplePicker(context *SampleContext) {
	var automaticItemSize imgui.Vec2

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

	const (
		pickerHorizontalCenter = 0.5
		pickerVerticalPosition = 0.35
		pickerPivot            = 0.5
	)
	imgui.SetNextWindowPosV(
		imgui.Vec2{
			X: context.camera.width * pickerHorizontalCenter,
			Y: context.camera.height * pickerVerticalPosition,
		},
		imgui.CondAppearing,
		imgui.Vec2{X: pickerPivot, Y: pickerPivot},
	)

	const pickerWidthEm = 32.0
	imgui.SetNextWindowSizeV(imgui.Vec2{X: pickerWidthEm * fontSize, Y: 0.0}, imgui.CondAppearing)

	if !imgui.BeginPopupV(
		"##sample_picker",
		imgui.WindowFlagsAlwaysAutoResize|imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoSavedSettings,
	) {
		return
	}

	filteredCount := updatePickerQuery()

	commit := imgui.IsKeyPressedBoolV(imgui.KeyEnter, false) ||
		imgui.IsKeyPressedBoolV(imgui.KeyKeypadEnter, false)

	const pickerResultsHeightEm = 14.0
	imgui.BeginChildStrV(
		"##results",
		imgui.Vec2{X: 0.0, Y: pickerResultsHeightEm * fontSize},
		imgui.ChildFlagsBorders,
		0,
	)

	for row := range filteredCount {
		i := picker.filtered[row]
		label := fmt.Sprintf("%s  >  %s", sampleEntries[i].category, sampleEntries[i].name)

		sel := row == picker.highlight
		if imgui.SelectableBoolV(label, sel, 0, automaticItemSize) {
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

func updatePickerQuery() int {
	if picker.justOpened {
		imgui.SetKeyboardFocusHere()

		picker.justOpened = false
	}

	imgui.PushItemWidth(-1.0)
	imgui.InputTextWithHint("##q", "Search by name or category...", &picker.query, 0, nil)
	imgui.PopItemWidth()

	const maxPickerQueryLength = 63
	if len(picker.query) > maxPickerQueryLength {
		picker.query = picker.query[:maxPickerQueryLength]
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

	return filteredCount
}

func helpMarker(text string) {
	imgui.SameLineV(0.0, imgui.CurrentStyle().ItemInnerSpacing().X)

	width := imgui.FontSize()
	height := imgui.FrameHeight()
	pos := imgui.CursorScreenPos()

	imgui.Dummy(imgui.Vec2{X: width, Y: height})

	hovered := imgui.IsItemHovered()

	const markerIdleAlpha = 0.6

	alpha := float32(markerIdleAlpha)
	if hovered {
		alpha = 1.0
	}

	color := imgui.ColorU32ColV(imgui.ColCheckMark, alpha)

	const (
		markerCenter = 0.5
		markerRadius = 0.2
	)

	center := imgui.Vec2{X: pos.X + markerCenter*width, Y: pos.Y + markerCenter*height}
	imgui.WindowDrawList().AddCircleFilled(center, markerRadius*width, color)
	imSetItemTooltipf("%s", text)
}

func drawInfoPanel(context *SampleContext, frameTime float32) {
	const (
		controlWidthEm       = 6.0
		maxSubSteps          = 32
		maxBounceIterations  = 8
		minHertz             = 5.0
		maxHertz             = 240.0
		centimetersPerMeter  = 100.0
		metersPerCentimeter  = 0.01
		maxRecycleDistanceCM = 10.0
	)

	entry := &sampleEntries[context.sampleIndex]
	sample := context.sample.base()
	fontSize := imgui.FontSize()
	menuWidth := context.sample.InfoPanelWidthEm() * fontSize
	menuBarHeight := imgui.FrameHeight()

	imgui.SetNextWindowPos(
		imgui.Vec2{
			X: context.camera.width - menuWidth - panelMarginEm*fontSize,
			Y: menuBarHeight + panelMarginEm*fontSize,
		},
	)
	imgui.SetNextWindowSize(
		imgui.Vec2{X: menuWidth, Y: context.camera.height - menuBarHeight - fontSize},
	)

	imgui.BeginV(
		"Info",
		nil,
		imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize|imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoTitleBar,
	)

	imTextColoredf(MakeColor(box2d.ColorGoldenRod), "%s", entry.name)
	imTextColoredf(MakeColor(box2d.ColorLightGray), "%s", entry.category)
	imgui.Separator()

	if context.pause {
		imTextColoredf(MakeColor(box2d.ColorRed), "PAUSED")
		imgui.SameLine()
		imgui.TextDisabled("(P)")
		imgui.Separator()
	}

	imTextColoredf(MakeColor(box2d.ColorSeaGreen), "frame %.1f ms", millisecondsPerSecond*frameTime)
	imTextColoredf(MakeColor(box2d.ColorSeaGreen), "step %d", sample.stepCount)
	imgui.Separator()
	imTextColoredf(
		MakeColor(box2d.ColorSeaGreen),
		"cam (%.1f, %.1f)",
		context.camera.center.X,
		context.camera.center.Y,
	)
	imTextColoredf(MakeColor(box2d.ColorSeaGreen), "zoom %.2f", context.camera.zoom)

	imgui.Separator()

	imgui.PushItemWidth(controlWidthEm * fontSize)

	if context.sample.DrawControls() {
		imgui.Separator()
	}

	imgui.PopItemWidth()

	if context.sample.HasSolverControls() &&
		imgui.CollapsingHeaderTreeNodeFlagsV("Solver", imgui.TreeNodeFlagsDefaultOpen) {
		imgui.PushItemWidth(controlWidthEm * fontSize)

		imgui.SliderInt("Sub-steps##Solver", &context.subStepCount, 1, maxSubSteps)
		helpMarker(
			"The solver breaks the full step into several sub-steps.\n" +
				"More sub-steps usually lead to more accurate results.",
		)

		imgui.SliderInt(
			"Bounce Iters##Solver",
			&context.restitutionIterations,
			0,
			maxBounceIterations,
		)
		helpMarker("Iterations for the restitution solver.")

		imgui.SliderFloatV("Hertz##Solver", &context.hertz, minHertz, maxHertz, "%.0f Hz", 0)
		helpMarker("The number of world steps per second.")

		if imgui.SliderInt("Workers##Solver", &context.workerCount, 1, box2d.MaxWorkers) {
			context.workerCount = min(max(context.workerCount, 1), box2d.MaxWorkers)
			SelectSample(context, context.sampleIndex, true)
			sample = context.sample.base()
		}

		helpMarker("The number worker threads used by the world step.")

		recyclingCentimeters := centimetersPerMeter * context.recycleDistance
		if imgui.SliderFloatV(
			"Recycle##Solver",
			&recyclingCentimeters,
			0.0,
			maxRecycleDistanceCM,
			"%.1f cm",
			0,
		) {
			context.recycleDistance = metersPerCentimeter * recyclingCentimeters
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

	drawRecordingPanel(context, sample, fontSize)

	imgui.End()
}

func drawRecordingPanel(context *SampleContext, sample *Sample, fontSize float32) {
	if !context.sample.HasSolverControls() ||
		!imgui.CollapsingHeaderTreeNodeFlagsV("Recording", imgui.TreeNodeFlagsDefaultOpen) {
		return
	}

	const recordingPathWidthEm = 9.0
	imgui.PushItemWidth(recordingPathWidthEm * fontSize)
	imgui.InputTextWithHint("File##Recording", "", &context.recordingFile, 0, nil)
	imgui.PopItemWidth()

	const maxRecordingPathLength = 255
	if len(context.recordingFile) > maxRecordingPathLength {
		context.recordingFile = context.recordingFile[:maxRecordingPathLength]
	}

	sample = drawRecordingButtons(context, sample)

	if sample.recording != nil {
		const bytesPerKiB = 1024.0

		kilobytes := float32(sample.recording.Size()) / bytesPerKiB

		steps := sample.stepCount - sample.recordStartStep
		if kilobytes < bytesPerKiB {
			imTextColoredf(
				MakeColor(box2d.ColorRed),
				"recording %d steps, %.0f KB",
				steps,
				kilobytes,
			)
		} else {
			imTextColoredf(
				MakeColor(box2d.ColorRed),
				"recording %d steps, %.1f MB",
				steps,
				kilobytes/bytesPerKiB,
			)
		}
	}
}

func drawRecordingButtons(context *SampleContext, sample *Sample) *Sample {
	if sample.recording != nil {
		if imgui.Button("Stop##Recording") {
			sample.FinishRecording()
		}

		return sample
	}

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

		imSetItemTooltipf("Open %s in the replay viewer", context.savedRecordingFile)
	}

	return sample
}

func imCol32(r uint8, g uint8, b uint8) uint32 {
	const opaqueAlpha uint32 = 0xff
	return opaqueAlpha<<24 | uint32(b)<<16 | uint32(g)<<8 | uint32(r)
}

func clampFloat(a float32, lower float32, upper float32) float32 {
	return min(max(a, lower), upper)
}

func escapeFormat(format string, args []any) string {
	return strings.ReplaceAll(fmt.Sprintf(format, args...), "%", "%%")
}

func imTextf(format string, args ...any) {
	imgui.Text(escapeFormat(format, args))
}

func imTextColoredf(color imgui.Vec4, format string, args ...any) {
	imgui.TextColored(color, escapeFormat(format, args))
}

func imTextDisabledf(format string, args ...any) {
	imgui.TextDisabled(escapeFormat(format, args))
}

func imTextWrappedf(format string, args ...any) {
	imgui.TextWrapped(escapeFormat(format, args))
}

func imSetItemTooltipf(format string, args ...any) {
	imgui.SetItemTooltip(escapeFormat(format, args))
}
