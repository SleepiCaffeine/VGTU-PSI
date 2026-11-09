package model

// Role distinguishes the two members of a pair.
type Role string

const (
	RoleGambler    Role = "gambler"
	RoleSupervisor Role = "supervisor"
)

// Tenant is one gambler and supervisor pair. All other data belongs to a tenant.
type Tenant struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"size:100;not null"`
}

type User struct {
	ID       int64  `gorm:"primaryKey"`
	TenantID int64  `gorm:"not null;index"`
	Name     string `gorm:"size:100;not null"`
	Role     Role   `gorm:"size:20;not null"`
}
