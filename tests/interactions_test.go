// Unit tests for like, match, block, and message restrictions in the interactions service.

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"way-d-interactions/config"
	"way-d-interactions/models"
	"way-d-interactions/routes"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateTestJWT returns a valid JWT for the test user with the correct secret and claims.
func GenerateTestJWT(userID string) string {
	secret := os.Getenv("JWT_SECRET")
	type testClaims struct {
		UserID string `json:"user_id"`
		jwt.RegisteredClaims
	}
	claims := testClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	routes.RegisterRoutes(r)
	return r
}

func setupTestDB() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "wayd_user")
	os.Setenv("DB_PASSWORD", "test")
	os.Setenv("DB_NAME", "wayd_interactions_test")
	os.Setenv("JWT_SECRET", "e5b9922f19cf240b093a3e851f905bce71d8444b44c13d616c9c58bf2cbb8b78")
	config.ConnectDB()
	db := config.GetDB()
	db.Migrator().DropTable(&models.Like{}, &models.Dislike{}, &models.Match{}, &models.Message{}, &models.Block{})
	db.AutoMigrate(&models.Like{}, &models.Dislike{}, &models.Match{}, &models.Message{}, &models.Block{})
}

func TestLikeAndMatch(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	body := `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Errorf("Like failed: %d %s", w.Code, w.Body.String())
	}
}

func TestBlockAndCleanup(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	body := `{"blocked_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/block", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Errorf("Block failed: %d %s", w.Code, w.Body.String())
	}
}

func TestMessageRestriction(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	body := `{"match_id": "11111111-1111-1111-1111-111111111111", "content": "hi"}`
	req, _ := http.NewRequest("POST", "/api/message", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated && w.Code != http.StatusForbidden {
		t.Errorf("Message restriction failed: %d %s", w.Code, w.Body.String())
	}
}

func TestDislikePreventsMatch(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	// Dislike a user
	body := `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/dislike", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Dislike failed: %d %s", w.Code, w.Body.String())
	}
	// Try to like after dislike (should not create match)
	body = `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ = http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Errorf("Like after dislike should not be allowed: %d %s", w.Code, w.Body.String())
	}
}

func TestGetMatches(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	// Like user 2
	body := `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Like back as user 2
	jwt2 := GenerateTestJWT("11111111-1111-1111-1111-111111111111")
	body = `{"target_id": "00000000-0000-0000-0000-000000000001"}`
	req, _ = http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt2)
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	// Get matches for user 1
	req, _ = http.NewRequest("GET", "/api/matches", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Get matches failed: %d %s", w.Code, w.Body.String())
	}
	var matches []map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &matches)
	if len(matches) == 0 {
		t.Errorf("Expected at least one match, got 0")
	}
}

func TestGetBlocks(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	// Block a user
	body := `{"blocked_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/block", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Get blocks
	req, _ = http.NewRequest("GET", "/api/blocks", nil)
	req.Header.Set("Authorization", "Bearer "+jwt)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Get blocks failed: %d %s", w.Code, w.Body.String())
	}
	var blocks []map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &blocks)
	if len(blocks) == 0 {
		t.Errorf("Expected at least one block, got 0")
	}
}

func TestMessagingAfterMatch(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt1 := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	jwt2 := GenerateTestJWT("11111111-1111-1111-1111-111111111111")
	// Like each other
	body := `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt1)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	body = `{"target_id": "00000000-0000-0000-0000-000000000001"}`
	req, _ = http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt2)
	req.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	// Get matches for user 1 to find match_id
	req, _ = http.NewRequest("GET", "/api/matches", nil)
	req.Header.Set("Authorization", "Bearer "+jwt1)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var matches []map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &matches)
	if len(matches) == 0 {
		t.Fatalf("No match found for messaging test")
	}
	matchID, ok := matches[0]["id"].(string)
	if !ok {
		t.Fatalf("Match ID not found in response")
	}
	// Send message as user 1
	msgBody := fmt.Sprintf(`{"match_id": "%s", "content": "Hello!"}`, matchID)
	req, _ = http.NewRequest("POST", "/api/message", bytes.NewBufferString(msgBody))
	req.Header.Set("Authorization", "Bearer "+jwt1)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Message after match failed: %d %s", w.Code, w.Body.String())
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	body := `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	// No Authorization header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Errorf("Unauthorized like should be rejected: %d %s", w.Code, w.Body.String())
	}
}

func TestDoubleLikeAndBlockEdgeCases(t *testing.T) {
	setupTestDB()
	r := setupRouter()
	jwt := GenerateTestJWT("00000000-0000-0000-0000-000000000001")
	// Like user 2
	body := `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ := http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Like again (should not duplicate)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	if w2.Code == http.StatusCreated {
		t.Errorf("Double like should not be allowed: %d %s", w2.Code, w2.Body.String())
	}
	// Block user 2
	body = `{"blocked_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ = http.NewRequest("POST", "/api/block", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Try to like again after block (should be forbidden)
	body = `{"target_id": "11111111-1111-1111-1111-111111111111"}`
	req, _ = http.NewRequest("POST", "/api/like", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusCreated {
		t.Errorf("Like after block should not be allowed: %d %s", w.Code, w.Body.String())
	}
}
