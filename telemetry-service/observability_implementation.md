# Phase 6: Observability Implementation Walkthrough (+10% Bonus)

We have successfully integrated a complete, state-of-the-art Observability stack inside the `telemetry-service`. The implementation covers **Prometheus Metrics**, **OpenTelemetry Tracing**, and standard **Structured Logging** using Go's modern `log/slog` framework.

Below is an overview of the design and setup.

---

## 1. Structured Logging
We replaced Go's default plain-text logger with standard library `log/slog` outputting **JSON formats**. 

* **Default Bootstrap (`cmd/main.go`):** 
  ```go
  slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
  ```
* **Event Logging:** Every incoming NATS event is processed, validated, and logged along with critical fields like `booking_id` and `user_id` in a structured JSON manner.
  ```json
  {
    "time": "2026-05-18T13:20:00.123456+05:00",
    "level": "INFO",
    "msg": "Processing NATS event",
    "event_type": "booking.created",
    "booking_id": "booking-xyz-789",
    "user_id": "user-456",
    "asset_id": "asset-123"
  }
  ```

---

## 2. Prometheus Metrics
All Prometheus metrics are registered globally and served on an HTTP server via `http://localhost:2112/metrics`.

### Exposed Metrics

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `telemetry_events_processed_total` | `Counter` | `event_type`, `status` | Total number of NATS events processed by this microservice. |
| `telemetry_emails_sent_total` | `Counter` | `status` | Total thank-you emails sent to members. |
| `telemetry_nats_consumer_lag` | `Gauge` | `subject` | Number of pending messages in the client subscription queue. |

### Consumer Lag Monitoring
To capture **NATS consumer lag** accurately for core NATS subscriptions, we:
1. Checked client-side pending/queued messages via `nats.Subscription.Pending()`.
2. Updated the lag gauge immediately on every message processed to ensure real-time accuracy.
3. Spun up a background monitoring loop (`monitorLag()`) that updates the Prometheus gauge every 2 seconds, assuring continuous reliability even when message processing stalls.

---

## 3. OpenTelemetry (OTel) Tracing
We initialized a global **OpenTelemetry Tracer Provider** that exports spans in a formatted, highly legible style.

* **Span Creation:** For every incoming NATS event, a span is created (e.g., `NATS booking.created` or `NATS booking.returned`).
* **Trace Propagation:** The traced `context.Context` carrying the span is seamlessly propagated down to the repository and email sender layers.
* **Span Attributes & Error Recording:** Spans are enriched with semantic messaging fields:
  * `messaging.system` = `nats`
  * `messaging.destination` = subject (e.g. `booking.created`)
  * `booking_id`
  * `user_id`
  * `asset_id`
* If any error occurs (e.g. email SMTP failure or database write failure), the error is recorded on the span using `span.RecordError(err)` for visual debuggability in APM dashboards like Jaeger, Grafana Tempo, or Datadog.

---

## 4. Docker Compose & Environment Variables
The metrics HTTP server port is exposed on both the container level and docker-compose.
* **Exposed Metrics Port:** `2112`
* **Port mapping added to `docker-compose.yml`:**
  ```yaml
    ports:
      - "50053:50053"
      - "2112:2112"
    environment:
      - PORT=50053
      - METRICS_PORT=2112
      - NATS_URL=nats://nats:4222
  ```

---

## 5. Verification & Tests
All tests are compilable and run successfully!
```bash
go test ./...
# OUTPUT:
# ok      gym-membership/telemetry-service/internal/usecase       (cached)
# ok      gym-membership/telemetry-service/tests/integration      (cached)
```
