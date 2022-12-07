package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/richard-on/auth-service/pkg/authService"
	"github.com/richard-on/task-service/internal/db"
	"github.com/richard-on/task-service/pkg/server/handlers"
)

func TaskRouter(app fiber.Router, db *db.DB, authClient authService.AuthServiceClient) {

	handler := handlers.NewTaskHandler(app, db, authClient)

	app.Get("/tasks", handler.List)

	app.Post("/add", handler.Add)

	app.Delete("/delete/:task_id", handler.Delete)

	app.Post("/approve/:coordinator/:task_id", handler.Approve)

	app.Post("/decline/:coordinator/:task_id", handler.Decline)

	/*app.Post("/approve/:approvalLogin:task_id", handler.Approve)

	app.Post("/tasks/:task_id/decline/:approvalLogin", handler.Decline)

	app.Post("/tasks/run", handler.Run)*/

}
