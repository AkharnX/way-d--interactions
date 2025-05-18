package controllers

import (
	"fmt"
	"net/http"
	"time"

	"way-d-interactions/config"
	"way-d-interactions/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PostLike handles liking a user and creates a match if reciprocal.
// @Summary Like a user
// @Description Like a user. If the other user has already liked you, a match is created. Cannot like yourself, users you blocked, or users who blocked you. Duplicate likes/dislikes are prevented.
// @Tags interactions
// @Accept json
// @Produce json
// @Param like body struct{target_id string} true "Target user ID"
// @Success 201 {object} models.Like
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/like [post]
func PostLike(c *gin.Context) {
	userID := c.GetString("user_id")
	var input struct {
		TargetID string `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if userID == input.TargetID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot like yourself"})
		return
	}
	db := config.GetDB()
	// Check for block
	var block models.Block
	blockErr := db.Where("(user_id = ? AND blocked_id = ?) OR (user_id = ? AND blocked_id = ?)", userID, input.TargetID, input.TargetID, userID).First(&block).Error
	fmt.Printf("[DEBUG] PostLike block check: userID=%s, targetID=%s, blockErr=%v\n", userID, input.TargetID, blockErr)
	if blockErr == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Blocked"})
		return
	}
	// Prevent duplicate like
	var existing models.Like
	if err := db.Where("user_id = ? AND target_id = ?", userID, input.TargetID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already liked"})
		return
	}
	// Prevent duplicate dislike
	var dislike models.Dislike
	if err := db.Where("user_id = ? AND target_id = ?", userID, input.TargetID).First(&dislike).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already disliked"})
		return
	}
	like := models.Like{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		TargetID:  uuid.MustParse(input.TargetID),
		CreatedAt: time.Now(),
	}
	// Check for reciprocal like and create match if needed
	var reciprocal models.Like
	if err := db.Where("user_id = ? AND target_id = ?", input.TargetID, userID).First(&reciprocal).Error; err == nil {
		like.Match = true
		reciprocal.Match = true
		db.Save(&reciprocal)
		// Create match
		match := models.Match{
			ID:        uuid.New(),
			User1ID:   like.UserID,
			User2ID:   like.TargetID,
			CreatedAt: time.Now(),
		}
		db.Create(&match)
	}
	db.Create(&like)
	c.JSON(http.StatusCreated, like)

	// DEBUG: Print userID and input.TargetID for troubleshooting
	fmt.Printf("[DEBUG] PostLike: userID=%s, targetID=%s\n", userID, input.TargetID)
}

// POST /dislike
// @Summary Dislike a user
// @Description Dislike a user. Cannot dislike yourself, users you blocked, or users who blocked you. Duplicate dislikes/likes are prevented.
// @Tags interactions
// @Accept json
// @Produce json
// @Param dislike body struct{target_id string} true "Target user ID"
// @Success 201 {object} models.Dislike
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/dislike [post]
func PostDislike(c *gin.Context) {
	userID := c.GetString("user_id")
	var input struct {
		TargetID string `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if userID == input.TargetID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot dislike yourself"})
		return
	}
	db := config.GetDB()
	// Check for block
	var block models.Block
	if err := db.Where("(user_id = ? AND blocked_id = ?) OR (user_id = ? AND blocked_id = ?)", userID, input.TargetID, input.TargetID, userID).First(&block).Error; err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Blocked"})
		return
	}
	// Prevent duplicate dislike
	var existing models.Dislike
	if err := db.Where("user_id = ? AND target_id = ?", userID, input.TargetID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already disliked"})
		return
	}
	// Prevent duplicate like
	var like models.Like
	if err := db.Where("user_id = ? AND target_id = ?", userID, input.TargetID).First(&like).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already liked"})
		return
	}
	dislike := models.Dislike{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		TargetID:  uuid.MustParse(input.TargetID),
		CreatedAt: time.Now(),
	}
	db.Create(&dislike)
	c.JSON(http.StatusCreated, dislike)
}

// GET /matches
// @Summary List matches
// @Description Get all matches for the current user.
// @Tags interactions
// @Produce json
// @Success 200 {array} models.Match
// @Router /api/matches [get]
func GetMatches(c *gin.Context) {
	userID := c.GetString("user_id")
	var matches []models.Match
	db := config.GetDB()
	db.Where("user1_id = ? OR user2_id = ?", userID, userID).Find(&matches)
	c.JSON(http.StatusOK, matches)
}

