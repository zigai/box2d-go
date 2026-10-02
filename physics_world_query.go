// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type OverlapResultFunc func(shapeID Shape) bool
type CastResultFunc func(shapeID Shape, point Pos, normal Vec2, fraction float32) float32
type PlaneResultFunc func(shapeID Shape, plane *PlaneResult) bool

type RayResult struct {
	ShapeID    Shape
	Point      Pos
	Normal     Vec2
	Fraction   float32
	NodeVisits int
	LeafVisits int
	Hit        bool
}

type ExplosionDef struct {
	MaskBits         uint64
	Position         Pos
	Radius           float32
	Falloff          float32
	ImpulsePerLength float32
}

type HexColor int

const (
	ColorAliceBlue            HexColor = 0xF0F8FF
	ColorAntiqueWhite         HexColor = 0xFAEBD7
	ColorAqua                 HexColor = 0x00FFFF
	ColorAquamarine           HexColor = 0x7FFFD4
	ColorAzure                HexColor = 0xF0FFFF
	ColorBeige                HexColor = 0xF5F5DC
	ColorBisque               HexColor = 0xFFE4C4
	ColorBlack                HexColor = 0x000000
	ColorBlanchedAlmond       HexColor = 0xFFEBCD
	ColorBlue                 HexColor = 0x0000FF
	ColorBlueViolet           HexColor = 0x8A2BE2
	ColorBrown                HexColor = 0xA52A2A
	ColorBurlywood            HexColor = 0xDEB887
	ColorCadetBlue            HexColor = 0x5F9EA0
	ColorChartreuse           HexColor = 0x7FFF00
	ColorChocolate            HexColor = 0xD2691E
	ColorCoral                HexColor = 0xFF7F50
	ColorCornflowerBlue       HexColor = 0x6495ED
	ColorCornsilk             HexColor = 0xFFF8DC
	ColorCrimson              HexColor = 0xDC143C
	ColorCyan                 HexColor = 0x00FFFF
	ColorDarkBlue             HexColor = 0x00008B
	ColorDarkCyan             HexColor = 0x008B8B
	ColorDarkGoldenRod        HexColor = 0xB8860B
	ColorDarkGray             HexColor = 0xA9A9A9
	ColorDarkGreen            HexColor = 0x006400
	ColorDarkKhaki            HexColor = 0xBDB76B
	ColorDarkMagenta          HexColor = 0x8B008B
	ColorDarkOliveGreen       HexColor = 0x556B2F
	ColorDarkOrange           HexColor = 0xFF8C00
	ColorDarkOrchid           HexColor = 0x9932CC
	ColorDarkRed              HexColor = 0x8B0000
	ColorDarkSalmon           HexColor = 0xE9967A
	ColorDarkSeaGreen         HexColor = 0x8FBC8F
	ColorDarkSlateBlue        HexColor = 0x483D8B
	ColorDarkSlateGray        HexColor = 0x2F4F4F
	ColorDarkTurquoise        HexColor = 0x00CED1
	ColorDarkViolet           HexColor = 0x9400D3
	ColorDeepPink             HexColor = 0xFF1493
	ColorDeepSkyBlue          HexColor = 0x00BFFF
	ColorDimGray              HexColor = 0x696969
	ColorDodgerBlue           HexColor = 0x1E90FF
	ColorFireBrick            HexColor = 0xB22222
	ColorFloralWhite          HexColor = 0xFFFAF0
	ColorForestGreen          HexColor = 0x228B22
	ColorFuchsia              HexColor = 0xFF00FF
	ColorGainsboro            HexColor = 0xDCDCDC
	ColorGhostWhite           HexColor = 0xF8F8FF
	ColorGold                 HexColor = 0xFFD700
	ColorGoldenRod            HexColor = 0xDAA520
	ColorGray                 HexColor = 0x808080
	ColorGreen                HexColor = 0x008000
	ColorGreenYellow          HexColor = 0xADFF2F
	ColorHoneyDew             HexColor = 0xF0FFF0
	ColorHotPink              HexColor = 0xFF69B4
	ColorIndianRed            HexColor = 0xCD5C5C
	ColorIndigo               HexColor = 0x4B0082
	ColorIvory                HexColor = 0xFFFFF0
	ColorKhaki                HexColor = 0xF0E68C
	ColorLavender             HexColor = 0xE6E6FA
	ColorLavenderBlush        HexColor = 0xFFF0F5
	ColorLawnGreen            HexColor = 0x7CFC00
	ColorLemonChiffon         HexColor = 0xFFFACD
	ColorLightBlue            HexColor = 0xADD8E6
	ColorLightCoral           HexColor = 0xF08080
	ColorLightCyan            HexColor = 0xE0FFFF
	ColorLightGoldenRodYellow HexColor = 0xFAFAD2
	ColorLightGray            HexColor = 0xD3D3D3
	ColorLightGreen           HexColor = 0x90EE90
	ColorLightPink            HexColor = 0xFFB6C1
	ColorLightSalmon          HexColor = 0xFFA07A
	ColorLightSeaGreen        HexColor = 0x20B2AA
	ColorLightSkyBlue         HexColor = 0x87CEFA
	ColorLightSlateGray       HexColor = 0x778899
	ColorLightSteelBlue       HexColor = 0xB0C4DE
	ColorLightYellow          HexColor = 0xFFFFE0
	ColorLime                 HexColor = 0x00FF00
	ColorLimeGreen            HexColor = 0x32CD32
	ColorLinen                HexColor = 0xFAF0E6
	ColorMagenta              HexColor = 0xFF00FF
	ColorMaroon               HexColor = 0x800000
	ColorMediumAquaMarine     HexColor = 0x66CDAA
	ColorMediumBlue           HexColor = 0x0000CD
	ColorMediumOrchid         HexColor = 0xBA55D3
	ColorMediumPurple         HexColor = 0x9370DB
	ColorMediumSeaGreen       HexColor = 0x3CB371
	ColorMediumSlateBlue      HexColor = 0x7B68EE
	ColorMediumSpringGreen    HexColor = 0x00FA9A
	ColorMediumTurquoise      HexColor = 0x48D1CC
	ColorMediumVioletRed      HexColor = 0xC71585
	ColorMidnightBlue         HexColor = 0x191970
	ColorMintCream            HexColor = 0xF5FFFA
	ColorMistyRose            HexColor = 0xFFE4E1
	ColorMoccasin             HexColor = 0xFFE4B5
	ColorNavajoWhite          HexColor = 0xFFDEAD
	ColorNavy                 HexColor = 0x000080
	ColorOldLace              HexColor = 0xFDF5E6
	ColorOlive                HexColor = 0x808000
	ColorOliveDrab            HexColor = 0x6B8E23
	ColorOrange               HexColor = 0xFFA500
	ColorOrangeRed            HexColor = 0xFF4500
	ColorOrchid               HexColor = 0xDA70D6
	ColorPaleGoldenRod        HexColor = 0xEEE8AA
	ColorPaleGreen            HexColor = 0x98FB98
	ColorPaleTurquoise        HexColor = 0xAFEEEE
	ColorPaleVioletRed        HexColor = 0xDB7093
	ColorPapayaWhip           HexColor = 0xFFEFD5
	ColorPeachPuff            HexColor = 0xFFDAB9
	ColorPeru                 HexColor = 0xCD853F
	ColorPink                 HexColor = 0xFFC0CB
	ColorPlum                 HexColor = 0xDDA0DD
	ColorPowderBlue           HexColor = 0xB0E0E6
	ColorPurple               HexColor = 0x800080
	ColorRebeccaPurple        HexColor = 0x663399
	ColorRed                  HexColor = 0xFF0000
	ColorRosyBrown            HexColor = 0xBC8F8F
	ColorRoyalBlue            HexColor = 0x4169E1
	ColorSaddleBrown          HexColor = 0x8B4513
	ColorSalmon               HexColor = 0xFA8072
	ColorSandyBrown           HexColor = 0xF4A460
	ColorSeaGreen             HexColor = 0x2E8B57
	ColorSeaShell             HexColor = 0xFFF5EE
	ColorSienna               HexColor = 0xA0522D
	ColorSilver               HexColor = 0xC0C0C0
	ColorSkyBlue              HexColor = 0x87CEEB
	ColorSlateBlue            HexColor = 0x6A5ACD
	ColorSlateGray            HexColor = 0x708090
	ColorSnow                 HexColor = 0xFFFAFA
	ColorSpringGreen          HexColor = 0x00FF7F
	ColorSteelBlue            HexColor = 0x4682B4
	ColorTan                  HexColor = 0xD2B48C
	ColorTeal                 HexColor = 0x008080
	ColorThistle              HexColor = 0xD8BFD8
	ColorTomato               HexColor = 0xFF6347
	ColorTurquoise            HexColor = 0x40E0D0
	ColorViolet               HexColor = 0xEE82EE
	ColorWheat                HexColor = 0xF5DEB3
	ColorWhite                HexColor = 0xFFFFFF
	ColorWhiteSmoke           HexColor = 0xF5F5F5
	ColorYellow               HexColor = 0xFFFF00
	ColorYellowGreen          HexColor = 0x9ACD32
	ColorBox2DRed             HexColor = 0xDC3132
	ColorBox2DBlue            HexColor = 0x30AEBF
	ColorBox2DGreen           HexColor = 0x8CC924
	ColorBox2DYellow          HexColor = 0xFFEE8C
)

