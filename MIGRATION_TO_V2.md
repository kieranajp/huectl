# Migration Plan: Hue API v1 to v2

## Executive Summary

Current implementation uses `huego` (v1.2.1) which is built on the **Hue API v1** (released ~2012). The **Hue API v2** was released in November 2021 and is the future of Hue development. Signify has stated that:
- New features (like dynamic scenes) will **only** be available on API v2
- API v1 will **eventually be removed** in the long term

## Current State

**Library:** `github.com/amimof/huego v1.2.1`
**API Version:** Hue API v1
**Features Used:**
- Group control (GetGroup, SetGroupState)
- Scene recall (RecallScene)
- Brightness adjustment
- On/off toggle
- Scene rotation

## Why Migrate?

### Features You'll Gain

1. **Native Dynamic Scenes Support**
   - Toggle dynamics on/off directly via API
   - Control color-changing behavior of scenes
   - This is what you're trying to implement now with workarounds

2. **Server-Sent Events (SSE)**
   - Real-time state updates from bridge
   - Much faster than polling
   - Could enable more responsive controls

3. **Better Security**
   - HTTPS only (v1 uses HTTP)
   - Improved authentication

4. **Future-Proofing**
   - All new Hue features will be v2-only
   - v1 API will be deprecated/removed eventually

### What You'll Lose

- API v2 requires:
  - Bridge firmware v1948 or higher
  - Not supported on v1 (round) bridges
  - v2 (square) bridges with updated firmware

## Recommended Migration Path

### Option 1: openhue-go (Recommended)

**Repository:** https://github.com/openhue/openhue-go
**Version:** v0.4.0 (April 2025)
**Status:** Actively maintained, auto-generated from OpenHue API spec

**Pros:**
- Official OpenHue project
- Auto-generated from OpenAPI spec (always up-to-date)
- Comprehensive v2 API coverage
- Active maintenance (recent 2025 release)

**Cons:**
- Larger API surface (more complex)
- Different architecture from huego
- Less documentation than huego

### Option 2: hue-v2-api

**Repository:** https://github.com/niomwungeri-fabrice/hue-v2-api

**Pros:**
- Specifically designed for v2 API

**Cons:**
- Less stars/adoption
- Unknown maintenance status
- Less mature than openhue-go

## Migration Strategy

### Phase 1: Research & Validation (1-2 hours)

1. **Check Bridge Compatibility**
   - SSH to `rpi-zero2w.local`
   - Check bridge firmware version
   - Verify it's >= v1948 and is a v2 bridge

2. **Prototype Key Operations**
   - Create a branch: `feature/api-v2-prototype`
   - Add openhue-go as dependency
   - Test basic operations:
     - Connect to bridge
     - Get room/group
     - Toggle lights
     - Adjust brightness
     - Recall scenes
     - **Toggle dynamics** (the feature you want!)

### Phase 2: Incremental Migration

1. **Create Abstraction Layer**
   - Define your own interface for bridge operations
   - Current `Bridge` interface is a good start
   - Extend it for v2-specific features (dynamics, SSE)

2. **Implement v2 Bridge Adapter**
   - Create `V2Bridge` struct implementing `Bridge` interface
   - Wrap openhue-go calls
   - Map v1 concepts to v2 (groups → grouped_light resources)

3. **Side-by-Side Testing**
   - Keep huego code initially
   - Add feature flag or config option
   - Test both implementations in parallel

### Phase 3: Full Cutover

1. **Remove huego Dependency**
2. **Update Tests**
   - Mock openhue-go instead of huego
3. **Update Documentation**
4. **Deploy to Pi**

## API Concept Mapping (v1 → v2)

| v1 API Concept | v2 API Concept | Notes |
|----------------|----------------|-------|
| Group | Grouped Light (resource) | Different structure, same purpose |
| Scene ID | Scene resource | Native dynamics control available |
| State (On, Bri) | GroupedLightPut | Similar but more features |
| Effect ("none", "colorloop") | Dynamics object | Proper dynamics support |

