package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/zigai/box2d-go/box2d"
)

func TestSampleRefusedRecordingPreservesSavedFile(t *testing.T) {
	worldDef := box2d.DefaultWorldDef()
	worldDef.WorkerCount = 0
	world := box2d.NewWorld(&worldDef)
	t.Cleanup(world.Close)

	previous := box2d.NewRecording(0)
	world.StartRecording(previous)
	world.Step(1.0/60, 1)
	world.StopRecording()

	data := bytes.Clone(previous.Data())
	path := filepath.Join(t.TempDir(), "recording.b2rec")

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	world.SetCustomFilterCallback(func(box2d.Shape, box2d.Shape) bool { return false })

	context := &SampleContext{recordingFile: path}
	sample := Sample{worldID: world, context: context}
	sample.StartRecording()
	sample.FinishRecording()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, data) {
		t.Errorf("refused recording changed saved file: size = %d, want %d", len(got), len(data))
	}
}
