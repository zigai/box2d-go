// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#define SCENE_STATE_LIMIT (32u*1024u*1024u)

static uint8_t* scene_snapshots[6];
static size_t scene_sizes[6];
static int scene_steps[6];
static int scene_checkpoint_count;

static void scene_full(int step, const uint8_t* bytes, size_t size)
{
 if (!(scene_checkpoint_count < 6 && size <= SCENE_STATE_LIMIT)) abort();
 int index = scene_checkpoint_count++;
 scene_steps[index] = step;
 scene_sizes[index] = size;
 scene_snapshots[index] = malloc(size ? size : 1);
 if (!(scene_snapshots[index] != NULL)) abort();
 memcpy(scene_snapshots[index], bytes, size);
}

// One record holds the six checkpoints in order: step, body count, then the
// bodies as written by scene_dump.
static void scene_flush(void)
{
 if (!(scene_checkpoint_count == 6)) abort();
 const size_t stride = 2*(sizeof(b2Pos)/2)+12;
 size_t total = 0;
 for (int k = 0; k < 6; ++k)
 {
  if (!(scene_steps[k] == (k == 0 ? -1 : k*60-1))) abort();
  if (!(scene_sizes[k]%stride == 0 && scene_sizes[k]/stride <= 100000)) abort();
  total += 8+scene_sizes[k];
 }
 uint8_t* raw = malloc(total);
 if (!(raw != NULL)) abort();
 size_t used = 0;
 for (int k = 0; k < 6; ++k)
 {
  scene_u32(raw+used, (uint32_t)scene_steps[k]);
  scene_u32(raw+used+4, (uint32_t)(scene_sizes[k]/stride));
  memcpy(raw+used+8, scene_snapshots[k], scene_sizes[k]);
  used += 8+scene_sizes[k];
  free(scene_snapshots[k]);
  scene_snapshots[k] = NULL;
 }
 ORACLE_BYTES("full", raw, used);
 free(raw);
 scene_checkpoint_count = 0;
}
