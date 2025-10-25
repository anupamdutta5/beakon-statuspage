# Phase 1, Week 2 (Day 3-5): Escalation & On-Call API Clients - Implementation Complete

**Date**: 2025-10-24
**Feature**: Escalation & On-Call Management API Clients
**Service**: tenant-admin-frontend (Port 3002)
**Backend**: monitoring-service (Port 8092)
**Status**: ✅ **API CLIENTS COMPLETE** (UI pages pending)

---

## 📋 Overview

Successfully created comprehensive API clients for Escalation Policies and On-Call Schedules. These clients provide full CRUD operations and helper methods for managing alert escalation workflows and on-call rotations.

---

## ✅ Completed Work

### 1. Escalation Policies API Client
**File**: [lib/api/escalation.ts](tenant-admin-frontend/lib/api/escalation.ts)

**Interfaces**:
```typescript
export interface EscalationLevel {
  level: number;                    // 1, 2, 3, etc.
  delay_minutes: number;            // Delay before escalating to this level
  notify_users?: string[];          // User IDs (UUIDs) to notify
  notify_schedule?: number;         // On-call schedule ID
  notify_channels: string[];        // email, sms, webhook, slack
}

export interface EscalationPolicy {
  id: number;
  tenant_id: string;
  name: string;
  description?: string;
  levels: string;                   // JSON string of EscalationLevel[]
  is_default: boolean;
  created_at: string;
  updated_at: string;
}
```

**API Methods**:
- `getPolicies()` - List all escalation policies
- `getPolicy(id)` - Get specific policy
- `createPolicy(data)` - Create new policy
- `updatePolicy(id, data)` - Update policy
- `deletePolicy(id)` - Delete policy
- `parseLevels(levelsJson)` - Helper to parse JSON levels
- `getDefaultPolicy()` - Get the default policy

**Key Features**:
- Multi-level escalation support
- Configurable delay between levels
- Multiple notification channels per level
- User or schedule-based notifications
- Default policy support

---

### 2. On-Call Schedules API Client
**File**: [lib/api/oncall.ts](tenant-admin-frontend/lib/api/oncall.ts)

**Interfaces**:
```typescript
export type RotationType = 'daily' | 'weekly' | 'custom';

export interface OnCallParticipant {
  user_id: string;                  // UUID
  order: number;                    // Rotation order
  name?: string;                    // For display
  email?: string;                   // For display
}

export interface OnCallSchedule {
  id: number;
  tenant_id: string;
  name: string;
  rotation_type: RotationType;
  rotation_start: string;           // ISO datetime
  rotation_interval_hours: number;  // Default: 168 (weekly)
  participants: string;             // JSON string of OnCallParticipant[]
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface OnCallRotation {
  schedule_id: number;
  current_user_id: string;
  current_user_name?: string;
  rotation_start: string;
  rotation_end: string;
  next_user_id?: string;
  next_user_name?: string;
}

export interface OnCallOverride {
  id: number;
  schedule_id: number;
  user_id: string;
  start_time: string;
  end_time: string;
  reason?: string;
  created_by: string;
  created_at: string;
}
```

**API Methods**:
- `getSchedules(includeInactive?)` - List all schedules
- `getSchedule(id)` - Get specific schedule
- `createSchedule(data)` - Create new schedule
- `updateSchedule(id, data)` - Update schedule
- `deleteSchedule(id)` - Delete schedule
- `getWhoIsOnCall(scheduleId)` - Get current on-call person
- `getRotations(scheduleId, startDate, endDate)` - Get rotations for calendar
- `createOverride(scheduleId, data)` - Create on-call override
- `getOverrides(scheduleId)` - List overrides
- `deleteOverride(scheduleId, overrideId)` - Delete override
- `parseParticipants(participantsJson)` - Helper to parse JSON participants
- `calculateCurrentOnCall(schedule)` - Client-side calculation of current on-call
- `getActiveSchedules()` - Filter active schedules

**Key Features**:
- Daily, weekly, and custom rotations
- Configurable rotation intervals
- Multi-participant rotations
- On-call overrides (for vacations, swaps)
- Calendar view support
- Client-side on-call calculation

