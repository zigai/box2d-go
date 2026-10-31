// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "benchmarks.h"
#include "box2d/box2d.h"
#include "determinism.h"
#include "human.h"
#include "utils.h"

#include <errno.h>
#include <limits.h>
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int positive_integer(const char* text) {
    char* end  = NULL;
    errno      = 0;
    long value = strtol(text, &end, 10);
    if (errno != 0 || end == text || *end != '\0' || value < 1 || value > INT_MAX) { return 0; }

    return (int)value;
}

int main(int argc, char** argv) {
    if (argc > 5) {
        fprintf(stderr, "usage: bench_scenes [samples] [workers] [cpus] [steps]\n");
        return 2;
    }

    int samples = argc > 1 ? positive_integer(argv[1]) : 6;
    int workers = argc > 2 ? positive_integer(argv[2]) : 1;
    int cpus    = argc > 3 ? positive_integer(argv[3]) : 2;
    int steps   = argc > 4 ? positive_integer(argv[4]) : 300;
    if (samples < 1 || workers < 1 || workers > B2_MAX_WORKERS || cpus < 1 || steps < 1) {
        fprintf(
            stderr,
            "samples/cpus/steps must be positive integers; workers must be in 1..%d\n",
            B2_MAX_WORKERS
        );

        return 2;
    }
    // Matched custom workload: all requested steps from step zero, no timed setup/teardown.
#if defined(BOX2D_DOUBLE_PRECISION)

    const char* precision = "double";
#else
    const char* precision = "float";
#endif
#if defined(NDEBUG)
    const char* scene_config = "release";
#else
    const char* scene_config = "debug";
#endif

    printf(
        "box2d-bench: precision=%s scene-config=%s steps=%d timestep=1/60 substeps=4 seed=12345 "
        "continuous=true "
        "rebuild-static=false cpus=%d\n",
        precision,
        scene_config,
        steps,
        cpus
    );

    const char* names[] = {
        "JointGrid",
        "LargePyramid",
        "ManyPyramids",
        "Rain",
        "Spinner",
        "Smash",
        "Tumbler",
        "Washer",
        "Junkyard",
        "Sleep",
        "Compounds",
        "Queries",
        "TreeCast",
        "TileWorld",
        "FallingHinges",
        "Human"
    };

    for (int scene = 0; scene < 16; ++scene) {
        for (int sample = 0; sample < samples; ++sample) {
            g_randomSeed            = 12345;
            b2WorldDef def          = b2DefaultWorldDef();
            def.workerCount         = workers;
            b2WorldId        world  = b2CreateWorld(&def);
            FallingHingeData hinges = {0};
            Human            human  = {0};

            switch (scene) {
                case 0: CreateJointGrid(world); break;
                case 1: CreateLargePyramid(world); break;
                case 2: CreateManyPyramids(world); break;
                case 3: CreateRain(world); break;
                case 4: CreateSpinner(world); break;
                case 5: CreateSmash(world); break;
                case 6: CreateTumbler(world); break;
                case 7: CreateWasher(world); break;
                case 8: CreateJunkyard(world); break;
                case 9: CreateSleep(world); break;
                case 10: CreateCompounds(world); break;
                case 11: CreateQueries(world); break;
                case 12: CreateTreeCast(world); break;
                case 13: CreateTileWorld(world); break;
                case 14: hinges = CreateFallingHinges(world); break;
                case 15:
                    CreateHuman(&human, world, (b2Pos){0, 5}, 1, 0.05f, 5, 0.5f, 1, NULL, true);

                    break;
                default: break;
            }

            uint64_t ticks = b2GetTicks();

            for (int step = 0; step < steps; ++step) {
                switch (scene) {
                    case 3: StepRain(world, step); break;
                    case 4: StepSpinner(world, step); break;
                    case 8: StepJunkyard(world, step); break;
                    case 9: StepSleep(world, step); break;
                    case 11: StepQueries(world, step); break;
                    case 12: StepTreeCast(world, step); break;
                    case 13: StepTileWorld(world, step); break;
                    case 15:
                        switch (step) {
                            case 0: Human_SetVelocity(&human, (b2Vec2){1, 0}); break;
                            case 20: Human_ApplyRandomAngularImpulse(&human, 0.5f); break;
                            case 40: Human_SetJointFrictionTorque(&human, 0); break;
                            case 60: Human_SetJointFrictionTorque(&human, 0.1f); break;
                            case 80: Human_SetJointSpringHertz(&human, 0); break;
                            case 100: Human_SetJointSpringHertz(&human, 2); break;
                            case 120: Human_SetJointDampingRatio(&human, 0.75f); break;
                            case 140: Human_EnableSensorEvents(&human, true); break;
                            case 160: Human_SetScale(&human, 1.5f); break;
                            case 180: Human_EnableSensorEvents(&human, false); break;
                            default: break;
                        }

                        break;
                    default: break;
                }

                b2World_Step(world, 1.0f / 60.0f, 4);

                if (scene == 14) { UpdateFallingHinges(world, &hinges); }
            }

            double nanoseconds = 1000000.0 * b2GetMilliseconds(ticks);
            if (!isfinite(nanoseconds) || nanoseconds <= 0) {
                fprintf(
                    stderr,
                    "invalid timing for %s workers=%d sample=%d\n",
                    names[scene],
                    workers,
                    sample
                );

                return 1;
            }

            if (scene == 14) { DestroyFallingHinges(&hinges); }
            if (scene == 15) { DestroyHuman(&human); }
            if (scene == 12) { DestroyTreeCast(); }
            if (scene == 13) { DestroyTileWorld(); }

            b2DestroyWorld(world);

            if (cpus == 1) {
                printf(
                    "BenchmarkScenes/%s/workers=%d 1 %.0f ns/op\n",
                    names[scene],
                    workers,
                    nanoseconds
                );
            } else {
                printf(
                    "BenchmarkScenes/%s/workers=%d-%d 1 %.0f ns/op\n",
                    names[scene],
                    workers,
                    cpus,
                    nanoseconds
                );
            }
        }
    }

    return 0;
}
