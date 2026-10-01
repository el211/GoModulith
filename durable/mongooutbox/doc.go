// Package mongooutbox implements reliable.Store using the official MongoDB Go
// driver. EnqueueSession participates in a caller-managed MongoDB transaction:
// pass the transaction's mongo.SessionContext for atomic business changes.
// MongoDB transactions require a replica set or sharded deployment.
package mongooutbox
