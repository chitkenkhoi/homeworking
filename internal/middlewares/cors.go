package middlewares

import(
	"log"
	"strings"

	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsConfig() fiber.Handler {
	allowedOrigins := utils.GetenvStringValue("ALLOWED_ORIGINS","*")

	if allowedOrigins == "*" {
		log.Println("Warning: CORS AllowOrigins is set to '*' - Allows ALL origins. Restrict this in production!")
	} else {
		log.Printf("CORS AllowedOrigins: %s", allowedOrigins)
	}

	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
		 if allowedOrigins == "*" {
		     return true
		 }
		 for allowed := range strings.SplitSeq(allowedOrigins, ",") {
		     if origin == strings.TrimSpace(allowed) {
		         return true
		     }
		 }
		 return false
		},

		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodOptions,
		}, ","),

		AllowHeaders: strings.Join([]string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization", 
		}, ","),

		AllowCredentials: true, // Important if your frontend needs to send cookies or Auth header

		// ExposeHeaders: "X-Custom-Header", // Uncomment if frontend needs to read custom response headers

		MaxAge: 86400, // Cache preflight request for 1 day (optional)
	})
}