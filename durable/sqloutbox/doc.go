// Package sqloutbox implements the reliable outbox over database/sql.
// Supports SQLite and PostgreSQL using caller-supplied SQL drivers.
// Run Schema for the chosen dialect before use. EnqueueTx MUST be called with
// the same *sql.Tx as the business mutation to achieve transaction-bound
// enqueue. Claim uses conditional UPDATE compare-and-swap; leases can expire
// and delivery remains at-least-once, so consumers must deduplicate IDs.
package sqloutbox
