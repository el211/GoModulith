// Package durable defines an outbox abstraction for durable application
// events. Persistence adapters implement Store, while Dispatcher provides
// at-least-once delivery. Consumers must be idempotent: delivery may repeat
// after a crash between publishing and acknowledgement.
package durable
