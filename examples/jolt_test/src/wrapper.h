#pragma once

#ifdef __cplusplus
extern "C" {
#endif

typedef struct JoltC_System JoltC_System;
typedef struct JoltC_Body JoltC_Body;

typedef struct JoltC_Vec3 {
    float x;
    float y;
    float z;
} JoltC_Vec3;

JoltC_System* jolt_init();
JoltC_Body* jolt_create_sphere(JoltC_System* system, float x, float y, float z, float radius);
JoltC_Body* jolt_create_floor(JoltC_System* system);

void jolt_optimize_broadphase(JoltC_System* system);

void jolt_step(JoltC_System* system, float deltaTime);
JoltC_Vec3 jolt_get_position(JoltC_System* system, JoltC_Body* body);
void jolt_destroy(JoltC_System* system);

#ifdef __cplusplus
}
#endif