---

## 🎨 API Design Highlights

### Escalation Policies:

**Multi-Level Escalation Example**:
```json
{
  "name": "Critical Alerts Policy",
  "description": "3-tier escalation for critical alerts",
  "levels": [
    {
      "level": 1,
      "delay_minutes": 0,
      "notify_users": ["user-uuid-1"],
      "notify_channels": ["email", "slack"]
    },
    {
      "level": 2,
      "delay_minutes": 15,
      "notify_schedule": 1,
      "notify_channels": ["email", "sms", "slack"]
    },
    {
      "level": 3,
      "delay_minutes": 30,
      "notify_users": ["manager-uuid-1", "manager-uuid-2"],
      "notify_channels": ["email", "sms", "webhook"]
    }
  ],
  "is_default": true
}
```

**Escalation Flow**:
1. **Level 1** (T+0min): Notify specific user via email & Slack
2. **Level 2** (T+15min): If not acknowledged, notify on-call person via email, SMS & Slack
3. **Level 3** (T+30min): If still not acknowledged, notify managers via email, SMS & webhook

### On-Call Schedules:

**Weekly Rotation Example**:
```json
{
  "name": "Engineering On-Call",
  "rotation_type": "weekly",
  "rotation_start": "2025-10-21T00:00:00Z",
  "rotation_interval_hours": 168,
  "participants": [
    { "user_id": "user-1-uuid", "order": 1, "name": "Alice" },
    { "user_id": "user-2-uuid", "order": 2, "name": "Bob" },
    { "user_id": "user-3-uuid", "order": 3, "name": "Charlie" }
  ],
  "is_active": true
}
```

**Rotation Flow**:
- Week 1: Alice is on-call
- Week 2: Bob is on-call
- Week 3: Charlie is on-call
- Week 4: Alice is on-call (cycle repeats)

**On-Call Override Example**:
```json
{
  "schedule_id": 1,
  "user_id": "user-4-uuid",
  "start_time": "2025-10-25T00:00:00Z",
  "end_time": "2025-10-27T00:00:00Z",
  "reason": "Bob is on vacation, David covering"
}
```

---

## 🔗 Backend Integration

### Escalation Endpoints (monitoring-service:8092):
```
GET    /api/v1/escalations/policies          → List policies
GET    /api/v1/escalations/policies/:id      → Get policy
POST   /api/v1/escalations/policies          → Create policy
PUT    /api/v1/escalations/policies/:id      → Update policy
DELETE /api/v1/escalations/policies/:id      → Delete policy
```

### On-Call Endpoints (monitoring-service:8092):
```
GET    /api/v1/oncall/schedules                      → List schedules
GET    /api/v1/oncall/schedules/:id                  → Get schedule
POST   /api/v1/oncall/schedules                      → Create schedule
PUT    /api/v1/oncall/schedules/:id                  → Update schedule
DELETE /api/v1/oncall/schedules/:id                  → Delete schedule
GET    /api/v1/oncall/schedules/:id/current          → Get current on-call
GET    /api/v1/oncall/schedules/:id/rotations        → Get rotations (calendar)
POST   /api/v1/oncall/schedules/:id/overrides        → Create override
GET    /api/v1/oncall/schedules/:id/overrides        → List overrides
DELETE /api/v1/oncall/schedules/:id/overrides/:oid   → Delete override
```

### Database Schemas:

**escalation_policies table**:
```sql
- id (uint)
- tenant_id (uuid)
- name (varchar)
- description (text)
- levels (jsonb) -- Array of EscalationLevel
- is_default (boolean)
- created_at, updated_at
```

**on_call_schedules table**:
```sql
- id (uint)
- tenant_id (uuid)
- name (varchar)
- rotation_type (varchar: daily/weekly/custom)
- rotation_start (timestamp)
- rotation_interval_hours (int, default: 168)
- participants (jsonb) -- Array of OnCallParticipant
- is_active (boolean)
- created_at, updated_at
```