// POST /message
// @Summary Send message
// @Description Send a message to a matched user. Blocked users cannot send/receive messages.
// @Tags interactions
// @Accept json
// @Produce json
// @Param message body struct{match_id string; content string} true "Message content and match ID"
// @Success 201 {object} models.Message
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/message [post]
func PostMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	var input struct {
		MatchID string `json:"match_id" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Check match exists and user is part of it
	var match models.Match
	db := config.GetDB()
	if err := db.Where("id = ? AND (user1_id = ? OR user2_id = ?)", input.MatchID, userID, userID).First(&match).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No such match or not a participant"})
		return
	}
	// Check for block between users
	var block models.Block
	var otherID string
	if match.User1ID.String() == userID {
		otherID = match.User2ID.String()
	} else {
		otherID = match.User1ID.String()
	}
	if err := db.Where("(user_id = ? AND blocked_id = ?) OR (user_id = ? AND blocked_id = ?)", userID, otherID, otherID, userID).First(&block).Error; err == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Blocked"})
		return
	}
	msg := models.Message{
		ID:         uuid.New(),
		SenderID:   uuid.MustParse(userID),
		ReceiverID: uuid.MustParse(otherID),
		Content:    input.Content,
		CreatedAt:  time.Now(),
		Seen:       false,
		Deleted:    false,
	}
	db.Create(&msg)
	c.JSON(http.StatusCreated, msg)
}

// GET /messages/:match_id
// @Summary List messages
// @Description Get all messages for a match (must be a participant).
// @Tags interactions
// @Produce json
// @Param match_id path string true "Match ID"
// @Success 200 {array} models.Message
// @Failure 403 {object} map[string]string
// @Router /api/messages/{match_id} [get]
func GetMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	matchID := c.Param("match_id")
	var match models.Match
	if err := config.DB.Where("id = ? AND (user1_id = ? OR user2_id = ?)", matchID, userID, userID).First(&match).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "No such match or not a participant"})
		return
	}
	var messages []models.Message
	db := config.GetDB()
	db.Where(
		"((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND deleted = false",
		match.User1ID, match.User2ID, match.User2ID, match.User1ID,
	).Order("created_at asc").Find(&messages)
	c.JSON(http.StatusOK, messages)
}

// POST /block
// @Summary Block a user
// @Description Block a user. Cleans up likes, dislikes, matches, and messages between users. Cannot block yourself or block twice.
// @Tags interactions
// @Accept json
// @Produce json
// @Param block body struct{blocked_id string} true "Blocked user ID"
// @Success 201 {object} models.Block
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/block [post]
func PostBlock(c *gin.Context) {
	userID := c.GetString("user_id")
	var input struct {
		BlockedID string `json:"blocked_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if userID == input.BlockedID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block yourself"})
		return
	}
	db := config.GetDB()
	// Prevent duplicate block
	var existing models.Block
	if err := db.Where("user_id = ? AND blocked_id = ?", userID, input.BlockedID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already blocked"})
		return
	}
	block := models.Block{
		ID:        uuid.New(),
		UserID:    uuid.MustParse(userID),
		BlockedID: uuid.MustParse(input.BlockedID),
		CreatedAt: time.Now(),
	}
	db.Create(&block)
	// Cleanup: delete likes, dislikes, matches, messages between users
	db.Where("(user_id = ? AND target_id = ?) OR (user_id = ? AND target_id = ?)", userID, input.BlockedID, input.BlockedID, userID).Delete(&models.Like{})
	db.Where("(user_id = ? AND target_id = ?) OR (user_id = ? AND target_id = ?)", userID, input.BlockedID, input.BlockedID, userID).Delete(&models.Dislike{})
	db.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", userID, input.BlockedID, input.BlockedID, userID).Delete(&models.Match{})
	db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, input.BlockedID, input.BlockedID, userID).Delete(&models.Message{})
	c.JSON(http.StatusCreated, block)
}

// GET /blocks
// @Summary List blocks
// @Description Get all users blocked by the current user.
// @Tags interactions
// @Produce json
// @Success 200 {array} models.Block
// @Router /api/blocks [get]
func GetBlocks(c *gin.Context) {
	userID := c.GetString("user_id")
	var blocks []models.Block
	db := config.GetDB()
	db.Where("user_id = ?", userID).Find(&blocks)
	c.JSON(http.StatusOK, blocks)
}
