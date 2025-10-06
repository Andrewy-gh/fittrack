### 🎯 Implementation Sequence

For key details and decisions see [Implementation Plan](./unified-route-implementation-plan.md).

**Phase 0**: Create factory functions ✅ Required
**Phase 1**: Fix `/workouts/new` (4 steps) ✅ Required
**Phase 1.5**: Fix RecentSets component 🚨 Critical
**Phase 2**: Add demo support to `/workouts/$workoutId/edit` (5 steps) ✅ Required
**Phase 3**: Type safety verification ✅ Required
**Phase 4**: Testing (demo mode focus) ✅ Required
**Phase 5**: Documentation updates ✅ Required

### 📊 Scope Boundaries

**In Scope**:
- ✅ Conditional mutations (API vs demo)
- ✅ Conditional queries (API vs demo)
- ✅ Draft persistence (both user types)
- ✅ RecentSets component fix
- ✅ TypeScript compilation (0 errors)
- ✅ Basic demo mode testing
- ✅ Documentation updates

**Out of Scope (Future Work)**:
- ⏸️ Query cache invalidation on auth transitions
- ⏸️ Demo-to-auth data migration
- ⏸️ Error handling standardization
- ⏸️ localStorage quota handling
- ⏸️ Multi-tab sync
- ⏸️ Comprehensive auth mode testing
- ⏸️ Navigation to workout detail after create

### ⏱️ Timeline

- **Original Estimate**: 2.5 hours
- **With Factory Pattern**: ~2 hours (saves implementation time)
- **RecentSets Fix**: +15 min (critical)
- **Final Estimate**: 2-2.5 hours

### 🚀 Ready to Implement

**Status**: ✅ **APPROVED - Ready for Implementation**

**Next Steps**:
1. [X] Start with Phase 0 (factory functions)
2. [X] Proceed to Phase 1 + 1.5 (`/workouts/new` + RecentSets)
3. [X] Test thoroughly before Phase 2
4. [ ] Complete Phase 2 (`/workouts/$workoutId/edit`)
5. [ ] Type check, test, document

**Important**
Only complete one step at a step. Mark each step as complete in the summary document after user had reviewed the changes and approved them.

**Blocker Resolution**: All critical questions answered, no blockers remain.

---

**End of Plan**
