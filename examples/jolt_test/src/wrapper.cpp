#include "wrapper.h"

// Jolt includes
#include <Jolt/Jolt.h>
#include <Jolt/RegisterTypes.h>
#include <Jolt/Core/Factory.h>
#include <Jolt/Core/TempAllocator.h>
#include <Jolt/Core/JobSystemThreadPool.h>
#include <Jolt/Physics/PhysicsSettings.h>
#include <Jolt/Physics/PhysicsSystem.h>
#include <Jolt/Physics/Collision/Shape/BoxShape.h>
#include <Jolt/Physics/Collision/Shape/SphereShape.h>
#include <Jolt/Physics/Body/BodyCreationSettings.h>
#include <Jolt/Physics/Body/BodyActivationListener.h>

#include <iostream>

using namespace JPH;

namespace Layers {
    static constexpr ObjectLayer NON_MOVING = 0;
    static constexpr ObjectLayer MOVING = 1;
    static constexpr ObjectLayer NUM_LAYERS = 2;
}

namespace BroadPhaseLayers {
    static constexpr BroadPhaseLayer NON_MOVING(0);
    static constexpr BroadPhaseLayer MOVING(1);
    static constexpr uint NUM_LAYERS(2);
}

class BPLayerInterfaceImpl final : public BroadPhaseLayerInterface {
public:
    BPLayerInterfaceImpl() {
        mObjectToBroadPhase[Layers::NON_MOVING] = BroadPhaseLayers::NON_MOVING;
        mObjectToBroadPhase[Layers::MOVING] = BroadPhaseLayers::MOVING;
    }
    virtual uint GetNumBroadPhaseLayers() const override { return BroadPhaseLayers::NUM_LAYERS; }
    virtual BroadPhaseLayer GetBroadPhaseLayer(ObjectLayer inLayer) const override { return mObjectToBroadPhase[inLayer]; }
#if defined(JPH_EXTERNAL_PROFILE) || defined(JPH_PROFILE_ENABLED)
    virtual const char* GetBroadPhaseLayerName(BroadPhaseLayer inLayer) const override { return "Layer"; }
#endif
private:
    BroadPhaseLayer mObjectToBroadPhase[Layers::NUM_LAYERS];
};

class ObjectVsBroadPhaseLayerFilterImpl : public ObjectVsBroadPhaseLayerFilter {
public:
    virtual bool ShouldCollide(ObjectLayer inLayer1, BroadPhaseLayer inLayer2) const override {
        switch (inLayer1) {
            case Layers::NON_MOVING: return inLayer2 == BroadPhaseLayers::MOVING;
            case Layers::MOVING: return true;
            default: return false;
        }
    }
};

class ObjectLayerPairFilterImpl : public ObjectLayerPairFilter {
public:
    virtual bool ShouldCollide(ObjectLayer inLayer1, ObjectLayer inLayer2) const override {
        switch (inLayer1) {
            case Layers::NON_MOVING: return inLayer2 == Layers::MOVING;
            case Layers::MOVING: return true;
            default: return false;
        }
    }
};

struct JoltSystem {
    TempAllocatorImpl* temp_allocator;
    JobSystemThreadPool* job_system;
    BPLayerInterfaceImpl broad_phase_layer_interface;
    ObjectVsBroadPhaseLayerFilterImpl object_vs_broadphase_layer_filter;
    ObjectLayerPairFilterImpl object_vs_object_layer_filter;
    PhysicsSystem* physics_system;
};

// Required Jolt callbacks
#include <iostream>
#include <cstdarg>

void TraceImpl(const char* inFMT, ...) {
    va_list list;
    va_start(list, inFMT);
    char buffer[1024];
    vsnprintf(buffer, sizeof(buffer), inFMT, list);
    va_end(list);
    std::cout << buffer << std::endl;
}

#ifdef JPH_ENABLE_ASSERTS
bool AssertFailedImpl(const char* inExpression, const char* inMessage, const char* inFile, uint32_t inLine) {
    printf("%s:%d: (%s) %s\n", inFile, inLine, inExpression, inMessage != nullptr ? inMessage : "");
    fflush(stdout);
    return false; // Don't trigger breakpoint
}
#endif

// Define AssertFailed in JPH namespace so the linker finds it, since Release Jolt.lib might not have it
namespace JPH {
#ifdef JPH_ENABLE_ASSERTS
    AssertFailedFunction AssertFailed = AssertFailedImpl;
#endif
}

