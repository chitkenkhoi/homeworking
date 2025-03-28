package migration

import (
	"fmt"
	"log"

	"lqkhoi-go-http-api/internal/models"

	"gorm.io/gorm"
)

type constraintDefinition struct {
	Model any
	RelationField string
	ConstraintName string
	Description string
}

func createEnumUserRole(tx *gorm.DB) error {
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

func createEnumProjectStatus(tx *gorm.DB) error {
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

func createEnumTaskStatus(tx *gorm.DB) error {
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

func createEnumTaskPriority(tx *gorm.DB) error {
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

func createTables(tx *gorm.DB) error {
	log.Println("Running GORM AutoMigrate for creating tables...")

	modelsToMigrate := []any{
		&models.User{},
		&models.Project{},
		&models.Sprint{},
		&models.Task{},
	}

	for _, model := range modelsToMigrate {
		log.Printf("AutoMigrating %T...", model)
		if err := tx.AutoMigrate(model); err != nil {
			log.Printf("Error during GORM AutoMigrate for %T: %v\n", model, err)
			return fmt.Errorf("gorm automigrate failed for %T: %w", model, err)
		}
	}

	log.Println("GORM AutoMigrate for tables completed.")
	return nil
}

func createForeignKeyTranSaction(tx *gorm.DB) error {
	log.Println("Manually adding foreign key constraints...")

	migrator := tx.Migrator()
	constraints := []constraintDefinition{
		{ // 1. Project.ManagerID -> users.id
			Model:          &models.User{},      
			RelationField:  "ManagedProjects",    
			ConstraintName: "fk_users_managed_projects", 
			Description:    "projects.manager_id -> users.id",
		},
		{ // 2. Sprint.ProjectID -> projects.id
			Model:          &models.Project{},    
			RelationField:  "Sprints",            
			ConstraintName: "fk_projects_sprints",
			Description:    "sprints.project_id -> projects.id",
		},
		{ // 3. Task.ProjectID -> projects.id
			Model:          &models.Project{},   
			RelationField:  "Tasks",             
			ConstraintName: "fk_projects_tasks",
			Description:    "tasks.project_id -> projects.id",
		},
		{ // 4. Task.AssigneeID -> users.id (Nullable)
			Model:          &models.User{},       
			RelationField:  "AssignedTasks",      
			ConstraintName: "fk_users_assigned_tasks",
			Description:    "tasks.assignee_id -> users.id",
		},
		{ // 5. Task.SprintID -> sprints.id (Nullable)
			Model:          &models.Sprint{},     
			RelationField:  "Tasks",              
			ConstraintName: "fk_sprints_tasks",    
			Description:    "tasks.sprint_id -> sprints.id",
		},
		{ // 6. User.CurrentProjectID -> projects.id (Nullable)
			Model:          &models.Project{},    
			RelationField:  "TeamMembers",        
			ConstraintName: "fk_projects_team_members",
			Description:    "users.current_project_id -> projects.id",
		},
	}
	for _, c := range constraints {
		log.Printf("Processing constraint: %s", c.Description)

		if !migrator.HasConstraint(c.Model, c.ConstraintName) {
			log.Printf("Constraint %q does not exist, creating relation %q on %T...", c.ConstraintName, c.RelationField, c.Model)

			err := migrator.CreateConstraint(c.Model, c.RelationField)
			if err != nil {
				log.Printf("Error creating constraint for %T.%s (%s): %v", c.Model, c.RelationField, c.Description, err)
				return fmt.Errorf("failed to create constraint for %T.%s (%s): %w", c.Model, c.RelationField, c.Description, err)
			}
			log.Printf("Successfully created constraint for %s", c.Description) 
		} else {
			log.Printf("Constraint %q already exists.", c.ConstraintName)
		}
	}

	log.Println("Foreign key constraints added successfully (or already existed).")
	return nil
}

func autoMigrateWithoutCreateFkConstraint(db *gorm.DB) error {
	log.Println("Starting database migration...")

	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v\n", tx.Error)
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Use a variable to track errors and decide whether to commit or rollback
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic during migration, rolling back transaction:", r)
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Rolling back transaction due to error:", err)
			tx.Rollback()
		} else {
			log.Println("Committing transaction...")
			commitErr := tx.Commit().Error
			if commitErr != nil {
				log.Printf("Error committing transaction: %v\n", commitErr)
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	if err = createEnumUserRole(tx); err != nil {
		return err // Return immediately on error
	}

	if err = createEnumProjectStatus(tx); err != nil {
		return err // Return immediately on error
	}

	if err = createEnumTaskStatus(tx); err != nil {
		return err // Return immediately on error
	}

	if err = createEnumTaskPriority(tx); err != nil {
		return err // Return immediately on error
	}

	if err = createTables(tx); err != nil {
		return err // Return immediately on error
	}

	log.Println("Database migration completed successfully.")
	return err
}

func createFkConstraint(db *gorm.DB) error {
	log.Println("Starting foreign key creation...")
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v\n", tx.Error)
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic during migration, rolling back transaction:", r)
			tx.Rollback()
			panic(r)
		} else if err != nil {
			log.Println("Rolling back transaction due to error:", err)
			tx.Rollback()
		} else {
			log.Println("Committing transaction...")
			commitErr := tx.Commit().Error
			if commitErr != nil {
				log.Printf("Error committing transaction: %v\n", commitErr)
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()
	if err = createForeignKeyTranSaction(tx);err != nil{
		return err
	}
	log.Println("Database migration completed successfully.")
	return nil
}

func AutoMigrate(db *gorm.DB) error {
	if err := autoMigrateWithoutCreateFkConstraint(db);err != nil{
		return err
	}
	if err := createFkConstraint(db);err != nil{
		return err
	}
	db.Migrator()
	return nil
}
