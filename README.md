# Way-d Interactions Microservice

This microservice handles all user interactions for the Way-d dating app, including likes, dislikes, matches, messaging, and blocks.

## Features
- Like and Dislike users
- Automatic match creation on mutual like
- Block users (removes all related interactions)
- Messaging between matched users (blocked users cannot message)
- JWT-protected endpoints (user_id extracted from token)
- Modular Go structure (Gin, GORM, PostgreSQL)
- OpenAPI/Swagger documentation
- Comprehensive unit/integration tests

## API Endpoints
All endpoints require a valid JWT in the `Authorization: Bearer <token>` header.

| Method | Path                  | Description                                 |
|--------|-----------------------|---------------------------------------------|
| POST   | /like                 | Like a user, triggers match on mutual like  |
| POST   | /dislike              | Dislike a user, prevents future matches     |
| GET    | /matches              | List all matches for current user           |
| POST   | /message              | Send message to a match                     |
| GET    | /messages/{match_id}  | List messages for a match                   |
| POST   | /block                | Block a user, removes all interactions      |
| GET    | /blocks               | List all users blocked by current user      |

## Business Logic
- **Like:** Creates a like, checks for reciprocal like, creates match, prevents duplicates/blocks.
- **Dislike:** Records dislike, prevents future matches.
- **Match:** Created automatically on mutual like, only active/unblocked matches are listed.
- **Message:** Only allowed if match exists and not blocked.
- **Block:** Blocks user, deletes all related likes, matches, messages, prevents further interaction.

## Setup
1. Copy `.env.example` to `.env` and set DB/JWT config.
2. Start PostgreSQL and create the main and test databases.
3. Run the service:
   ```bash
   go run main.go
   ```
4. Run tests:
   ```bash
   go test ./tests/...
   ```

## OpenAPI/Swagger Docs
- See `openapi.yaml` for the full API schema.
- You can generate Swagger UI using [swagger-ui](https://swagger.io/tools/swagger-ui/) or [swaggo/swag](https://github.com/swaggo/swag).

## Test Coverage
- All critical flows are covered:
  - Like → Match detection
  - Block and its effects
  - Message rejection for non-matched users
  - Messaging after match
  - Dislike behavior
  - Unauthorized access protection
  - Edge cases (double-like, block-then-like, dislike-then-like)

---
For questions or contributions, open an issue or PR!