type DebugDraw struct {
	DrawPolygonFunc      func(transform WorldTransform, vertices []Vec2, color HexColor)
	DrawSolidPolygonFunc func(transform WorldTransform, vertices []Vec2, radius float32, color HexColor)
	DrawCircleFunc       func(center Pos, radius float32, color HexColor)
	DrawSolidCircleFunc  func(transform WorldTransform, center Vec2, radius float32, color HexColor)
	DrawSolidCapsuleFunc func(p1, p2 Pos, radius float32, color HexColor)
	DrawLineFunc         func(p1, p2 Pos, color HexColor)
	DrawTransformFunc    func(transform WorldTransform)
	DrawPointFunc        func(p Pos, size float32, color HexColor)
	DrawStringFunc       func(p Pos, s string, color HexColor)
	DrawBoundsFunc       func(aabb AABB, color HexColor)
	DrawingBounds        AABB
	ForceScale           float32
	JointScale           float32
	DrawContacts         bool
	DrawAnchorA          bool
	DrawShapes           bool
	DrawChainNormals     bool
	DrawJoints           bool
	DrawJointExtras      bool
	DrawBounds           bool
	DrawMass             bool
	DrawBodyNames        bool
	DrawGraphColors      bool
	DrawContactFeatures  bool
	DrawContactNormals   bool
	DrawContactForces    bool
	DrawFrictionForces   bool
	DrawIslands          bool
}

