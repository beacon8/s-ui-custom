package model

// ImportBatch records one importClientNodes request. It is written in the same
// transaction as the clients, inbounds, outbounds and route rules, so an
// idempotency key can only ever refer to data that actually committed.
//
// IdempotencyKey is a generated placeholder when the caller sent no key: the
// unique index then still separates every keyless batch without rejecting it.
type ImportBatch struct {
	Id          uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	BatchId     string `json:"batchId" gorm:"uniqueIndex"`
	Actor       string `json:"actor" gorm:"uniqueIndex:idx_import_actor_key"`
	IdemKey     string `json:"idemKey" gorm:"uniqueIndex:idx_import_actor_key"`
	RequestHash string `json:"requestHash"`
	Group       string `json:"group"`
	Count       int    `json:"count"`
	// Status is committed, applied or apply_failed. A committed batch whose
	// core application never finished is what the reapply endpoint recovers.
	Status string `json:"status"`
	Error  string `json:"error"`
	// TEXT rather than json.RawMessage: SQLite hands TEXT back as a string,
	// which database/sql cannot scan into a named []byte type.
	Request   string `json:"request"`
	Result    string `json:"result"`
	CreatedAt int64  `json:"createdAt"`
	AppliedAt int64  `json:"appliedAt"`
}
