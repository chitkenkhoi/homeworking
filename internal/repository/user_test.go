package repository

import (
	"context"
	"testing"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}
func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		user := &models.User{Email: "test@example.com", Role: "TEAM_MEMBER"}
		createdUser, err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, "test@example.com", createdUser.Email)
		assert.Equal(t, "TEAM_MEMBER", string(createdUser.Role))
		assert.NotZero(t, createdUser.ID) // ID should be auto-incremented
	})

	t.Run("creation failure due to duplicate email", func(t *testing.T) {
		// Assuming Email has a unique constraint in the model
		user1 := &models.User{Email: "duplicate@example.com", Role: "user"}
		_, err := repo.Create(ctx, user1)
		assert.NoError(t, err)

		user2 := &models.User{Email: "duplicate@example.com", Role: "admin"}
		_, err = repo.Create(ctx, user2)
		assert.Error(t, err)
		assert.Equal(t, structs.ErrDataViolateConstraint, err)
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed a user
	seedUser, _ := repo.Create(ctx, &models.User{Email: "find@example.com", Role: "user"})

	t.Run("find existing user", func(t *testing.T) {
		user, err := repo.FindByID(ctx, seedUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, seedUser.ID, user.ID)
		assert.Equal(t, "find@example.com", user.Email)
	})

	t.Run("find non-existing user", func(t *testing.T) {
		user, err := repo.FindByID(ctx, 999)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, structs.ErrUserNotExist, err)
	})
}

func TestUserRepository_FindByIDs(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed users
	user1, _ := repo.Create(ctx, &models.User{Email: "user1@example.com"})
	user2, _ := repo.Create(ctx, &models.User{Email: "user2@example.com"})

	t.Run("find multiple existing users", func(t *testing.T) {
		users, err := repo.FindByIDs(ctx, []int{user1.ID, user2.ID})
		assert.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, user1.ID, users[0].ID)
		assert.Equal(t, user2.ID, users[1].ID)
	})

	t.Run("find with some non-existing IDs", func(t *testing.T) {
		users, err := repo.FindByIDs(ctx, []int{user1.ID, 999})
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, user1.ID, users[0].ID)
	})

	t.Run("empty ID list", func(t *testing.T) {
		users, err := repo.FindByIDs(ctx, []int{})
		assert.NoError(t, err)
		assert.Len(t, users, 0)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed a user
	seedUser, _ := repo.Create(ctx, &models.User{Email: "email@example.com", Role: "user"})

	t.Run("find existing email", func(t *testing.T) {
		user, err := repo.FindByEmail(ctx, "email@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, seedUser.ID, user.ID)
		assert.Equal(t, "email@example.com", user.Email)
	})

	t.Run("find non-existing email", func(t *testing.T) {
		user, err := repo.FindByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, structs.ErrUserNotExist, err)
	})
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("list with users", func(t *testing.T) {
		repo.Create(ctx, &models.User{Email: "user1@example.com"})
		repo.Create(ctx, &models.User{Email: "user2@example.com"})

		users, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed a user
	user, _ := repo.Create(ctx, &models.User{Email: "update@example.com", Role: "TEAM_MEMBER"})

	t.Run("successful update", func(t *testing.T) {
		updateMap := map[string]any{"role": "ADMIN"}
		err := repo.Update(ctx, user.ID, updateMap)
		assert.NoError(t, err)

		updatedUser, _ := repo.FindByID(ctx, user.ID)
		assert.Equal(t, "ADMIN", string(updatedUser.Role))
	})

	t.Run("update non-existing user", func(t *testing.T) {
		updateMap := map[string]any{"role": "ADMIN"}
		err := repo.Update(ctx, 999, updateMap)
		assert.Error(t, err)
		assert.Equal(t, structs.ErrUserNotExist, err)
	})

	t.Run("empty update map", func(t *testing.T) {
		err := repo.Update(ctx, user.ID, map[string]any{})
		assert.NoError(t, err) // Should succeed with no changes
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Seed a user
	user, _ := repo.Create(ctx, &models.User{Email: "delete@example.com"})

	t.Run("successful deletion", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(ctx, user.ID)
		assert.Error(t, err)
		assert.Equal(t, structs.ErrUserNotExist, err)
	})

	t.Run("delete non-existing user", func(t *testing.T) {
		err := repo.Delete(ctx, 999)
		assert.Error(t, err)
		assert.Equal(t, structs.ErrUserNotExist, err)
	})
}
