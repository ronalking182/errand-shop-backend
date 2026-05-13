package database

import (
	"errors"
	"fmt"
	"log"

	cr "errandShop/internal/domain/custom_requests"
	"gorm.io/gorm"
)

// CustomRequestsPhysicalTablePresent checks via GORM, then falls back to information_schema (public.custom_requests).
func CustomRequestsPhysicalTablePresent(db *gorm.DB) (bool, error) {
	if db.Migrator().HasTable(&cr.CustomRequest{}) {
		return true, nil
	}
	var count int64
	err := db.Raw(`
		SELECT COUNT(*)
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'custom_requests'
	`).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// EnsureCustomRequestsStack runs AutoMigrate for the custom-requests domain only, in FK-safe order with logs.
func EnsureCustomRequestsStack(db *gorm.DB) error {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		log.Printf("custom_requests: CREATE EXTENSION pgcrypto skipped or failed (non-fatal): %v", err)
	}

	for _, step := range []struct {
		name string
		fn   func() error
	}{
		{"custom_requests.CustomRequest", func() error { return db.AutoMigrate(&cr.CustomRequest{}) }},
		{"custom_requests.RequestItem", func() error { return db.AutoMigrate(&cr.RequestItem{}) }},
		{"custom_requests.Quote", func() error { return db.AutoMigrate(&cr.Quote{}) }},
		{"custom_requests.QuoteItem", func() error { return db.AutoMigrate(&cr.QuoteItem{}) }},
		{"custom_requests.CustomRequestMessage", func() error { return db.AutoMigrate(&cr.CustomRequestMessage{}) }},
	} {
		log.Printf("custom_requests: AutoMigrate %s", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
	}
	return nil
}

// RepairCustomRequestsIfMissing verifies public.custom_requests; if absent, rebuilds stack and fatals if still missing.
func RepairCustomRequestsIfMissing(db *gorm.DB) error {
	ok, err := CustomRequestsPhysicalTablePresent(db)
	if err != nil {
		log.Printf("custom_requests verify: scan error: %v", err)
		return err
	}
	if ok {
		log.Printf("✅ verify: table public.custom_requests exists")
		return nil
	}
	log.Printf("⚠️ verify: table public.custom_requests MISSING — forcing EnsureCustomRequestsStack()")
	if err := EnsureCustomRequestsStack(db); err != nil {
		return err
	}
	ok2, err := CustomRequestsPhysicalTablePresent(db)
	if err != nil {
		return err
	}
	if !ok2 {
		return errors.New("custom_requests repair failed: public.custom_requests still missing after EnsureCustomRequestsStack (check Postgres role/permissions or deploy latest binary)")
	}
	log.Printf("✅ repair: public.custom_requests created successfully")
	return nil
}
