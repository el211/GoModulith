// Package events implements a concurrency-safe, in-process publish/subscribe
// bus. Publish delivers to a snapshot of handlers in subscription order. Each
// subscriber can be unsubscribed and can propagate an error. It is intentionally
// transport independent; see durable for persistent at-least-once delivery.
package events
