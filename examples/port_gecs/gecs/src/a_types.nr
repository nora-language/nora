package gecs

import "collections"

// ==========================================
// 1. INDEPENDENT ROOT TYPES
// ==========================================
pub type BitSet = struct {
    words: @i64[],
    size: i32,
    capacity: i32
}

[serialize]
pub type Entity = struct {
    id: i64
}

pub type AnyPair = struct {
    p1: any,
    p2: any
}

[serialize]
pub type Vec3 = struct {
    x: f32,
    y: f32,
    z: f32
}

pub type EventQueue = struct {
    events: @collections.Vector[any]
}

// ==========================================
// 2. TYPES DEPENDING ON ROOT TYPES
// ==========================================
pub type ComponentMeta = interface {
    fn Init(cap: i32) any
    fn SwapRemove(col_any: &any, row: i32) void
    fn MoveCol(src_any: &any, src_row: i32, dest_any: &any) void
    fn DropCol(col_any: &any) void
    fn Add(col_any: &any, val_any: &any) void
    fn Set(col_any: &any, row: i32, val_any: &any) void
}

pub type Archetype = struct {
    id: i32,
    mask: @BitSet,
    entities: @collections.Vector[Entity],
    columns: @collections.HashMap[i32, any],
    columnMeta: @collections.HashMap[i32, ComponentMeta],
    addEdges: @collections.HashMap[i32, i64],
    removeEdges: @collections.HashMap[i32, i64]
}

pub fn (a: #Archetype) GetEntities() #collections.Vector[Entity] {
    return #a.entities
}

pub fn (a: #Archetype) GetColumn(cid: i32) Option[#any] {
    return a.columns.Get[i32, any](#cid)
}

pub type RemoveResult = struct {
    moved_entity: Entity,
    ok: bool
}

pub type EntityRecord = struct {
    gen: i32,
    archetype: i64, // Mask hash of Archetype, or 0 if inactive/reserved
    row: i32
}

pub type ObserverCallback = struct {
    f: fn(#Entity, &any)
}

pub type ObserverTrigger = struct {
    adds: @collections.Vector[ObserverCallback],
    removes: @collections.Vector[ObserverCallback],
    sets: @collections.Vector[ObserverCallback]
}

pub type EventManager = struct {
    queues: @collections.HashMap[i32, @EventQueue] // map CID -> @EventQueue
}

[serialize]
pub type Parent = struct {
    id: Entity
}

pub type Children = struct {
    entities: @collections.Vector[Entity]
}

[serialize]
pub type ChildrenSnapshot = struct {
    ids: i64[]
}

[serialize]
pub type Transform = struct {
    position: Vec3,
    rotation: Vec3,
    scale: Vec3
}

[serialize]
pub type GlobalTransform = struct {
    position: Vec3
}

pub type Command = struct {
    entity: Entity,
    cid: i32,
    comp_any: any,
    action: i32 // 1 = Add, 2 = Remove, 3 = Despawn, 4 = Spawn
}

[serialize]
pub type EntityRecordRef = struct {
    id: i32,
    gen: i32
}

[serialize]
pub type ComponentSnapshot = struct {
    name: str,
    data: str
}

[serialize]
pub type ArchSnapshot = struct {
    mask: str[],
    entities: Entity[],
    components: ComponentSnapshot[]
}

[serialize]
pub type Snapshot = struct {
    next_id: i32,
    entities: EntityRecordRef[],
    archetypes: ArchSnapshot[]
}

pub type ComponentSerializer = interface {
    fn Name() str
    fn Serialize(col_any: &any) str
    fn Deserialize(json_str: str) any
    fn Meta() ComponentMeta
}

// ==========================================
// 3. TYPES DEPENDING ON WORLD-LEVEL PREREQUISITES
// ==========================================
[shared]
pub type World = struct {
    nextEntityID: i32,
    entities: @collections.Vector[EntityRecord],
    freeEntities: @collections.Vector[i32],
    archetypes: @collections.HashMap[i64, @Archetype],
    root: i64, // Root archetype hash
    observers: @collections.HashMap[i32, @ObserverTrigger], // CID -> @ObserverTrigger
    eventMgr: @EventManager,
    serializers: @collections.HashMap[i32, ComponentSerializer],
    nameToCid: @collections.HashMap[str, i32]
}

// ==========================================
// 4. TYPES DEPENDING ON WORLD
// ==========================================
pub type Query1[A] = struct {
    world: &World,
    mask: @BitSet,
    excl: @BitSet,
    cidA: i32
}

pub type Query2[A, B] = struct {
    world: &World,
    mask: @BitSet,
    excl: @BitSet,
    cidA: i32,
    cidB: i32
}

pub type CommandBuffer = struct {
    commands: @collections.Vector[Command],
    boxes: @collections.Vector[any]
}

pub type Access = struct {
    reads: @BitSet,
    writes: @BitSet
}

[shared]
pub type System = interface {
    fn Update(world: #World, cb: &CommandBuffer) void
    fn GetAccess() #Access
}

pub type Scheduler = struct {
    systems: @collections.Vector[System],
    batches: @collections.Vector[@collections.Vector[#System]]
}