type drawContext struct {
	world *world
	draw  *DebugDraw
}

type worldQueryContext struct {
	world  *world
	fcn    OverlapResultFunc
	filter QueryFilter
}

type worldOverlapContext struct {
	world  *world
	fcn    OverlapResultFunc
	filter QueryFilter
	proxy  *ShapeProxy
	origin Pos
}

type worldRayCastContext struct {
	world    *world
	fcn      CastResultFunc
	filter   QueryFilter
	fraction float32
	origin   Pos
}

type worldShapeCastContext struct {
	world    *world
	fcn      CastResultFunc
	filter   QueryFilter
	fraction float32
	origin   Pos
	input    ShapeCastInput
}

type worldMoverCastContext struct {
	world    *world
	filter   QueryFilter
	fraction float32
	origin   Pos
	input    ShapeCastInput
}

type worldMoverContext struct {
	world  *world
	fcn    PlaneResultFunc
	filter QueryFilter
	mover  Capsule
	origin Pos
}

type explosionContext struct {
	world            *world
	position         Pos
	radius           float32
	falloff          float32
	impulsePerLength float32
}

func drawShape(draw *DebugDraw, s *shape, transform WorldTransform, color HexColor, drawChainNormals bool) {
	todo()
}

func drawQueryCallback(proxyID int, userData uint64, context *drawContext) bool {
	todo()
	return false
}

func (w World) Draw(draw *DebugDraw) {
	todo()
}

func treeQueryCallback(proxyID int, userData uint64, context *worldQueryContext) bool {
	todo()
	return false
}

func (w World) OverlapAABB(origin Pos, aabb AABB, filter QueryFilter, fcn OverlapResultFunc) TreeStats {
	todo()
	return TreeStats{}
}

func treeOverlapCallback(proxyID int, userData uint64, context *worldOverlapContext) bool {
	todo()
	return false
}

func (w World) OverlapShape(origin Pos, proxy *ShapeProxy, filter QueryFilter, fcn OverlapResultFunc) TreeStats {
	todo()
	return TreeStats{}
}

func rayCastCallback(input *RayCastInput, proxyID int, userData uint64, context *worldRayCastContext) float32 {
	todo()
	return 0
}

func (w World) CastRay(origin Pos, translation Vec2, filter QueryFilter, fcn CastResultFunc) TreeStats {
	todo()
	return TreeStats{}
}

func rayCastClosestFcn(shapeID Shape, point Pos, normal Vec2, fraction float32, result *RayResult) float32 {
	todo()
	return 0
}

func (w World) CastRayClosest(origin Pos, translation Vec2, filter QueryFilter) RayResult {
	todo()
	return RayResult{}
}

func shapeCastCallback(input *BoxCastInput, proxyID int, userData uint64, context *worldShapeCastContext) float32 {
	todo()
	return 0
}

func (w World) CastShape(origin Pos, proxy *ShapeProxy, translation Vec2, filter QueryFilter, fcn CastResultFunc) TreeStats {
	todo()
	return TreeStats{}
}

func moverCastCallback(input *BoxCastInput, proxyID int, userData uint64, context *worldMoverCastContext) float32 {
	todo()
	return 0
}

func (w World) CastMover(origin Pos, mover *Capsule, translation Vec2, filter QueryFilter) float32 {
	todo()
	return 0
}

func treeCollideCallback(proxyID int, userData uint64, context *worldMoverContext) bool {
	todo()
	return false
}

func (w World) CollideMover(origin Pos, mover *Capsule, filter QueryFilter, fcn PlaneResultFunc) {
	todo()
}

func explosionCallback(proxyID int, userData uint64, context *explosionContext) bool {
	todo()
	return false
}

func (w World) Explode(def *ExplosionDef) {
	todo()
}

func validateConnectivity(w *world) {
	todo()
}

func validateSolverSets(w *world) {
	todo()
}

func validateContacts(w *world) {
	todo()
}
