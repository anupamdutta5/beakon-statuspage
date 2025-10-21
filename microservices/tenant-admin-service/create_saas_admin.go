package main

import (
    "context"
    "fmt"
    "log"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type SaaSAdminUser struct {
    ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
    Username     string    `gorm:"uniqueIndex;not null"`
    Email        string    `gorm:"uniqueIndex;not null"`
    PasswordHash string    `gorm:"not null"`
    IsActive     bool      `gorm:"default:true"`
}

func main() {
    dsn := "host=localhost user=postgres password=postgres dbname=saas_admin_db port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }

    user := &SaaSAdminUser{
        Username:     "admin",
        Email:        "admin@example.com",
        PasswordHash: string(hashedPassword),
        IsActive:     true,
    }

    if err := db.WithContext(context.Background()).Create(user).Error; err != nil {
        log.Fatal(err)
    }

    fmt.Printf("SaaS Admin user created: Username=%s, Email=%s, Password=admin123\n", user.Username, user.Email)
}
