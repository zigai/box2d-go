// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

type ownershipScene struct {
	w      *box2d.World
	bodies [12]box2d.Body
	shapes [14]box2d.Shape
	joint  box2d.Joint
}

func newOwnershipScene(t *testing.T, variant, workers int, enqueue box2d.EnqueueTaskFunc) ownershipScene {
	t.Helper()

	var s ownershipScene

	wd := box2d.DefaultWorldDef()
	wd.WorkerCount = workers
	wd.EnqueueTask = enqueue
	wd.EnableSleep = false

	wd.Gravity = box2d.Vec2{Y: -10}
	if variant == 1 {
		wd.Gravity = box2d.Vec2{X: 2, Y: -6}
	}

	wd.UserData = variant
	s.w = box2d.NewWorld(&wd)
	box2d.CleanupWorld(t, s.w)

	bd := box2d.DefaultBodyDef()
	bd.Position = box2d.Pos{Y: -0.5}
	floor := s.w.CreateBody(&bd)
	box := box2d.MakeBox(20, 0.5)
	sd := box2d.DefaultShapeDef()
	sd.EnableContactEvents = true
	sd.EnableSensorEvents = true
	s.shapes[0] = floor.CreatePolygonShape(&sd, &box)
	bd.Position = box2d.Pos{Y: 2}
	sensor := s.w.CreateBody(&bd)
	sd.IsSensor = true
	box = box2d.MakeBox(8, 0.5)
	s.shapes[1] = sensor.CreatePolygonShape(&sd, &box)
	sd.IsSensor = false
	box = box2d.MakeBox(0.4, 0.4)

	bd.Type = box2d.DynamicBody
	for i := range 12 {
		bd.Position = box2d.Vec2{X: float32(i%4-2) * 1.1, Y: 0.5 + float32(float32(i/4)*1.1)}.Pos()

		bd.LinearVelocity = box2d.Vec2{X: -0.2}
		if variant == 1 {
			bd.LinearVelocity.X = 0.3
		}

		bd.UserData = variant*100 + i
		s.bodies[i] = s.w.CreateBody(&bd)
		s.shapes[i+2] = s.bodies[i].CreatePolygonShape(&sd, &box)
	}

	s.joint = lifecycleJoint(s.w, s.bodies[8], s.bodies[9])

	return s
}

