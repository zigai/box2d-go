// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#define SCENES_LIMIT (32u*1024u*1024u)
#define SCENES_WINDOW (1u<<20)

static uint8_t* scenes_snapshots[6];
static size_t scenes_sizes[6];
static int scenes_steps[6];
static int scenes_checkpoint_count;

static uint32_t scenes_read_u32(const uint8_t* p)
{
 return (uint32_t)p[0] | (uint32_t)p[1]<<8 | (uint32_t)p[2]<<16 | (uint32_t)p[3]<<24;
}

static void scenes_varint(uint8_t* out, size_t* size, uint32_t value)
{
 while (value >= 128) { out[(*size)++] = (uint8_t)(value | 128); value >>= 7; }
 out[(*size)++] = (uint8_t)value;
}

static uint32_t scenes_dictionary_key(const uint8_t* p)
{
 return (scenes_read_u32(p)*UINT32_C(2654435761)) >> 12;
}

static uint8_t* scenes_lz(const uint8_t* raw, size_t size, size_t* packed_size)
{
 if (!(size <= SCENES_LIMIT)) abort();
 uint8_t* out = malloc(2*size+16);
 uint32_t* table = calloc(SCENES_WINDOW, sizeof(uint32_t));
 if (!(out != NULL && table != NULL)) abort();
 scenes_u32(out, (uint32_t)size);
 size_t used = 4, start = 0, i = 0;
 while (i+4 <= size)
 {
  uint32_t key = scenes_dictionary_key(raw+i);
  uint32_t entry = table[key];
  table[key] = (uint32_t)i+1;
  size_t count = 0, previous = entry ? entry-1 : 0;
  if (entry && i-previous <= SCENES_WINDOW && memcmp(raw+previous, raw+i, 4) == 0)
  {
   count = 4;
   while (count < 65535 && i+count < size && raw[previous+count] == raw[i+count]) ++count;
  }
  if (count == 0) { ++i; continue; }
  if (i > start)
  {
   scenes_varint(out, &used, (uint32_t)((i-start-1)<<1));
   memcpy(out+used, raw+start, i-start);
   used += i-start;
  }
  scenes_varint(out, &used, (uint32_t)(((count-4)<<1)|1));
  scenes_varint(out, &used, (uint32_t)(i-previous));
  for (size_t j = i+1; j < i+count && j+4 <= size; ++j)
   table[scenes_dictionary_key(raw+j)] = (uint32_t)j+1;
  i += count;
  start = i;
 }
 if (start < size)
 {
  scenes_varint(out, &used, (uint32_t)((size-start-1)<<1));
  memcpy(out+used, raw+start, size-start);
  used += size-start;
 }
 free(table);
 *packed_size = used;
 return out;
}

static void scenes_full(int step, const uint8_t* bytes, size_t size)
{
 if (!(scenes_checkpoint_count < 6 && size <= SCENES_LIMIT)) abort();
 int index = scenes_checkpoint_count++;
 scenes_steps[index] = step;
 scenes_sizes[index] = size;
 scenes_snapshots[index] = malloc(size ? size : 1);
 if (!(scenes_snapshots[index] != NULL)) abort();
 memcpy(scenes_snapshots[index], bytes, size);
}

static void scenes_flush(void)
{
 if (!(scenes_checkpoint_count == 6)) abort();
 const size_t width = sizeof(b2Pos)/2;
 const size_t components = 2*width+8, stride = components+4;
 size_t total = 0;
 uint32_t max_id = 0;
 for (int k = 0; k < 6; ++k)
 {
  if (!(scenes_steps[k] == (k == 0 ? -1 : k*60-1))) abort();
  if (!(scenes_sizes[k]%stride == 0 && scenes_sizes[k]/stride <= 100000)) abort();
  total += scenes_sizes[k]/stride;
  for (size_t i = 0; i < scenes_sizes[k]; i += stride)
  {
   uint32_t id = scenes_read_u32(scenes_snapshots[k]+i);
   if (!(id < SCENES_WINDOW)) abort();
   if (id > max_id) max_id = id;
  }
 }
 if (!(total*(components+5)+54 <= SCENES_LIMIT)) abort();
 uint8_t* raw = malloc(total*(components+5)+54);
 uint8_t* planes = malloc(total*components+1);
 uint8_t* history = calloc((size_t)max_id+1, components);
 if (!(raw != NULL && planes != NULL && history != NULL)) abort();
 memcpy(raw, "SCNL", 4);
 raw[4] = (uint8_t)width;
 raw[5] = 6;
 size_t used = 6, record = 0;
 for (int k = 0; k < 6; ++k)
 {
  scenes_u32(raw+used, (uint32_t)scenes_steps[k]);
  scenes_u32(raw+used+4, (uint32_t)(scenes_sizes[k]/stride));
  used += 8;
 }
 for (int k = 0; k < 6; ++k)
 {
  uint32_t preceding = UINT32_MAX;
  for (size_t i = 0; i < scenes_sizes[k]; i += stride)
  {
   const uint8_t* body = scenes_snapshots[k]+i;
   uint32_t id = scenes_read_u32(body);
   if (!(preceding == UINT32_MAX || id > preceding)) abort();
   scenes_varint(raw, &used, id-preceding);
   preceding = id;
   for (size_t lane = 0; lane < components; ++lane)
   {
    planes[lane*total+record] = body[4+lane]^history[(size_t)id*components+lane];
    history[(size_t)id*components+lane] = body[4+lane];
   }
   ++record;
  }
  free(scenes_snapshots[k]);
  scenes_snapshots[k] = NULL;
 }
 memcpy(raw+used, planes, total*components);
 used += total*components;
 size_t packed_size;
 uint8_t* packed = scenes_lz(raw, used, &packed_size);
 ORACLE_BYTES("full", packed, packed_size);
 free(packed);
 free(raw);
 free(planes);
 free(history);
 scenes_checkpoint_count = 0;
}
