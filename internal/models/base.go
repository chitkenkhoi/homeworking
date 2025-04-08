package models

// Identifiable defines an interface for models that have a primary key.
// K is the type of the primary key (e.g., int, string, uuid.UUID).
type Identifiable[K comparable] interface {
	GetID() K
	GetPKColumnName() string
	// TableName() string // GORM usually infers this, but needed if explicit generic query needed
}

// Example BaseModel implementing Identifiable
type BaseModel[K comparable] struct {
	// Define your common fields like ID, CreatedAt, UpdatedAt
	// The actual ID field needs to be in the concrete struct for GORM mapping
}

// func (b *BaseModel[K]) GetPKColumnName() string {
// 	return "id" 
// }