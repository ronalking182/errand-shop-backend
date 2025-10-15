package main

import (
	"errandShop/config"
	"errandShop/internal/database"
	"fmt"
	"log"
)

func main() {
	log.Println("🧹 Cleaning up customer records...")

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := database.ConnectDB(cfg.DatabaseUrl)

	log.Println("✅ Database connection established")

	// Emails to target for deletion
	targetEmails := []string{"abutankokingdavid@icloud.com", "abuatnkokingdavid@gmail.com"}

	// Delete customer records that match these emails (via JOIN with users table)
	for _, email := range targetEmails {
		result := db.Exec(`
			DELETE FROM customers 
			WHERE user_id IN (
				SELECT id FROM users WHERE email = ?
			)
		`, email)
		
		if result.Error != nil {
			log.Printf("⚠️ Failed to delete customer for email %s: %v", email, result.Error)
		} else if result.RowsAffected > 0 {
			fmt.Printf("✅ Deleted %d customer record(s) for email: %s\n", result.RowsAffected, email)
		} else {
			fmt.Printf("ℹ️ No customer records found for email: %s\n", email)
		}
	}

	// Also clean up any orphaned customer records (customers without corresponding users)
	orphanResult := db.Exec(`
		DELETE FROM customers 
		WHERE user_id NOT IN (
			SELECT id FROM users
		)
	`)
	
	if orphanResult.Error != nil {
		log.Printf("⚠️ Failed to delete orphaned customers: %v", orphanResult.Error)
	} else if orphanResult.RowsAffected > 0 {
		fmt.Printf("🧹 Cleaned up %d orphaned customer record(s)\n", orphanResult.RowsAffected)
	}

	// Show all remaining customers
	fmt.Println("\n📊 All customers in database after cleanup:")
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

	err := db.Table("customers").Select("customers.*, users.email, users.name").Joins("LEFT JOIN users ON customers.user_id = users.id").Find(&customers).Error
	if err != nil {
		log.Printf("❌ Failed to query customers: %v", err)
	} else {
		fmt.Printf("\nTotal customers: %d\n", len(customers))
		for i, customer := range customers {
			fmt.Printf("%d. ID: %d, Email: %s, Name: %s %s, Phone: %s, Status: %s\n",
				i+1, customer.ID, customer.Email, customer.FirstName, customer.LastName, customer.Phone, customer.Status)
		}
	}

	log.Println("\n🎉 Customer cleanup completed!")
}