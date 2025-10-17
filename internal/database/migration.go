package database

import (
	"errandShop/internal/domain/chat"
	"errandShop/internal/domain/coupons"
	"errandShop/internal/domain/customers"
	"errandShop/internal/domain/delivery"
	"errandShop/internal/domain/notifications"
	"errandShop/internal/domain/orders"
	"errandShop/internal/domain/payments"
	"errandShop/internal/domain/products"
	"errandShop/internal/pkg/models"
	"errandShop/internal/services/audit"
	"log"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type MigrationStatus struct {
	Version   int       `json:"version"`
	AppliedAt time.Time `json:"applied_at"`
	Pending   bool      `json:"pending"`
}

// Initialize all migrations
func getMigrations() []*gormigrate.Migration {
    return []*gormigrate.Migration{
		// Initial migration (v1)
		{
			ID: "0001_initial",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(
					&models.User{},
					&models.OTP{},
					&models.Address{},
					&models.RefreshToken{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
				&models.RefreshToken{},
				&models.Address{},
				&models.OTP{},
				&models.User{},
			)
			},
		},
		// Add permissions column migration
		{
			ID: "0002_add_permissions_column",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding permissions column to users table...")
				return tx.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions TEXT DEFAULT '[]'").Error
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Removing permissions column from users table...")
				return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS permissions").Error
			},
		},
		// Update existing users with default permissions
		{
			ID: "0003_update_user_permissions",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Updating existing users with default permissions...")

				// Update superadmin users
				superadminPerms := `["products:read","products:write","products:delete","orders:read","orders:write","orders:cancel","chat:read","chat:write","coupons:read","coupons:create","reports:read"]`
				if err := tx.Exec("UPDATE users SET permissions = ? WHERE role = 'superadmin' AND (permissions IS NULL OR permissions = '[]')", superadminPerms).Error; err != nil {
					return err
				}

				// Update admin users
				adminPerms := `["products:read","products:write","orders:read","orders:write","chat:read","chat:write","coupons:read"]`
				if err := tx.Exec("UPDATE users SET permissions = ? WHERE role = 'admin' AND (permissions IS NULL OR permissions = '[]')", adminPerms).Error; err != nil {
					return err
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Resetting user permissions...")
				return tx.Exec("UPDATE users SET permissions = '[]'").Error
			},
		},

		// Add new migrations below:
		// Add customers module migration
		{
			ID: "0004_add_customers_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding customers and addresses tables...")
				return tx.AutoMigrate(
					&customers.Customer{},
					&customers.Address{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping customers and addresses tables...")
				return tx.Migrator().DropTable(
					&customers.Address{},
					&customers.Customer{},
				)
			},
		},

		// Add products module migration
		{
			ID: "0005_add_products_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding products, categories, and stock_history tables...")
				
				// Check if products table exists and has data
				if tx.Migrator().HasTable(&products.Product{}) {
					log.Println("Products table exists, dropping and recreating...")
					// Drop the existing products table to avoid constraint issues
					if err := tx.Migrator().DropTable(&products.Product{}); err != nil {
						return err
					}
				}
				
				// Create all tables with proper constraints
				return tx.AutoMigrate(
					&products.Product{},
					&products.Category{},
					&products.StockHistory{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping products, categories, and stock_history tables...")
				return tx.Migrator().DropTable(
					&products.StockHistory{},
					&products.Category{},
					&products.Product{},
				)
			},
		},

		// Add orders module migration
		{
			ID: "0006_add_orders_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding orders and order_items tables...")
				
				// Check if order_items table exists and drop it to avoid type conflicts
				if tx.Migrator().HasTable(&orders.OrderItem{}) {
					log.Println("Order items table exists, dropping and recreating...")
					if err := tx.Migrator().DropTable(&orders.OrderItem{}); err != nil {
						return err
					}
				}
				
				// Check if orders table exists and drop it to avoid conflicts
				if tx.Migrator().HasTable(&orders.Order{}) {
					log.Println("Orders table exists, dropping and recreating...")
					if err := tx.Migrator().DropTable(&orders.Order{}); err != nil {
						return err
					}
				}
				
				return tx.AutoMigrate(
					&orders.Order{},
					&orders.OrderItem{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping orders and order_items tables...")
				return tx.Migrator().DropTable(
					&orders.OrderItem{},
					&orders.Order{},
				)
			},
		},
		// Add notifications migration
		{
			ID: "0009_add_notifications_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding notifications tables...")
				return tx.AutoMigrate(
					&notifications.Notification{},
					&notifications.NotificationTemplate{},
					&notifications.PushToken{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping notifications tables...")
				return tx.Migrator().DropTable(
					&notifications.PushToken{},
					&notifications.NotificationTemplate{},
					&notifications.Notification{},
				)
			},
		},
		// Add FCM tables migration
		{
			ID: "0017_add_fcm_tables",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding FCM tables...")
				return tx.AutoMigrate(
					&notifications.FCMToken{},
					&notifications.FCMMessage{},
					&notifications.FCMMessageRecipient{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping FCM tables...")
				return tx.Migrator().DropTable(
					&notifications.FCMMessageRecipient{},
					&notifications.FCMMessage{},
					&notifications.FCMToken{},
				)
			},
		},
		// Add payments import
		// "errandShop/internal/domain/payments", // Add missing comma

		// Add payments migration after orders migration
		{
			ID: "0007_add_payments_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding payments, refunds, webhooks, and orders tables...")
				return tx.AutoMigrate(
					&payments.Payment{},
					&payments.PaymentRefund{},
					&payments.PaymentWebhook{},
					&payments.Order{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping payments, refunds, webhooks, and orders tables...")
				return tx.Migrator().DropTable(
					&payments.Order{},
					&payments.PaymentWebhook{},
					&payments.PaymentRefund{},
					&payments.Payment{},
				)
			},
		},
		// Add delivery migration (simplified for third-party logistics)
		{
			ID: "0008_add_delivery_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding delivery, tracking, and zones tables...")
				return tx.AutoMigrate(
					&delivery.DeliveryZone{},
					&delivery.Delivery{},
					&delivery.TrackingUpdate{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping delivery, tracking, and zones tables...")
				return tx.Migrator().DropTable(
					&delivery.TrackingUpdate{},
					&delivery.Delivery{},
					&delivery.DeliveryZone{},
				)
			},
		},
		// Add coupons module migration
		{
			ID: "0010_add_coupons_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding coupons, coupon_usage, and user_refund_credits tables...")
				return tx.AutoMigrate(
					&coupons.Coupon{},
					&coupons.CouponUsage{},
					&coupons.UserRefundCredit{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping coupons, coupon_usage, and user_refund_credits tables...")
				return tx.Migrator().DropTable(
					&coupons.UserRefundCredit{},
					&coupons.CouponUsage{},
					&coupons.Coupon{},
				)
			},
		},
		// Add gender column to users table
		{
			ID: "0011_add_gender_column",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding gender column to users table...")
				return tx.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS gender VARCHAR(50)").Error
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Removing gender column from users table...")
				return tx.Exec("ALTER TABLE users DROP COLUMN IF EXISTS gender").Error
			},
		},
		// Add avatar column to customers table
		{
			ID: "0012_add_avatar_to_customers",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding avatar column to customers table...")
				return tx.Exec("ALTER TABLE customers ADD COLUMN IF NOT EXISTS avatar VARCHAR(500)").Error
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Removing avatar column from customers table...")
				return tx.Exec("ALTER TABLE customers DROP COLUMN IF EXISTS avatar").Error
			},
		},
		// Add idempotency_key column to orders table
		{
			ID: "0013_add_idempotency_key_to_orders",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding idempotency_key column to orders table...")
				return tx.Exec("ALTER TABLE orders ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(255)").Error
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Removing idempotency_key column from orders table...")
				return tx.Exec("ALTER TABLE orders DROP COLUMN IF EXISTS idempotency_key").Error
			},
		},
		// Fix delivery_address_id column type from UUID to integer
		{
			ID: "0014_fix_delivery_address_id_type",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Changing delivery_address_id column type from UUID to integer...")
				// Drop the column if it exists as UUID and recreate as integer
				if err := tx.Exec("ALTER TABLE orders DROP COLUMN IF EXISTS delivery_address_id").Error; err != nil {
					return err
				}
				return tx.Exec("ALTER TABLE orders ADD COLUMN delivery_address_id INTEGER").Error
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Reverting delivery_address_id column type back to UUID...")
				if err := tx.Exec("ALTER TABLE orders DROP COLUMN IF EXISTS delivery_address_id").Error; err != nil {
					return err
				}
				return tx.Exec("ALTER TABLE orders ADD COLUMN delivery_address_id UUID").Error
			},
		},
		// Add audit_logs table migration
		{
			ID: "0015_add_audit_logs_table",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding audit_logs table...")
				return tx.AutoMigrate(&audit.AuditLog{})
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping audit_logs table...")
				return tx.Migrator().DropTable(&audit.AuditLog{})
			},
		},
		// Add chat module migration
		{
			ID: "0016_add_chat_module",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Adding chat_rooms and chat_messages tables...")
				return tx.AutoMigrate(
					&chat.ChatRoom{},
					&chat.ChatMessage{},
				)
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Dropping chat_rooms and chat_messages tables...")
				return tx.Migrator().DropTable(
					&chat.ChatMessage{},
					&chat.ChatRoom{},
				)
			},
		},

		// Fix FCM system UUID type mismatch
		{
			ID: "0018_update_fcm_tables_to_uuid",
			Migrate: func(tx *gorm.DB) error {
				log.Println("Updating FCM tables to use UUID for user IDs...")
				// Update fcm_tokens table
				if err := tx.Exec("ALTER TABLE fcm_tokens ALTER COLUMN user_id TYPE UUID USING user_id::text::uuid").Error; err != nil {
					log.Printf("Warning: fcm_tokens user_id update failed: %v", err)
				}
				// Update fcm_messages table
				if err := tx.Exec("ALTER TABLE fcm_messages ALTER COLUMN sent_by TYPE UUID USING sent_by::text::uuid").Error; err != nil {
					log.Printf("Warning: fcm_messages sent_by update failed: %v", err)
				}
				// Update fcm_message_recipients table
				if err := tx.Exec("ALTER TABLE fcm_message_recipients ALTER COLUMN user_id TYPE UUID USING user_id::text::uuid").Error; err != nil {
					log.Printf("Warning: fcm_message_recipients user_id update failed: %v", err)
				}
				// Update push_tokens table
				if err := tx.Exec("ALTER TABLE push_tokens ALTER COLUMN user_id TYPE UUID USING user_id::text::uuid").Error; err != nil {
					log.Printf("Warning: push_tokens user_id update failed: %v", err)
				}
				// Update notifications table
				if err := tx.Exec("ALTER TABLE notifications ALTER COLUMN recipient_id TYPE UUID USING recipient_id::text::uuid").Error; err != nil {
					log.Printf("Warning: notifications recipient_id update failed: %v", err)
				}
				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				log.Println("Rolling back FCM UUID changes...")
				// Rollback changes (convert back to integer)
				tx.Exec("ALTER TABLE fcm_tokens ALTER COLUMN user_id TYPE INTEGER USING user_id::text::integer")
				tx.Exec("ALTER TABLE fcm_messages ALTER COLUMN sent_by TYPE INTEGER USING sent_by::text::integer")
				tx.Exec("ALTER TABLE fcm_message_recipients ALTER COLUMN user_id TYPE INTEGER USING user_id::text::integer")
				tx.Exec("ALTER TABLE push_tokens ALTER COLUMN user_id TYPE INTEGER USING user_id::text::integer")
				tx.Exec("ALTER TABLE notifications ALTER COLUMN recipient_id TYPE INTEGER USING recipient_id::text::integer")
				return nil
			},
		},

        {
            ID: "0019_fix_orders_module_create_missing_tables",
            Migrate: func(tx *gorm.DB) error {
                log.Println("Ensuring orders tables exist (non-destructive AutoMigrate)...")
                // Minimum set
                if err := tx.AutoMigrate(
                    &orders.Order{},
                    &orders.OrderItem{},
                ); err != nil {
                    return err
                }

                // Optional extras — include ONLY if these types exist in the codebase:
                // These structs exist in the codebase, so ensure their tables too
                _ = tx.AutoMigrate(&orders.Cart{})
                _ = tx.AutoMigrate(&orders.CartItem{})
                _ = tx.AutoMigrate(&orders.OrderStatusHistory{})

                return nil
            },
            Rollback: func(tx *gorm.DB) error {
                log.Println("Rollback skipped for 0019 (no destructive changes).")
                return nil
            },
        },

        // Seed superadmin: Errandshop3js@gmail.com (one-time, idempotent)
        {
            ID: "0020_seed_superadmin_errandshop3js",
            Migrate: func(tx *gorm.DB) error {
                log.Println("Seeding superadmin user: Errandshop3js@gmail.com ...")

                // Check if the user already exists
                var count int64
                if err := tx.Model(&models.User{}).
                    Where("email = ?", "Errandshop3js@gmail.com").
                    Count(&count).Error; err != nil {
                    return err
                }

                if count > 0 {
                    log.Println("Superadmin already exists, skipping seed.")
                    return nil
                }

                // Create the superadmin (password hashed by BeforeCreate hook)
                user := &models.User{
                    FirstName:   "Michelle",
                    LastName:    "Onwuaso",
                    Name:        "Michelle Onwuaso",
                    Email:       "Errandshop3js@gmail.com",
                    Password:    "Admin123!",
                    Phone:       "08144611443",
                    Role:        "superadmin",
                    Permissions: []string{"*"},
                    Status:      "active",
                    IsVerified:  true,
                    ForceReset:  false,
                    CreatedAt:   time.Now(),
                    UpdatedAt:   time.Now(),
                }

                if err := tx.Create(user).Error; err != nil {
                    return err
                }

                log.Println("✅ Superadmin seeded: Errandshop3js@gmail.com")
                return nil
            },
            Rollback: func(tx *gorm.DB) error {
                log.Println("Removing seeded superadmin: Errandshop3js@gmail.com ...")
                return tx.Where("email = ?", "Errandshop3js@gmail.com").Delete(&models.User{}).Error
            },
        },
        // Enforce case-insensitive uniqueness on email and dedupe existing
        {
            ID: "0021_enforce_unique_lower_email",
            Migrate: func(tx *gorm.DB) error {
                log.Println("🔧 Preparing to normalize emails and dedupe by LOWER(email)...")

                // Ensure gorm migrations bookkeeping table exists if queried elsewhere
                if err := tx.Exec(`CREATE TABLE IF NOT EXISTS gorm_migrations (
                    id SERIAL PRIMARY KEY,
                    migration_id TEXT UNIQUE,
                    applied_at TIMESTAMPTZ DEFAULT now()
                )`).Error; err != nil {
                    log.Printf("⚠️ Failed to ensure gorm_migrations table: %v", err)
                }

                // Ensure base tables/columns exist before custom SQL
                if err := tx.AutoMigrate(
                    &models.User{},
                    &models.OTP{},
                    &models.Address{},
                    &models.RefreshToken{},
                ); err != nil {
                    return err
                }

                // Wrap normalization, dedupe, and index creation in a transaction
                return tx.Transaction(func(tx *gorm.DB) error {
                    // Drop any legacy case-sensitive unique constraint/index on users.email
                    if err := tx.Exec("ALTER TABLE users DROP CONSTRAINT IF EXISTS uni_users_email").Error; err != nil {
                        log.Printf("⚠️ Failed to drop legacy constraint uni_users_email: %v", err)
                    }
                    if err := tx.Exec("DROP INDEX IF EXISTS uni_users_email").Error; err != nil {
                        log.Printf("⚠️ Failed to drop legacy index uni_users_email: %v", err)
                    }
                    // Common default names created by Postgres/GORM
                    if err := tx.Exec("ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key").Error; err != nil {
                        log.Printf("⚠️ Failed to drop default constraint users_email_key: %v", err)
                    }
                    if err := tx.Exec("DROP INDEX IF EXISTS idx_users_email").Error; err != nil {
                        log.Printf("⚠️ Failed to drop default index idx_users_email: %v", err)
                    }

                    // Log current duplicate groups count (case-insensitive)
                    var dupGroups int
                    countSQL := `SELECT COUNT(*) FROM (
                        SELECT LOWER(email) AS e, COUNT(*) AS c
                        FROM users
                        WHERE email IS NOT NULL
                        GROUP BY 1
                        HAVING COUNT(*) > 1
                    ) d;`
                    if err := tx.Raw(countSQL).Scan(&dupGroups).Error; err == nil {
                        if dupGroups > 0 {
                            log.Printf("🔎 Found %d case-insensitive duplicate email groups", dupGroups)
                        } else {
                            log.Println("✅ No case-insensitive duplicate email groups detected")
                        }
                    } else {
                        log.Printf("⚠️ Failed counting duplicate email groups: %v", err)
                    }

                    // Normalize emails to LOWER(TRIM(email))
                    normalizeSQL := `UPDATE users SET email = LOWER(TRIM(email)) WHERE email IS NOT NULL;`
                    if err := tx.Exec(normalizeSQL).Error; err != nil {
                        log.Printf("⚠️ Email normalization failed: %v", err)
                        // Continue; dedupe may still succeed depending on state
                    }

                    // Keep oldest row per lower(email) via DISTINCT ON
                    keepSQL := `CREATE TEMP TABLE users_keep AS
                        SELECT DISTINCT ON (LOWER(email))
                               id,
                               LOWER(email) AS email_lower
                        FROM users
                        WHERE email IS NOT NULL
                        ORDER BY LOWER(email), created_at ASC, id ASC;`
                    if err := tx.Exec(keepSQL).Error; err != nil {
                        log.Printf("⚠️ Failed to create users_keep: %v", err)
                        return err
                    }

                    // Map duplicates to keeper
                    dupsSQL := `CREATE TEMP TABLE users_dups AS
                        SELECT u.id AS dup_id,
                               k.id AS keep_id
                        FROM users u
                        JOIN users_keep k ON LOWER(u.email) = k.email_lower
                        WHERE u.id <> k.id;`
                    if err := tx.Exec(dupsSQL).Error; err != nil {
                        log.Printf("⚠️ Failed to create users_dups: %v", err)
                        return err
                    }

                    // Reassign child rows (otps.user_id) to the keeper to prevent FK violations
                    reassignSQL := `UPDATE otps o
                        SET user_id = d.keep_id
                        FROM users_dups d
                        WHERE o.user_id = d.dup_id;`
                    res := tx.Exec(reassignSQL)
                    if res.Error != nil {
                        log.Printf("⚠️ Failed to reassign OTPs to keeper users: %v", res.Error)
                        return res.Error
                    }
                    log.Printf("🔧 Reassigned %d OTP rows from duplicate users", res.RowsAffected)

                    // Optional hardening: make otps.user_id FK ON DELETE CASCADE
                    if err := tx.Exec("ALTER TABLE otps DROP CONSTRAINT IF EXISTS fk_users_ot_ps").Error; err != nil {
                        log.Printf("⚠️ Failed to drop FK fk_users_ot_ps: %v", err)
                    }
                    if err := tx.Exec("ALTER TABLE otps DROP CONSTRAINT IF EXISTS otps_user_id_fkey").Error; err != nil {
                        log.Printf("⚠️ Failed to drop default FK otps_user_id_fkey: %v", err)
                    }
                    if err := tx.Exec("ALTER TABLE otps ADD CONSTRAINT fk_users_ot_ps FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE").Error; err != nil {
                        log.Printf("⚠️ Failed to add ON DELETE CASCADE FK for otps.user_id: %v", err)
                    } else {
                        log.Println("✅ Ensured otps.user_id FK uses ON DELETE CASCADE")
                    }

                    // Delete duplicate users using mapping table
                    deleteSQL := `DELETE FROM users u
                        USING users_dups d
                        WHERE u.id = d.dup_id;`
                    if err := tx.Exec(deleteSQL).Error; err != nil {
                        log.Printf("⚠️ Failed to delete duplicate users: %v", err)
                        return err
                    }

                    // Recreate case-insensitive unique index on LOWER(email)
                    if err := tx.Exec("DROP INDEX IF EXISTS idx_users_email_lower_unique").Error; err != nil {
                        log.Printf("⚠️ Failed to drop idx_users_email_lower_unique: %v", err)
                    }
                    indexSQL := `CREATE UNIQUE INDEX idx_users_email_lower_unique ON users (LOWER(email));`
                    if err := tx.Exec(indexSQL).Error; err != nil {
                        log.Printf("⚠️ Failed to create unique index on LOWER(email): %v", err)
                        return err
                    }

                    log.Println("✅ Reassigned otps, deleted duplicates, and enforced unique LOWER(email)")
                    return nil
                })
            },
            Rollback: func(tx *gorm.DB) error {
                log.Println("↩️ Dropping unique index on lower(email)...")
                return tx.Exec("DROP INDEX IF EXISTS idx_users_email_lower_unique").Error
            },
        },
    }
}