## Code Changes Required

### Current Handler Interface

```go
type Bridge interface {
    GetGroup(id int) (*huego.Group, error)
    SetGroupState(id int, state huego.State) (*huego.Response, error)
    RecallScene(sceneID string, groupID int) (*huego.Response, error)
}
```

### Proposed v2 Interface

```go
type Bridge interface {
    // Existing operations (adapted to v2)
    GetRoom(id string) (*openhue.Room, error)
    UpdateGroupedLight(id string, update openhue.GroupedLightPut) error
    RecallScene(sceneID string) error

    // New v2-specific operations
    SetDynamics(sceneID string, enabled bool) error
    SubscribeToEvents() (<-chan openhue.Event, error) // Optional: for SSE
}
```

### Example v2 Implementation

```go
// Toggle lights (v2 style)
home, _ := openhue.NewHome(openhue.LoadConfNoError())
rooms, _ := home.GetRooms()

for id, room := range rooms {
    for serviceId, serviceType := range room.GetServices() {
        if serviceType == openhue.ResourceIdentifierRtypeGroupedLight {
            light, _ := home.GetGroupedLightById(serviceId)
            home.UpdateGroupedLight(*light.Id, openhue.GroupedLightPut{
                On: light.Toggle(),
            })
        }
    }
}
```

## Estimated Effort

- **Phase 1 (Validation):** 1-2 hours
- **Phase 2 (Implementation):** 4-6 hours
- **Phase 3 (Testing & Deployment):** 2-3 hours

**Total:** ~8-12 hours of development time

## Risks & Mitigation

### Risk 1: Bridge Incompatibility
**Impact:** High
**Mitigation:** Check firmware version first (Phase 1). If incompatible, stay on v1.

### Risk 2: Breaking API Changes
**Impact:** Medium
**Mitigation:** Use abstraction layer to isolate changes. Keep v1 code until v2 is proven.

### Risk 3: Incomplete openhue-go Features
**Impact:** Medium
**Mitigation:** Prototype all needed operations first. Consider contributing missing features.

### Risk 4: Learning Curve
**Impact:** Low-Medium
**Mitigation:** Start with simple operations. The auto-generated code has good type safety.

## Decision Points

### Should You Migrate Now?

**YES, if:**
- ✅ You want native dynamics control
- ✅ You're building new features anyway
- ✅ Bridge firmware is compatible
- ✅ You have ~8-12 hours available

**NO, if:**
- ❌ Bridge doesn't support v2 API
- ❌ Current workaround (Effect toggle) is sufficient
- ❌ Time-constrained on other projects

### Recommended Approach

Given that you're already adding the dynamics feature, I'd recommend:

1. **Short-term:** Implement dynamics toggle using v1 Effect field (what we started)
2. **Medium-term:** Validate bridge compatibility and prototype v2 API
3. **Long-term:** Full migration to openhue-go for future-proofing

This gives you a working feature now while planning for the future.

## Resources

- [Hue API v2 Official Docs](https://developers.meethue.com/new-hue-api/)
- [openhue-go Repository](https://github.com/openhue/openhue-go)
- [openhue-go Package Docs](https://pkg.go.dev/github.com/openhue/openhue-go)
- [Hue API v2 Announcement](https://hueblog.com/2021/11/16/philips-hue-officially-announces-new-api/)
- [openHAB Hue v2 Migration Guide](https://www.openhab.org/addons/bindings/hue/)

## Next Steps

1. Check your bridge firmware version
2. Decide: Quick fix (v1 Effect) or full migration (v2)?
3. If migrating: Start Phase 1 prototype
4. If not migrating: Complete the dynamics toggle implementation with v1 API

---

**Created:** 2026-01-02
**Author:** Claude (via user request)
**Status:** Planning Document
