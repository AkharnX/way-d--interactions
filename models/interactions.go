// Like represents a user liking another user.
// @Description Like model
// @name Like
// @property id string
// @property user_id string
// @property target_id string
// @property created_at string
// @property match bool

// Dislike represents a user disliking another user.
// @Description Dislike model
// @name Dislike
// @property id string
// @property user_id string
// @property target_id string
// @property created_at string

// Match represents a match between two users.
// @Description Match model
// @name Match
// @property id string
// @property user1_id string
// @property user2_id string
// @property created_at string
// @property expire_at string

// Message represents a message between matched users.
// @Description Message model
// @name Message
// @property id string
// @property sender_id string
// @property receiver_id string
// @property content string
// @property created_at string
// @property seen bool
// @property deleted bool

// Block represents a block between users.
// @Description Block model
// @name Block
// @property id string
// @property user_id string
// @property blocked_id string
// @property reason string
// @property created_at string

package models

import (
	"time"

	"github.com/google/uuid"
)

// Like represents a user liking another user.
type Like struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TargetID  uuid.UUID `gorm:"type:uuid;not null" json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
	Match     bool      `json:"match"`
}

// Dislike represents a user disliking another user.
type Dislike struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	TargetID  uuid.UUID `gorm:"type:uuid;not null" json:"target_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Match represents a match between two users.
type Match struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	User1ID   uuid.UUID  `gorm:"type:uuid;not null" json:"user1_id"`
	User2ID   uuid.UUID  `gorm:"type:uuid;not null" json:"user2_id"`
	CreatedAt time.Time  `json:"created_at"`
	ExpireAt  *time.Time `json:"expire_at,omitempty"`
}

// Message represents a message between matched users.
type Message struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SenderID   uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	ReceiverID uuid.UUID `gorm:"type:uuid;not null" json:"receiver_id"`
	Content    string    `gorm:"type:text" json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	Seen       bool      `json:"seen"`
	Deleted    bool      `json:"deleted"`
}

// Block represents a block between users.
type Block struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	BlockedID uuid.UUID `gorm:"type:uuid;not null" json:"blocked_id"`
	Reason    string    `gorm:"type:text" json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