**escalation_trackers table** (for tracking active escalations):
```sql
- id (uint)
- tenant_id (uuid)
- incident_id (uuid)
- policy_id (uint)
- current_level (int)
- is_resolved (boolean)
- last_escalated (timestamp)
- notified_users (text, JSON array)
- created_at, updated_at
```

---

## 🎯 Feature Completion

| Component | Status | File |
|-----------|--------|------|
| Escalation API Client | ✅ Complete | `lib/api/escalation.ts` |
| On-Call API Client | ✅ Complete | `lib/api/oncall.ts` |
| Escalation Policies UI | ⏳ Pending | To be created |
| On-Call Schedules UI | ⏳ Pending | To be created |
| Calendar View | ⏳ Pending | To be created |
| Override Management UI | ⏳ Pending | To be created |

**API Clients Progress**: **100% Complete** ✅

**UI Pages Progress**: **0% Complete** (Next task)

---

## 📝 Use Case Examples

### Use Case 1: Critical Alert Escalation

**Scenario**: API monitor fails 3 times, creates critical alert

**Escalation Flow**:
1. **T+0min**: Alert created, Level 1 triggered
   - Notify primary on-call via email + Slack
   - Set 15-minute timer

2. **T+15min**: If not acknowledged, Level 2 triggered
   - Notify backup on-call via email + SMS + Slack
   - Set 15-minute timer

3. **T+30min**: If still not acknowledged, Level 3 triggered
   - Notify engineering managers via email + SMS + PagerDuty
   - Create high-priority incident ticket

**Implementation**:
```typescript
const policy = await escalationAPI.createPolicy({
  name: "Critical API Alerts",
  levels: [
    {
      level: 1,
      delay_minutes: 0,
      notify_schedule: 1, // Primary on-call schedule
      notify_channels: ["email", "slack"]
    },
    {
      level: 2,
      delay_minutes: 15,
      notify_schedule: 2, // Backup on-call schedule
      notify_channels: ["email", "sms", "slack"]
    },
    {
      level: 3,
      delay_minutes: 30,
      notify_users: ["manager-uuid-1", "manager-uuid-2"],
      notify_channels: ["email", "sms", "pagerduty"]
    }
  ],
  is_default: true
});
```

### Use Case 2: Weekly On-Call Rotation

**Scenario**: 4-person team with weekly rotations

**Setup**:
```typescript
const schedule = await oncallAPI.createSchedule({
  name: "Engineering Primary On-Call",
  rotation_type: "weekly",
  rotation_start: "2025-10-21T00:00:00Z", // Monday midnight
  rotation_interval_hours: 168, // 7 days
  participants: [
    { user_id: "alice-uuid", order: 1, name: "Alice" },
    { user_id: "bob-uuid", order: 2, name: "Bob" },
    { user_id: "charlie-uuid", order: 3, name: "Charlie" },
    { user_id: "david-uuid", order: 4, name: "David" }
  ]
});
```

**Rotation Schedule**:
- Oct 21-27: Alice on-call
- Oct 28 - Nov 3: Bob on-call
- Nov 4-10: Charlie on-call
- Nov 11-17: David on-call
- Nov 18-24: Alice on-call (cycle repeats)

**Query Current On-Call**:
```typescript
const rotation = await oncallAPI.getWhoIsOnCall(schedule.id);
console.log(`Current on-call: ${rotation.current_user_name}`);
console.log(`Next on-call: ${rotation.next_user_name}`);
```

### Use Case 3: On-Call Override (Vacation Coverage)

**Scenario**: Bob is on vacation Oct 28-30, David covers

**Implementation**:
```typescript
const override = await oncallAPI.createOverride(schedule.id, {
  schedule_id: schedule.id,
  user_id: "david-uuid",
  start_time: "2025-10-28T00:00:00Z",
  end_time: "2025-10-31T00:00:00Z",
  reason: "Bob on vacation, David covering",
  created_by: "admin-uuid"
});
```

**Result**:
- Oct 21-27: Alice on-call (as scheduled)
- Oct 28-30: David on-call (override)
- Oct 31 - Nov 3: Bob on-call (resumes normal schedule)

---

