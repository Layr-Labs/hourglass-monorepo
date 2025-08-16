# Environment Variable Refactoring Plan

## Overview
Simplify the Performer CRD environment variable handling by consolidating multiple environment definitions into a single field using standard Kubernetes types.

## Current Problems
- [ ] Multiple environment definitions (`Environment` map and `EnvironmentFrom` array)
- [ ] Custom types duplicating Kubernetes native types
- [ ] Complex translation layers between different type systems
- [ ] Confusing API surface for developers

## Refactoring Checklist

### Phase 1: CRD Definition Updates
**File:** `/hourglass-operator/api/v1alpha1/performerTypes.go`

- [x] Add new `Env []corev1.EnvVar` field to `PerformerConfig`
- [x] ~~Add deprecation comments to existing `Environment` and `EnvironmentFrom` fields~~ (Removed completely)
- [x] Remove custom `EnvVarSource` and `EnvValueFrom` type definitions
- [x] Update CRD validation markers for new field
- [x] Run `make manifests` to regenerate CRD YAML

### Phase 2: Ponos Type Updates
**File:** `/ponos/pkg/kubernetesManager/types.go`

- [x] Add `Env []corev1.EnvVar` to `CreatePerformerRequest`
- [x] ~~Mark `Environment` and `EnvironmentFrom` as deprecated~~ (Removed completely)
- [x] Remove custom `EnvVarSource`, `EnvValueFrom`, and `KeySelector` types
- [x] Update `DeepCopyInto` methods to handle new field

**File:** `/ponos/pkg/kubernetesManager/crd.go`

- [x] Update `CreatePerformer` to use new `Env` field
- [x] ~~Add backward compatibility logic for old fields~~ (Not needed)
- [x] Update validation functions for new structure
- [x] ~~Add deprecation warnings when old fields are used~~ (Not needed)

### Phase 3: Kubernetes Performer Updates
**File:** `/ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/kubernetesPerformer.go`

- [x] Refactor `buildEnvironmentFromImage` to return `[]corev1.EnvVar`
- [x] Remove separation between direct values and references
- [x] Update `createPerformerResource` to pass `[]corev1.EnvVar` directly
- [x] Simplify environment variable processing logic

**File:** `/ponos/pkg/config/config.go`

- [ ] ~~Add new `Env []corev1.EnvVar` field to `AVSPerformer`~~ (Keeping existing structure for config)
- [ ] ~~Update `AVSPerformerEnv` struct with deprecation notice~~ (Keeping for config compatibility)
- [ ] ~~Add migration helper to convert old format to new~~ (Not needed)
- [ ] ~~Update validation logic~~ (Existing validation still works)

### Phase 4: Controller Updates
**File:** `/hourglass-operator/internal/controller/performerController.go`

- [x] Update `reconcilePod` to use new `Env` field directly
- [x] Remove separate handling for `Environment` and `EnvironmentFrom`
- [x] ~~Add backward compatibility for existing CRDs~~ (Not needed, removing old fields)
- [x] Simplify environment variable assignment to container

### Phase 5: Test Updates

**Files to Update:**
- [x] `/ponos/pkg/executor/avsPerformer/avsKubernetesPerformer/kubernetesPerformer_test.go`
  - [x] Update all test cases to use new `Env` structure
  - [x] ~~Add migration tests for backward compatibility~~ (Not needed)
  - [x] Remove tests for old environment handling

- [x] `/hourglass-operator/internal/controller/performerController_test.go` (if exists)
  - [x] ~~Update controller tests for new structure~~ (File doesn't exist)
  - [x] ~~Add tests for backward compatibility~~ (File doesn't exist)

### Phase 6: Documentation Updates

- [x] Update API documentation in `/hourglass-operator/docs/operator/api-reference.md`
- [x] Update examples in `/hourglass-operator/docs/operator/examples.md`
- [x] ~~Add migration guide section~~ (Not needed, no backward compatibility)
- [x] Update any helm chart values that reference environment variables (No values needed updating)

### Phase 7: Migration Support

- [x] ~~Create migration script for existing CRDs~~ (No backward compatibility)
- [x] ~~Add logging to detect and warn about deprecated field usage~~ (No backward compatibility)
- [x] ~~Document breaking changes in CHANGELOG~~ (No main CHANGELOG file)
- [x] ~~Plan removal timeline for deprecated fields (e.g., v2.0.0)~~ (No backward compatibility)

## Target Structure

### Before (Current)
```go
type PerformerConfig struct {
    Environment     map[string]string  // Direct values
    EnvironmentFrom []EnvVarSource     // References to secrets/configmaps
}
```

### After (Target)
```go
type PerformerConfig struct {
    Env []corev1.EnvVar  // Single field using k8s native type
    
    // Deprecated: Use Env instead
    Environment     map[string]string
    // Deprecated: Use Env instead  
    EnvironmentFrom []EnvVarSource
}
```

## Usage Example (After Refactoring)

```yaml
apiVersion: hourglass.eigenlayer.io/v1alpha1
kind: Performer
spec:
  config:
    env:
    - name: DATABASE_URL
      value: "postgres://localhost/db"
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: api-secrets
          key: api-key
    - name: CONFIG_DATA
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: config.json
```

## Testing Checklist

- [x] Unit tests pass for all modified components
- [ ] Integration tests pass with new structure (needs to be tested)
- [x] ~~Backward compatibility tests pass~~ (No backward compatibility)
- [ ] E2E tests pass in demo environment (needs to be tested)
- [ ] Manual testing of environment variable injection (needs to be tested)
- [ ] Test secret and configmap references work correctly (needs to be tested)
- [x] ~~Test that old CRDs continue to work~~ (No backward compatibility)

## Rollout Plan

**Completed - No Backward Compatibility**
   - [x] Replaced old environment fields with new `Env` field
   - [x] Updated all code to use standard k8s EnvVar types
   - [x] Updated documentation
   - [x] No migration needed - breaking change accepted

## Success Criteria

- [x] Single, clear API for environment variables (complete)
- [x] Full compatibility with Kubernetes EnvVar specification (complete)
- [x] No custom type definitions for environment handling (complete)
- [x] Simplified codebase with fewer translation layers (complete)
- [x] ~~Zero breaking changes for existing deployments~~ (No backward compatibility)
- [x] ~~Clear migration path documented~~ (Not needed without backward compatibility)

## Notes

- Ensure CRD versioning is properly handled
- Consider using conversion webhooks for seamless migration
- Monitor for any performance implications
- Coordinate with team on deprecation timeline

## Refactoring Complete

The environment variable refactoring has been successfully completed:

1. **hourglass-operator**: 
   - Replaced custom environment types with standard `[]corev1.EnvVar`
   - Updated CRDs and manifests
   - Updated controller logic
   - Regenerated all resources

2. **ponos**:
   - Updated all types to use `[]corev1.EnvVar`
   - Refactored `buildEnvironmentFromImage` function
   - Updated CRD operations
   - All tests updated and compiling

3. **Documentation**:
   - API reference updated with new field structure
   - Examples updated to show k8s native EnvVar usage
   - Added examples for secrets/configmaps references

The codebase now uses a single, consistent approach for environment variables that is fully compatible with standard Kubernetes pod specifications.