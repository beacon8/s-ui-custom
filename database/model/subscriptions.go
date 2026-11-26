package model

// SubscriptionGroup is the independent identity behind one import batch's
// group subscription link. It is deliberately separate from ImportBatch: a
// new batch that happens to reuse the same visible group name gets a new row
// and can never inherit the old group's subscription authority.
type SubscriptionGroup struct {
	Id        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	SubID     string `json:"subscriptionId" gorm:"uniqueIndex"`
	Name      string `json:"name"`
	BatchId   string `json:"batchId" gorm:"index"`
	TokenEnc  string `json:"-"` // AES-GCM, key derived from the panel secret
	Revoked   bool   `json:"revoked" gorm:"default:false;not null"`
	CreatedAt int64  `json:"createdAt" gorm:"default:0;not null"`
}

// SubscriptionMember ties one stable client id to one group and carries that
// client's personal subscription identity. ClientId is unique: a client can
// belong to only one group, and a deleted-then-recreated client (new id) does
// not inherit the old member row.
type SubscriptionMember struct {
	Id        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupId   uint   `json:"groupId" gorm:"index"`
	ClientId  uint   `json:"clientId" gorm:"uniqueIndex"`
	SubID     string `json:"subscriptionId" gorm:"uniqueIndex"`
	TokenEnc  string `json:"-"`
	Revoked   bool   `json:"revoked" gorm:"default:false;not null"`
	CreatedAt int64  `json:"createdAt" gorm:"default:0;not null"`
}
