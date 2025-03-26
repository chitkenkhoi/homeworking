package routes
import(
	"lqkhoi-go-http-api/internal/middlewares"

	"github.com/gofiber/fiber/v2"
)
func sampleHanlder(c *fiber.Ctx) error{
	return nil
}

func SetupUserRoutes(app *fiber.App){
	app.Post("/users",sampleHanlder)
	app.Post("/login",sampleHanlder)

	authenticated := app.Group("/")
    authenticated.Use(middlewares.AuthMiddleware()) 
	
	authenticated.Get("/me", sampleHanlder)

	ownerOrAdmin := authenticated.Group("/users/:userId")
	ownerOrAdmin.Use(middlewares.RequireOwner())

    ownerOrAdmin.Get("/", sampleHanlder) 
	ownerOrAdmin.Put("/", sampleHanlder) 
	ownerOrAdmin.Delete("/", sampleHanlder) 

	adminOnly := authenticated.Group("/") 
    adminOnly.Use(middlewares.RequireAdmin()) 

    adminOnly.Get("/users", sampleHanlder)

}