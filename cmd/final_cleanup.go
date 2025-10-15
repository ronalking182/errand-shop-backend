package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"fmt"
	"log"
)

func main() {
	log.Println("🎯 Final cleanup for remaining target email...")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := database.ConnectDB(cfg.DatabaseUrl)

	log.Println("✅ Database connection established")

	// Target the remaining email
	targetEmail := "abutankokingdavid@gmail.com"

	// First, find the user
	var user struct {
		ID    string `gorm:"column:id"`
		Email string `gorm:"column:email"`
		Name  string `gorm:"column:name"`
	}

	err := db.Table("users").Where("email = ?", targetEmail).First(&user).Error
	if err != nil {
		log.Printf("ℹ️ No user found with email: %s", targetEmail)
	} else {
		fmt.Printf("📋 Found user: ID: %s, Email: %s, Name: %s\n", user.ID, user.Email, user.Name)

		// Delete customer record first
		customerResult := db.Exec("DELETE FROM customers WHERE user_id = ?", user.ID)
		if customerResult.Error != nil {
			log.Printf("⚠️ Failed to delete customer for user %s: %v", user.Email, customerResult.Error)
		} else if customerResult.RowsAffected > 0 {
			fmt.Printf("✅ Deleted customer record for user: %s\n", user.Email)
		}

		// Delete user record
		userResult := db.Exec("DELETE FROM users WHERE id = ?", user.ID)
		if userResult.Error != nil {
			log.Printf("❌ Failed to delete user %s: %v", user.Email, userResult.Error)
		} else {
			fmt.Printf("✅ Deleted user: %s\n", user.Email)
		}
	}

	// Show final customer list
	fmt.Println("\n📊 Final customer list:")
	var customers []struct {
		ID        uint   `gorm:"primaryKey"`
		UserID    string `gorm:"column:user_id"`
		FirstName string `gorm:"column:first_name"`
		LastName  string `gorm:"column:last_name"`
		Phone     string `gorm:"column:phone"`
		Status    string `gorm:"column:status"`
		Email     string `gorm:"column:email"`
		Name      string `gorm:"column:name"`
	}

	err = db.Table("customers").Select("customers.*, users.email, users.name").Joins("LEFT JOIN users ON customers.user_id = users.id").Find(&customers).Error
	if err != nil {
		log.Printf("❌ Failed to query customers: %v", err)
	} else {
		fmt.Printf("\nTotal customers: %d\n", len(customers))
		for i, customer := range customers {
			fmt.Printf("%d. ID: %d, Email: %s, Name: %s %s, Phone: %s, Status: %s\n",
				i+1, customer.ID, customer.Email, customer.FirstName, customer.LastName, customer.Phone, customer.Status)
		}
	}

	log.Println("\n🎉 Final cleanup completed!")
}