// Run all pending migrations
func RunMigrations(db *gorm.DB) error {
	m := gormigrate.New(
		db,
		gormigrate.DefaultOptions,
		getMigrations(),
	)

	// Initialize the migration table if not exists
	m.InitSchema(func(tx *gorm.DB) error {
		log.Println("Initializing database schema...")
		// Create initial tables with proper UUID support
		return tx.AutoMigrate(
			&models.User{},
			&models.OTP{},
			&models.Address{},
			&models.RefreshToken{},
		)
	})

	return m.Migrate()
}

// Rollback the last migration
func RollbackLast(db *gorm.DB) error {
	m := gormigrate.New(
		db,
		gormigrate.DefaultOptions,
		getMigrations(),
	)
	return m.RollbackLast()
}

// Add AuditLog to migration
func addAuditLogToMigration(tx *gorm.DB) error {
	// Add to AutoMigrate (around line 30-40)
	return tx.AutoMigrate(
		&models.User{},
		&models.Address{},
		&models.RefreshToken{},
		&models.OTP{},
		&audit.AuditLog{}, // Add this
		&orders.Order{},
		&orders.OrderItem{},
	)
	return tx.Error
}

// 📋 **Implementation Steps for Each Module**

// ### **Standard Structure for Each Domain:**
// 1. **Model** (`model.go`) - Database entities
// 2. **DTOs** (`dto.go`) - Request/Response structures
// 3. **Repository** (`repository.go`) - Database operations
// 4. **Service** (`service.go`) - Business logic
// 5. **Handler** (`handler.go`) - HTTP handlers
// 6. **Routes** - Add to `main.go`

// ### **Database Migrations**
// ```go
// // Add migrations for each new table
// // Include proper indexes as specified
// // Add foreign key constraints
// ```
