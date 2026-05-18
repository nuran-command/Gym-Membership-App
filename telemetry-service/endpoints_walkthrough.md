# Telemetry Service: 12 gRPC Endpoints Implementation

We have successfully designed, generated, and fully implemented **exactly 12 gRPC endpoints** inside the `telemetry-service`. These endpoints provide complete coverage of session tracking, usage stats, health status, event logging, active session lookups, and direct CRUD operations for physical assets/memberships!

---

## Complete API Reference (12 Endpoints)

### 1. Existing core endpoints (5 endpoints)
1. **`GetUsageSession(GetUsageSessionRequest)`** $\rightarrow$ `UsageSessionResponse`
   * Get a usage session by its unique booking ID.
2. **`ListUserSessions(ListUserSessionsRequest)`** $\rightarrow$ `ListUserSessionsResponse`
   * List all gym usage sessions (active & completed) for a given member.
3. **`GetUsageStats(GetUsageStatsRequest)`** $\rightarrow$ `UsageStatsResponse`
   * Retrieve total number of sessions and total minutes spent at the gym for a user.
4. **`GetAssetUsageHistory(GetAssetUsageHistoryRequest)`** $\rightarrow$ `GetAssetUsageHistoryResponse`
   * Retrieve full timeline of all usage sessions for a specific physical asset (e.g. treadmill-1).
5. **`GetSystemUsageStats(GetSystemUsageStatsRequest)`** $\rightarrow$ `SystemUsageStatsResponse`
   * Aggregate statistics for all assets in the system, showing total training sessions and average duration in minutes.

---

### 2. Direct CRUD Operations (3 endpoints)
6. **`CreateUsageSession(CreateUsageSessionRequest)`** $\rightarrow$ `UsageSessionResponse`
   * **Purpose:** Manually start/create a usage session (e.g., for walk-in entries or manual check-ins).
   * **Fields:** `booking_id`, `user_id`, `asset_id`. Sets starting timestamp to `time.Now()`.
7. **`UpdateUsageSession(UpdateUsageSessionRequest)`** $\rightarrow$ `UsageSessionResponse`
   * **Purpose:** Manually edit or finalize a usage session.
   * **Fields:** `booking_id`, `ended_at` (ISO 8601 string), `duration_minutes`, `email_sent`.
8. **`DeleteUsageSession(DeleteUsageSessionRequest)`** $\rightarrow$ `DeleteUsageSessionResponse`
   * **Purpose:** Remove/delete a session record from persistent storage (e.g., database correction).
   * **Fields:** `booking_id`. Returns `success = true`.

---

### 3. Active Session Lookups (2 endpoints)
9. **`GetUserActiveSession(GetUserActiveSessionRequest)`** $\rightarrow$ `UsageSessionResponse`
   * **Purpose:** Find the user's currently active (ongoing) training session.
   * **Behavior:** Searches for a session where `user_id = $1` and `ended_at IS NULL`.
10. **`GetAssetActiveSession(GetAssetActiveSessionRequest)`** $\rightarrow$ `UsageSessionResponse`
    * **Purpose:** Find if a specific physical asset is currently occupied and by whom.
    * **Behavior:** Searches for a session where `asset_id = $1` and `ended_at IS NULL`.

---

### 4. System Diagnostics & Logging (2 endpoints)
11. **`Heartbeat(HeartbeatRequest)`** $\rightarrow$ `HeartbeatResponse`
    * **Purpose:** Check the operational health and uptime of the `telemetry-service`.
    * **Response:** Returns `"status": "healthy"` along with current RFC3339 timestamp.
12. **`LogEvent(LogEventRequest)`** $\rightarrow$ `LogEventResponse`
    * **Purpose:** Send arbitrary telemetry event logs directly through gRPC.
    * **Response:** Returns `success = true`.

---

## Implementation Layers & File References

* **Protobuf Specification:** [proto/telemetry.proto](file:///Users/elkhammamedov/Desktop/Gym-Membership-App/telemetry-service/proto/telemetry.proto)
* **Domain Layer Interface:** [internal/domain/interfaces.go](file:///Users/elkhammamedov/Desktop/Gym-Membership-App/telemetry-service/internal/domain/interfaces.go)
* **In-Memory Mock Repository:** [internal/repository/session_repo.go](file:///Users/elkhammamedov/Desktop/Gym-Membership-App/telemetry-service/internal/repository/session_repo.go)
* **Postgres Database Repository:** [internal/repository/postgres/usage_session_repo.go](file:///Users/elkhammamedov/Desktop/Gym-Membership-App/telemetry-service/internal/repository/postgres/usage_session_repo.go)
* **Business Logic Layer:** [internal/usecase/telemetry_usecase.go](file:///Users/elkhammamedov/Desktop/Gym-Membership-App/telemetry-service/internal/usecase/telemetry_usecase.go)
* **gRPC Delivery Handlers:** [internal/delivery/grpc/telemetry_handler.go](file:///Users/elkhammamedov/Desktop/Gym-Membership-App/telemetry-service/internal/delivery/grpc/telemetry_handler.go)
