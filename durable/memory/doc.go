// Package memory provides a synchronized, volatile outbox for development
// and tests. It is NOT durable across process restarts. Use a transactional
// database adapter in production.
package memory