func observeOwnershipScene(r *oracle.Reader, s ownershipScene, step int) {
	r.U64(fmt.Sprintf("step%d.state", step), s.w.StateHash())
	g := s.w.Gravity()
	r.Vec2("gravity", g.X, g.Y)
	r.I32("workers", int32(s.w.WorkerCount()))

	for _, b := range s.bodies {
		p := b.Position()
		q := b.Rotation()
		v := b.LinearVelocity()

		r.Pos("position", p.X, p.Y)
		r.Rot("rotation", q.C, q.S)
		r.Vec2("velocity", v.X, v.Y)
		r.F32("angular", b.AngularVelocity())
		r.Bool("awake", b.Awake())
	}

	begins := slices.Collect(s.w.ContactBeginEvents())
	r.I32("begin.count", int32(len(begins)))

	for _, e := range begins {
		r.U64("begin.a", lifecycleShapeKey(e.ShapeA))
		r.U64("begin.b", lifecycleShapeKey(e.ShapeB))
		id := e.Contact.ID()
		r.U32("begin.contact.index", uint32(id>>32))
		r.U32("begin.contact.generation", uint32(id))

		valid := e.Contact.IsValid()
		r.Bool("begin.contact.valid", valid)

		if valid {
			data := e.Contact.Data()
			r.Vec2("begin.normal", data.Manifold.Normal.X, data.Manifold.Normal.Y)
			r.I32("begin.points", int32(data.Manifold.PointCount))
		}
	}

	ends := slices.Collect(s.w.ContactEndEvents())
	r.I32("end.count", int32(len(ends)))

	for _, e := range ends {
		r.U64("end.a", lifecycleShapeKey(e.ShapeA))
		r.U64("end.b", lifecycleShapeKey(e.ShapeB))
		id := e.Contact.ID()
		r.U32("end.contact.index", uint32(id>>32))
		r.U32("end.contact.generation", uint32(id))
		r.Bool("end.contact.valid", e.Contact.IsValid())
	}

	hits := slices.Collect(s.w.ContactHitEvents())
	r.I32("hit.count", int32(len(hits)))

	for _, e := range hits {
		r.U64("hit.a", lifecycleShapeKey(e.ShapeA))
		r.U64("hit.b", lifecycleShapeKey(e.ShapeB))
		id := e.Contact.ID()
		r.U32("hit.contact.index", uint32(id>>32))
		r.U32("hit.contact.generation", uint32(id))
		r.Pos("hit.point", e.Point.X, e.Point.Y)
		r.Vec2("hit.normal", e.Normal.X, e.Normal.Y)
		r.F32("hit.speed", e.ApproachSpeed)
	}

	sb := slices.Collect(s.w.SensorBeginEvents())
	r.I32("sensor.begin.count", int32(len(sb)))

	for _, e := range sb {
		r.U64("sensor.begin.sensor", lifecycleShapeKey(e.SensorShape))
		r.U64("sensor.begin.visitor", lifecycleShapeKey(e.VisitorShape))
	}

	se := slices.Collect(s.w.SensorEndEvents())
	r.I32("sensor.end.count", int32(len(se)))

	for _, e := range se {
		r.U64("sensor.end.sensor", lifecycleShapeKey(e.SensorShape))
		r.U64("sensor.end.visitor", lifecycleShapeKey(e.VisitorShape))
	}

	moves := slices.Collect(s.w.BodyMoveEvents())
	r.I32("move.count", int32(len(moves)))

	for _, e := range moves {
		r.U64("move.body", lifecycleBodyKey(e.Body))
		r.Pos("move.position", e.Transform.P.X, e.Transform.P.Y)
		r.Rot("move.rotation", e.Transform.Q.C, e.Transform.Q.S)
		r.Bool("move.sleep", e.FellAsleep)
	}
}

var ownershipWorkerCounts = []int{1, 4, 8, box2d.MaxWorkers, 0, -5, 4, 4}

func TestOwnershipWorkerReconfiguration(t *testing.T) {
	for variant := range 2 {
		for _, mode := range []string{"Serial", "Reconfigure"} {
			t.Run(fmt.Sprintf("Scene%d/%s", variant, mode), func(t *testing.T) {
				r := oracle.Open(t, fmt.Sprintf("ownership/Scene%d-%s", variant, mode))
				s := newOwnershipScene(t, variant, 1, nil)

				for step := range 48 {
					if mode == "Reconfigure" && step%6 == 0 {
						s.w.SetWorkerCount(ownershipWorkerCounts[step/6])
					}

					s.w.Step(1.0/60, 4)
					observeOwnershipScene(r, s, step)
				}

				s.w.Close()
			})
		}
	}
}

