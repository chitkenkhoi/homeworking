package handler

import (
	"log"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/service"
	
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct{
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler{
	return &UserHandler{
		userService: userService,
	}
}

func CreateUserHandler(c *fiber.Ctx) error {
	input := new(dto.CreateUserRequest) // Use new() or &CreateUserRequest{}

	// 1. Parse the request body into the DTO
	if err := c.BodyParser(input); err != nil {
		// Handle JSON parsing errors (e.g., malformed JSON)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
			"details": err.Error(),
		})
	}

    // 2. Validate the struct
    errors := ValidateStruct(*input) // Pass the struct value
    if errors != nil {
        // Return validation errors
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Validation failed",
            "details": errors,
        })
    }


	// 3. If validation passes, proceed to your service layer
	// The service layer should:
	//    - Hash input.Password
	//    - Create the main models.User struct
	//    - Save via repository
	// userService := // ... get your user service instance

	// createdUser, err := userService.CreateUser(*input) // Service takes DTO
	// if err != nil {
	// 	// Handle service errors (e.g., duplicate username/email)
	// 	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
	//      "error": "Failed to create user",
    //      // Optionally include more detail depending on the error type
	//  })
	// }

    // --- Placeholder Success ---
    // Replace with actual call to userService.CreateUser and return the createdUser
    // Remember the createdUser (type User) has json:"-" on Password, so it's safe.
    log.Printf("Validation successful for input: %+v\n", *input) // Log for debugging
    // Placeholder response - replace with actual created user
    placeholderUser := map[string]interface{}{
        "id": 1,
        "username": input.Username,
        "email": input.Email,
        "first_name": input.FirstName,
        "last_name": input.LastName,
        "role": input.Role,
        // Password is NOT included
    }
	return c.Status(fiber.StatusCreated).JSON(placeholderUser) // Return the created user (without password)
}