#include <windows.h>
LONG WINAPI CrashHandler(EXCEPTION_POINTERS* ExceptionInfo) {
    printf("CRASH! Exception Code: 0x%08X\n", ExceptionInfo->ExceptionRecord->ExceptionCode);
    fflush(stdout);
    ExitProcess(1);
}

extern "C" {

JoltC_System* jolt_init() {
    SetUnhandledExceptionFilter(CrashHandler);
    JPH::Trace = TraceImpl;

    JPH::RegisterDefaultAllocator();
    Factory::sInstance = new Factory();
    RegisterTypes();

    JoltSystem* sys = new JoltSystem();
    sys->temp_allocator = new TempAllocatorImpl(10 * 1024 * 1024);
    sys->job_system = new JobSystemThreadPool(cMaxPhysicsJobs, cMaxPhysicsBarriers, thread::hardware_concurrency() - 1);
    
    sys->physics_system = new PhysicsSystem();
    sys->physics_system->Init(1024, 0, 1024, 1024, 
        sys->broad_phase_layer_interface, 
        sys->object_vs_broadphase_layer_filter, 
        sys->object_vs_object_layer_filter);
        
    return reinterpret_cast<JoltC_System*>(sys);
}

JoltC_Body* jolt_create_sphere(JoltC_System* system, float x, float y, float z, float radius) {
    JoltSystem* sys = reinterpret_cast<JoltSystem*>(system);
    BodyInterface& body_interface = sys->physics_system->GetBodyInterface();
    
    SphereShapeSettings sphere_shape_settings(radius);
    ShapeSettings::ShapeResult sphere_shape_result = sphere_shape_settings.Create();
    ShapeRefC sphere_shape = sphere_shape_result.Get();
    
    BodyCreationSettings sphere_settings(sphere_shape, RVec3(x, y, z), Quat::sIdentity(), EMotionType::Dynamic, Layers::MOVING);
    JPH::Body* sphere = body_interface.CreateBody(sphere_settings);
    
    body_interface.AddBody(sphere->GetID(), EActivation::Activate);
    
    return reinterpret_cast<JoltC_Body*>(sphere);
}

JoltC_Body* jolt_create_floor(JoltC_System* system) {
    JoltSystem* sys = reinterpret_cast<JoltSystem*>(system);
    BodyInterface& body_interface = sys->physics_system->GetBodyInterface();
    
    BoxShapeSettings floor_shape_settings(JPH::Vec3(100.0f, 1.0f, 100.0f));
    ShapeSettings::ShapeResult floor_shape_result = floor_shape_settings.Create();
    ShapeRefC floor_shape = floor_shape_result.Get();
    
    BodyCreationSettings floor_settings(floor_shape, RVec3(0.0f, -1.0f, 0.0f), Quat::sIdentity(), EMotionType::Static, Layers::NON_MOVING);
    JPH::Body* floor = body_interface.CreateBody(floor_settings);
    
    if (floor == nullptr) {
        return nullptr;
    }
    
    body_interface.AddBody(floor->GetID(), EActivation::DontActivate);
    
    return reinterpret_cast<JoltC_Body*>(floor);
}

void jolt_optimize_broadphase(JoltC_System* system) {
    JoltSystem* sys = reinterpret_cast<JoltSystem*>(system);
    sys->physics_system->OptimizeBroadPhase();
}

void jolt_step(JoltC_System* system, float deltaTime) {
    JoltSystem* sys = reinterpret_cast<JoltSystem*>(system);
    sys->physics_system->Update(deltaTime, 1, sys->temp_allocator, sys->job_system);
}

JoltC_Vec3 jolt_get_position(JoltC_System* system, JoltC_Body* body) {
    JoltSystem* sys = reinterpret_cast<JoltSystem*>(system);
    BodyInterface& body_interface = sys->physics_system->GetBodyInterface();
    JPH::Body* b = reinterpret_cast<JPH::Body*>(body);
    RVec3 position = body_interface.GetPosition(b->GetID());
    Vec3 velocity = body_interface.GetLinearVelocity(b->GetID());
    
    // printf("Velocity = %f, %f, %f\n", velocity.GetX(), velocity.GetY(), velocity.GetZ()); fflush(stdout);
    
    JoltC_Vec3 result;
    result.x = position.GetX();
    result.y = position.GetY();
    result.z = position.GetZ();
    return result;
}

void jolt_destroy(JoltC_System* system) {
    JoltSystem* sys = reinterpret_cast<JoltSystem*>(system);
    delete sys->physics_system;
    delete sys->job_system;
    delete sys->temp_allocator;
    delete sys;
    UnregisterTypes();
    delete Factory::sInstance;
    Factory::sInstance = nullptr;
}

}
