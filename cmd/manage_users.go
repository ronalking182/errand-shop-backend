package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"errandShop/internal/pkg/models"
	"fmt"
	"log"
)

func main() {
	log.Println("🔍 Managing users with specified emails...")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := database.ConnectDB(cfg.DatabaseUrl)

	log.Println("✅ Database connection established")

	// Emails to search for
	emails := []string{"abutankokingdavid@icloud.com", "abuatnkokingdavid@gmail.com"}

	// Find users with these emails
	var users []models.User
	err := db.Where("email IN ?", emails).Find(&users).Error
	if err != nil {
		log.Fatalf("❌ Failed to query users: %v", err)
	}

	fmt.Printf("\n📋 Found %d users with specified emails:\n", len(users))
	for _, user := range users {
		fmt.Printf("- ID: %s, Email: %s, Name: %s\n", user.ID, user.Email, user.Name)
	}

	if len(users) == 0 {
		fmt.Println("✅ No users found with the specified emails.")
	} else {
		// Delete customers first (foreign key constraint)
		for _, user := range users {
			// Delete customer record using raw SQL
			result := db.Exec("DELETE FROM customers WHERE user_id = ?", user.ID)
			if result.Error != nil {
				log.Printf("⚠️ Failed to delete customer for user %s: %v", user.Email, result.Error)
			} else if result.RowsAffected > 0 {
				fmt.Printf("✅ Deleted customer record for user: %s\n", user.Email)
			}

			// Delete user record
			err = db.Delete(&user).Error
			if err != nil {
				log.Printf("❌ Failed to delete user %s: %v", user.Email, err)
			} else {
				fmt.Printf("✅ Deleted user: %s\n", user.Email)
			}
		}
	}

	// Show all remaining customers
	fmt.Println("\n📊 All customers in database:")
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

	log.Println("\n🎉 User management completed!")
}