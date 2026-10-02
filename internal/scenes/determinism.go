// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"unsafe"

	"github.com/zigai/box2d"
)

const (
	fallingHingePrecision        = int(unsafe.Sizeof(box2d.Pos{}))/8 - 1
	ExpectedSleepStep            = 295*(1-fallingHingePrecision) + 256*fallingHingePrecision
	ExpectedHash          uint32 = uint32(0x36564C94*(1-fallingHingePrecision) + 0x29D0B92D*fallingHingePrecision)
)

type FallingHingeData struct {
	BodyIDs   []box2d.Body
	BodyCount int
	StepCount int
	SleepStep int
	Hash      uint32
}

func CreateFallingHinges(worldID box2d.World) FallingHingeData {
	todo()
	return FallingHingeData{}
}

func UpdateFallingHinges(worldID box2d.World, data *FallingHingeData) bool {
	todo()
	return false
}

func DestroyFallingHinges(data *FallingHingeData) {
	todo()
}
