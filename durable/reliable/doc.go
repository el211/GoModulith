// Package reliable implements retryable at-least-once delivery with leases,
// exponential backoff and dead letters. A distributed Store must atomically
// claim eligible messages, fencing competing workers until lease expiration.
// Publish may succeed before Ack fails: all consumers must be idempotent.
package reliable