func ownershipWait(t *testing.T, ch <-chan struct{}, description string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

func TestOwnershipSynchronousNilFinish(t *testing.T) {
	r := oracle.Open(t, "ownership/Scene0-Serial")

	var enqueued, completed atomic.Int64

	enqueue := func(task func()) func() { enqueued.Add(1); task(); completed.Add(1); return nil }

	s := newOwnershipScene(t, 0, 1, enqueue)
	for step := range 48 {
		s.w.Step(1.0/60, 4)

		if enqueued.Load() != completed.Load() {
			t.Fatalf("step %d returned with incomplete synchronous work", step)
		}

		observeOwnershipScene(r, s, step)
	}

	if enqueued.Load() == 0 {
		t.Fatal("public Step never scheduled caller work")
	}

	s.w.Close()

	if enqueued.Load() != completed.Load() {
		t.Fatal("destroy left owned work incomplete")
	}
}

func TestOwnershipAsyncStepWaitsForExplicitRelease(t *testing.T) {
	r := oracle.Open(t, "ownership/Scene0-Serial")
	release := make(chan struct{})
	started := make(chan struct{})
	finishing := make(chan struct{})
	returned := make(chan struct{})

	var releaseOnce sync.Once

	work := ownershipHeldTasks{release: release, started: started, finishing: finishing}
	s := newOwnershipScene(t, 0, 1, work.enqueue)
	// Release and join before the registered world cleanup, including failure paths.
	t.Cleanup(func() {
		releaseOnce.Do(func() { close(release) })
		ownershipWait(t, returned, "blocked Step cleanup")
		work.tasks.Wait()
	})

	go func() { s.w.Step(1.0/60, 4); close(returned) }()

	ownershipWait(t, started, "first asynchronous task entry")

	select {
	case <-finishing:
	case <-returned:
		t.Fatal("Step returned without waiting for unfinished task")
	case <-time.After(10 * time.Second):
		t.Fatal("Step never invoked task finish")
	}

	select {
	case <-returned:
		t.Fatal("Step returned before explicit task release")
	default:
	}

	if work.heldCompleted.Load() {
		t.Fatal("held first task completed before release")
	}

	releaseOnce.Do(func() { close(release) })
	ownershipWait(t, returned, "Step after task release")
	work.tasks.Wait()

	if work.enqueued.Load() != work.completed.Load() || work.enqueued.Load() != work.finished.Load() {
		t.Fatalf("task completion: enqueued=%d completed=%d finished=%d", work.enqueued.Load(), work.completed.Load(), work.finished.Load())
	}

	observeOwnershipScene(r, s, 0)

	for step := 1; step < 48; step++ {
		s.w.Step(1.0/60, 4)
		work.tasks.Wait()

		if work.enqueued.Load() != work.completed.Load() || work.enqueued.Load() != work.finished.Load() {
			t.Fatalf("step %d incomplete or unowned work", step)
		}

		observeOwnershipScene(r, s, step)
	}

	s.w.Close()
	work.tasks.Wait()
}

func TestOwnershipRepeatedShutdownWithCompletedWork(t *testing.T) {
	for run := range 4 {
		t.Run(strconv.Itoa(run), func(t *testing.T) {
			r := oracle.Open(t, "ownership/Scene1-Reconfigure")

			var (
				enqueued, completed atomic.Int64
				tasks               sync.WaitGroup
			)

			enqueue := func(task func()) func() {
				enqueued.Add(1)

				done := make(chan struct{})

				tasks.Go(func() { ; task(); completed.Add(1); close(done) })

				return func() { <-done }
			}
			s := newOwnershipScene(t, 1, 4, enqueue)
			s.w.SetWorkerCount(1)

			for step := range 48 {
				if step%6 == 0 {
					tasks.Wait()

					ownershipCheckCompletedTasks(t, &enqueued, &completed, "reconfiguration with incomplete work")

					s.w.SetWorkerCount(ownershipWorkerCounts[step/6])
				}

				s.w.Step(1.0/60, 4)
				tasks.Wait()

				ownershipCheckCompletedTasks(t, &enqueued, &completed, "Step returned with incomplete owned work")

				observeOwnershipScene(r, s, step)
			}

			s.w.Close()
			tasks.Wait()

			if enqueued.Load() == 0 || enqueued.Load() != completed.Load() {
				t.Fatalf("destroy completion: enqueued=%d completed=%d", enqueued.Load(), completed.Load())
			}
		})
	}
}

func TestOwnershipConcurrentWorldsAgainstIndependentC(t *testing.T) {
	var (
		scenes    [2]ownershipScene
		readers   [2]*oracle.Reader
		requests  [2]chan struct{}
		done      [2]chan struct{}
		joined    [2]chan struct{}
		callbacks [2]atomic.Int64
		arrived   [2]chan struct{}
		gate      [2]chan struct{}
		first     [2]atomic.Bool
		release   func()
	)

	for variant := range 2 {
		enqueue := func(task func()) func() {
			callbacks[variant].Add(1)

			if first[variant].CompareAndSwap(false, true) {
				arrived[variant] <- struct{}{}

				<-gate[variant]
			}

			task()

			return nil
		}
		scenes[variant] = newOwnershipScene(t, variant, 1, enqueue)
		readers[variant] = oracle.Open(t, fmt.Sprintf("ownership/Scene%d-Serial", variant))
		requests[variant] = make(chan struct{})
		done[variant] = make(chan struct{}, 1)

		joined[variant] = make(chan struct{})
		go ownershipStepWorker(scenes[variant].w, requests[variant], done[variant], joined[variant])
	}

	t.Cleanup(func() {
		if release != nil {
			release()
		}

		for variant := range 2 {
			close(requests[variant])
			ownershipWait(t, joined[variant], "concurrent world worker shutdown")
		}
	})

	for step := range 48 {
		// Both worlds are inside caller work before release; no scheduling/sleep assumption.
		sharedGate := make(chan struct{})

		var once sync.Once

		release = func() { once.Do(func() { close(sharedGate) }) }

		for variant := range 2 {
			gate[variant] = sharedGate
			arrived[variant] = make(chan struct{}, 1)
			first[variant].Store(false)

			requests[variant] <- struct{}{}
		}

		for variant := range 2 {
			ownershipWait(t, arrived[variant], "both worlds entering task barrier")
		}

		if step == 24 {
			temporary := newOwnershipScene(t, 0, 2, nil)
			temporary.w.Step(1.0/60, 4)
			temporary.w.Close()
		}

		release()

		for variant := range 2 {
			ownershipWait(t, done[variant], "concurrent world step")

			ownershipCheckUserData(t, &scenes[variant], variant)

			observeOwnershipScene(readers[variant], scenes[variant], step)
		}
	}

	for variant := range 2 {
		if callbacks[variant].Load() == 0 {
			t.Fatal("world callback was not exercised")
		}
	}
}

type ownershipHeldTasks struct {
	release                       <-chan struct{}
	started, finishing            chan struct{}
	first, heldCompleted          atomic.Bool
	enqueued, completed, finished atomic.Int64
	tasks                         sync.WaitGroup
}

func (work *ownershipHeldTasks) enqueue(task func()) func() {
	work.enqueued.Add(1)

	held := work.first.CompareAndSwap(false, true)
	done := make(chan struct{})

	work.tasks.Go(func() {
		defer close(done)

		if held {
			close(work.started)
			<-work.release
		}

		task()
		work.completed.Add(1)

		if held {
			work.heldCompleted.Store(true)
		}
	})

	var (
		observed   atomic.Bool
		finishOnce sync.Once
	)

	return func() {
		finishOnce.Do(func() {
			if held {
				close(work.finishing)
			}
		})
		<-done

		if observed.CompareAndSwap(false, true) {
			work.finished.Add(1)
		}
	}
}

func ownershipCheckCompletedTasks(t *testing.T, enqueued, completed *atomic.Int64, message string) {
	t.Helper()

	if enqueued.Load() != completed.Load() {
		t.Fatal(message)
	}
}

func ownershipStepWorker(w *box2d.World, requests <-chan struct{}, done chan<- struct{}, joined chan struct{}) {
	defer close(joined)

	for range requests {
		w.Step(1.0/60, 4)

		done <- struct{}{}
	}
}

func ownershipCheckUserData(t *testing.T, s *ownershipScene, variant int) {
	t.Helper()

	if s.w.UserData() != variant {
		t.Fatal("world user data crossed owners")
	}

	for i, b := range s.bodies {
		if b.UserData() != variant*100+i {
			t.Fatal("body user data crossed owners")
		}
	}
}