## 🔮 Future UI Implementation

### Escalation Policies Page (`/admin/escalation-policies`):

**Features to Implement**:
- List all policies with default badge
- Create policy form with level builder
- Drag-and-drop level reordering
- User/schedule selector for each level
- Channel selection (email, SMS, Slack, etc.)
- Preview escalation flow
- Set default policy toggle
- Delete with confirmation

**Components Needed**:
1. `EscalationPoliciesList` - Table of policies
2. `EscalationPolicyForm` - Create/edit form
3. `EscalationLevelBuilder` - Multi-level editor
4. `EscalationFlowPreview` - Visual timeline of escalation

### On-Call Schedules Page (`/admin/oncall-schedules`):

**Features to Implement**:
- List all schedules with active/inactive
- Create schedule form
- Rotation type selector (daily/weekly/custom)
- Participant management (add/remove/reorder)
- Who's on-call now display
- Calendar view of rotations
- Override management
- Schedule history

**Components Needed**:
1. `OnCallSchedulesList` - Table of schedules
2. `OnCallScheduleForm` - Create/edit form
3. `OnCallCalendar` - Calendar view with rotations
4. `OnCallCurrentDisplay` - "Who's on-call now" widget
5. `OnCallOverrideForm` - Create override
6. `OnCallParticipantManager` - Manage rotation participants

---

## 📊 API Client Statistics

### Escalation API:
- **Methods**: 7
- **Interfaces**: 5
- **Lines of Code**: ~90

### On-Call API:
- **Methods**: 13
- **Interfaces**: 6
- **Lines of Code**: ~160

### Total:
- **2 API Clients**
- **20 Methods**
- **11 Interfaces**
- **~250 Lines of Code**

---

## 🎓 Key Achievements

### Technical Excellence:
1. **Comprehensive Coverage**: Full CRUD for both resources
2. **Helper Methods**: Client-side calculations and parsing
3. **Type Safety**: Full TypeScript typing
4. **JSON Handling**: Automatic serialization/deserialization
5. **Error Handling**: Graceful error returns
6. **Future-Proof**: Overrides support for flexibility

### Business Value:
1. **Reduced Alert Fatigue**: Smart escalation prevents over-notification
2. **Clear Accountability**: Always know who's on-call
3. **Flexible Rotations**: Daily, weekly, or custom intervals
4. **Vacation Coverage**: Override system for PTO/swaps
5. **Multi-Channel**: Email, SMS, Slack, Webhook, PagerDuty support

---

## 🚀 Next Steps

**Immediate (Week 2, Day 3-5 Continued)**:
- [ ] Create escalation policies list page
- [ ] Create escalation policy form with level builder
- [ ] Create on-call schedules list page
- [ ] Create on-call schedule form
- [ ] Create calendar view for rotations
- [ ] Create override management UI
- [ ] Create "Who's on-call now" widget

**Future Enhancements**:
- [ ] Mobile app for on-call notifications
- [ ] SMS confirmation workflow
- [ ] Escalation analytics (time to acknowledge)
- [ ] On-call schedule templates
- [ ] Multi-timezone support
- [ ] Integration with HR systems (sync PTO)
- [ ] On-call compensation tracking

---

## 📚 Related Documentation

- [PHASE1_WEEK2_ALERTS_COMPLETE.md](PHASE1_WEEK2_ALERTS_COMPLETE.md) - Alert management UI
- [PHASE1_WEEK1_COMPLETE.md](PHASE1_WEEK1_COMPLETE.md) - Week 1 completion
- [Escalation Handler](../monitoring-service/internal/handlers/escalation_handler.go) - Backend handler
- [On-Call Handler](../monitoring-service/internal/handlers/oncall_handler.go) - Backend handler
- [Escalation Service](../monitoring-service/internal/services/escalation_service.go) - Backend service
- [On-Call Service](../monitoring-service/internal/services/oncall_service.go) - Backend service

---

**Completed By**: Claude (AI Assistant)
**Date**: 2025-10-24
**Status**: ✅ **API CLIENTS COMPLETE - UI PAGES NEXT**
