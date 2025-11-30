package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user account in the system
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	TenantID     int       `gorm:"not null;index:idx_users_tenant_email" json:"tenant_id"`
	Email        string    `gorm:"type:text;not null;index:idx_users_tenant_email" json:"email"`
	PasswordHash string    `gorm:"type:text;not null;column:password_hash" json:"-"`
	Role         string    `gorm:"type:text;not null;check:role IN ('admin', 'editor', 'viewer')" json:"role"`
	IsActive     bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt    time.Time `gorm:"default:now()" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:now()" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "yt_users"
}

// BeforeCreate hook to generate UUID if not set
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// HashPassword hashes a plaintext password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword verifies a plaintext password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// CreateUser creates a new user in the database
func CreateUser(db *gorm.DB, tenantID int, email, password, role string) (*User, error) {
	// Validate role
	if role != "admin" && role != "editor" && role != "viewer" {
		return nil, errors.New("invalid role: must be admin, editor, or viewer")
	}

	// Hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &User{
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		IsActive:     true,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByEmailTenant retrieves a user by email and tenant ID
func GetUserByEmailTenant(db *gorm.DB, tenantID int, email string) (*User, error) {
	var user User
	err := db.Where("tenant_id = ? AND email = ? AND is_active = ?", tenantID, email, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(db *gorm.DB, userID uuid.UUID) (*User, error) {
	var user User
	err := db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsersByTenant retrieves all active users for a tenant
func ListUsersByTenant(db *gorm.DB, tenantID int, limit, offset int) ([]User, int64, error) {
	var users []User
	var total int64

	// Count total
	if err := db.Model(&User{}).Where("tenant_id = ? AND is_active = ?", tenantID, true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch users
	err := db.Where("tenant_id = ? AND is_active = ?", tenantID, true).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

// UpdateUser updates user fields (excluding password)
func UpdateUser(db *gorm.DB, userID uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(db *gorm.DB, userID uuid.UUID, newPassword string) error {
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return db.Model(&User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}

// DeactivateUser soft-deletes a user by setting is_active to false
func DeactivateUser(db *gorm.DB, userID uuid.UUID) error {
	return db.Model(&User{}).Where("id = ?", userID).Update("is_active", false).Error
}

// GetUserCountByTenant returns the count of active users for a tenant
func GetUserCountByTenant(db *gorm.DB, tenantID int) (int64, error) {
	var count int64
	err := db.Model(&User{}).Where("tenant_id = ? AND is_active = ?", tenantID, true).Count(&count).Error
	return count, err
}

// GetFirstUserForTenant checks if a user is the first user for a tenant
func GetFirstUserForTenant(db *gorm.DB, tenantID int) (bool, error) {
	var count int64
	err := db.Model(&User{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

// UserRepository provides database operations for users using sql.DB (for compatibility)
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByEmailTenantSQL retrieves a user using raw SQL (for compatibility with existing code)
func (r *UserRepository) GetUserByEmailTenantSQL(tenantID int, email string) (*User, error) {
	var user User
	err := r.db.QueryRow(`
		SELECT id, tenant_id, email, password_hash, role, is_active, created_at, updated_at
		FROM yt_users
		WHERE tenant_id = $1 AND email = $2 AND is_active = true
	`, tenantID, email).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

