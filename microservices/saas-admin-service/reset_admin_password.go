package main

import (
    "fmt"
    "log"
    
    "golang.org/x/crypto/bcrypt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    dsn := "host=localhost user=postgres password=postgres dbname=saas_admin port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    
    // Update the admin user password
    result := db.Exec("UPDATE saas_admin_users SET password = ? WHERE username = ?", string(hashedPassword), "admin")
    if result.Error != nil {
        log.Fatal(result.Error)
    }
    
    fmt.Printf("✓ Admin password reset successfully\n")
    fmt.Printf("Username: admin\n")
    fmt.Printf("Password: admin123\n")
    fmt.Printf("Rows affected: %d\n", result.RowsAffected)
}
