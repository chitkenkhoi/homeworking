package migration

import (
	"log"
	"fmt"
	
	"lqkhoi-go-http-api/internal/models"

	"gorm.io/gorm"
)
func createEnumUserRole(tx *gorm.DB)error{
	log.Println("Ensuring ENUM type 'user_role' exists...")
	sqlUserRoleSafe := `
	DO $$
	BEGIN
	    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
	        CREATE TYPE user_role AS ENUM ('ADMIN', 'PROJECT_MANAGER', 'TEAM_MEMBER');
	    END IF;
	END$$;
	`
	if err := tx.Exec(sqlUserRoleSafe).Error; err != nil {
		log.Printf("Error creating/ensuring ENUM type 'user_role': %v\n", err)
		return fmt.Errorf("failed to ensure enum 'user_role': %w", err)
	}
	log.Println("'user_role' ENUM type checked/created.")
	return nil
}

func createEnumProjectStatus(tx *gorm.DB)error{
	log.Println("Ensuring ENUM type 'project_status' exists...")
	sqlProjectStatusSafe := `
	DO $$
	BEGIN
	    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'project_status') THEN
	        CREATE TYPE project_status AS ENUM ('ACTIVE', 'COMPLETED', 'ON_HOLD', 'CANCELLED');
	    END IF;
	END$$;
	`
	if err := tx.Exec(sqlProjectStatusSafe).Error; err != nil {
		log.Printf("Error creating/ensuring ENUM type 'project_status': %v\n", err)
		return fmt.Errorf("failed to ensure enum 'project_status': %w", err)
	}
	log.Println("'project_status' ENUM type checked/created.")
	return nil
}

func createEnumTaskStatus(tx *gorm.DB)error{
	log.Println("Ensuring ENUM type 'task_status' exists...")
	sqlProjectStatusSafe := `
	DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
            CREATE TYPE task_status AS ENUM ('TO_DO', 'IN_PROGRESS', 'REVIEW', 'DONE', 'BLOCKED');
        END IF;
    END$$;
	`
	if err := tx.Exec(sqlProjectStatusSafe).Error; err != nil {
		log.Printf("Error creating/ensuring ENUM type 'task_status': %v\n", err)
		return fmt.Errorf("failed to ensure enum 'task_status': %w", err)
	}
	log.Println("'task_status' ENUM type checked/created.")
	return nil
}

func createEnumTaskPriority(tx *gorm.DB)error{
	log.Println("Ensuring ENUM type 'task_priority' exists...")
	sqlProjectStatusSafe := `
	DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_priority') THEN
            CREATE TYPE task_priority AS ENUM ('HIGH', 'MEDIUM', 'LOW', 'CRITICAL');
        END IF;
    END$$;

	`
	if err := tx.Exec(sqlProjectStatusSafe).Error; err != nil {
		log.Printf("Error creating/ensuring ENUM type 'task_priority': %v\n", err)
		return fmt.Errorf("failed to ensure enum 'task_priority': %w", err)
	}
	log.Println("'task_priority' ENUM type checked/created.")
	return nil
}

func createTables(tx *gorm.DB)error{
	log.Println("Running GORM AutoMigrate for creating tables...")
	if err := tx.AutoMigrate(&models.User{}, &models.Project{}, &models.Task{}, &models.Sprint{}); err != nil {
		log.Printf("Error during GORM AutoMigrate: %v\n", err)
		return fmt.Errorf("gorm automigrate failed: %w", err)
	}
	log.Println("GORM AutoMigrate completed.")
	return nil
}

func AutoMigrate(db *gorm.DB)error{
	log.Println("Starting database migration...")

	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v\n", tx.Error)
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic during migration, rolling back transaction:", r)
			tx.Rollback()
			panic(r) 
		} else if err := tx.Error; err != nil {
			log.Println("Rolling back transaction due to error:", err)
			tx.Rollback()
		}
	}()

	if err := createEnumUserRole(tx);err != nil{
		return err
	}

	if err := createEnumProjectStatus(tx);err != nil{
		return err
	}

	if err := createEnumTaskStatus(tx);err != nil{
		return err
	}

	if err := createEnumTaskPriority(tx);err != nil{
		return err
	}

	if err := createTables(tx);err!=nil{
		return err
	}
	log.Println("Committing transaction...")
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v\n", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Database migration completed successfully.")
	return nil
}