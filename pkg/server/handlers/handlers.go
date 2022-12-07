// Package handlers contains handlers for all Task API endpoints.
package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/richard-on/auth-service/pkg/authService"
	request2 "github.com/richard-on/mail-service/pkg/server/request"
	"github.com/richard-on/mail-service/pkg/templates"
	"github.com/richard-on/task-service/config"
	"github.com/richard-on/task-service/internal/db"
	"github.com/richard-on/task-service/internal/model"
	"github.com/richard-on/task-service/pkg/logger"
	"github.com/richard-on/task-service/pkg/server/request"
	"github.com/richard-on/task-service/pkg/server/response"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskHandler struct {
	Router      fiber.Router
	AuthService authService.AuthServiceClient
	Db          *db.DB
	log         logger.Logger
}

func NewTaskHandler(router fiber.Router, db *db.DB, authService authService.AuthServiceClient) *TaskHandler {
	return &TaskHandler{
		Router:      router,
		AuthService: authService,
		Db:          db,
		log:         logger.NewLogger(config.DefaultWriter, config.LogInfo.Level, "task-handler"),
	}
}

// List
// @Summary      List
// @Tags         List
// @Description  List tasks
// @ID           list-tasks
// @Produce      json
// @Success      200      {object}  handlers.ListResponse
// @Failure      403,500  {object}  handlers.ErrorResponse
// @Router       /tasks [get]
func (h *TaskHandler) List(ctx *fiber.Ctx) error {
	validateResponse, err := Validate(ctx)
	if err != nil {
		h.log.Error(err, "unable to get tasks")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	tasks, err := h.Db.GetAllTasks(validateResponse.Email)
	if err != nil {
		h.log.Error(err, "unable to get tasks")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if len(tasks) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(response.Error{Error: ErrNoTasks.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.ListResponse{Tasks: tasks})
}

// Add
// @Summary      Add
// @Tags         add
// @Description  Add task
// @ID           run
// @Accept       json
// @Produce      json
// @Param        input    body      handlers.RunRequest  true  "Run info"
// @Success      200      {object}  handlers.TaskResponse
// @Failure      400,403,500  {object}  handlers.ErrorResponse
// @Router       /add [post]
func (h *TaskHandler) Add(ctx *fiber.Ctx) error {
	validateRequest := &authService.ValidateRequest{
		AccessToken:  ctx.Cookies("accessToken"),
		RefreshToken: ctx.Cookies("refreshToken"),
	}

	// Check access token validity
	validateResponse, err := h.AuthService.Validate(ctx.Context(), validateRequest)
	if err != nil {
		h.log.Debug(err)

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{Error: err.Error()})
	}

	var addRequest request.AddRequest
	if err = ctx.BodyParser(&addRequest); err != nil {
		h.log.Debug(err, "parsing error")
		return ctx.Status(fiber.StatusBadRequest).JSON(response.Error{Error: err.Error()})
	}

	if len(addRequest.Coordinators) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(response.Error{Error: ErrNoCoordinators.Error()})
	}

	task, err := h.Db.AddTask(model.Task{
		ID:           primitive.NewObjectID(),
		Name:         addRequest.Name,
		Description:  addRequest.Description,
		Initiator:    validateResponse.Email,
		Coordinators: addRequest.Coordinators,
		Next:         0,
		Status:       model.NotStarted,
	})
	if err != nil {
		h.log.Error(err, "unable to add task to database")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	sendReq := request2.SendMail{
		From:    validateResponse.Email,
		Subject: task.Description,
		To:      task.Coordinators[task.Next],
		Type:    "coordination",
		Template: templates.Coordination{
			AcceptLink: fmt.Sprintf("localhost:5000/task/v1/approve/%v/%v",
				task.Coordinators[task.Next], task.ID.Hex()),
			DeclineLink: fmt.Sprintf("localhost:5000/task/v1/decline/%v/%v",
				task.Coordinators[task.Next], task.ID.Hex()),
		},
	}

	SendEmail(ctx, sendReq)

	return ctx.Status(fiber.StatusOK).JSON(response.AddResponse{
		ID:           task.ID,
		Initiator:    task.Initiator,
		Name:         task.Name,
		Description:  task.Description,
		Coordinators: task.Coordinators,
		Status:       task.Status,
	})
}

// Delete
// @Summary      Delete
// @Tags         Delete
// @Description  Delete model
// @ID           delete-model
// @Produce      json
// @Param        task_id  path      string  true  "Task ID"
// @Success      200      {object}  handlers.TaskResponse
// @Failure      400,403,500  {object}  handlers.ErrorResponse
// @Router       /delete:task_id [delete]
func (h *TaskHandler) Delete(ctx *fiber.Ctx) error {
	validateRequest := &authService.ValidateRequest{
		AccessToken:  ctx.Cookies("accessToken"),
		RefreshToken: ctx.Cookies("refreshToken"),
	}

	// Check access token validity
	validateResponse, err := h.AuthService.Validate(ctx.Context(), validateRequest)
	if err != nil {
		h.log.Debug(err)

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{Error: err.Error()})
	}

	taskId := ctx.Params("task_id")
	task, err := h.Db.GetTaskById(taskId)
	if err != nil {
		h.log.Debug(err)

		return ctx.Status(fiber.StatusBadRequest).JSON(response.Error{Error: err.Error()})
	}
	if validateResponse.Email != task.Initiator {
		h.log.Debug(ErrNoAccess)

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{
			Error: ErrNoAccess.Error(),
		})
	}

	err = h.Db.DeleteTask(taskId)
	if err != nil {
		h.log.Error(err, "unable to delete task")

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.Status(fiber.StatusOK).JSON(response.Info{
		Message: fmt.Sprintf("successfully deleted task %v", taskId),
	})
}

// Approve is an endpoint to approve.
// @Summary      Approve
// @Tags         Approve
// @Description  Approve model
// @ID           approve-model
// @Produce      json
// @Param        task_id        path      string  true  "Task ID"
// @Param        approvalLogin  path      string  true  "Approval login"
// @Success      200            {object}  handlers.TaskResponse
// @Failure      400,403,500        {object}  handlers.ErrorResponse
// @Router       /approve/:coordinator\:task_id [post]
func (h *TaskHandler) Approve(ctx *fiber.Ctx) error {
	validateRequest := &authService.ValidateRequest{
		AccessToken:  ctx.Cookies("accessToken"),
		RefreshToken: ctx.Cookies("refreshToken"),
	}

	// Check access token validity
	validateResponse, err := h.AuthService.Validate(ctx.Context(), validateRequest)
	if err != nil {
		h.log.Debug(err)

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{Error: err.Error()})
	}

	coordinator := ctx.Params("coordinator")
	taskID := ctx.Params("task_id")
	task, err := h.Db.GetTaskById(taskID)
	if err != nil {
		h.log.Debug(err)

		return ctx.Status(fiber.StatusBadRequest).JSON(response.Error{Error: err.Error()})
	} else if validateResponse.Email != task.Coordinators[task.Next] || validateResponse.Email != coordinator {
		h.log.Debug(ErrNoAccess)

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{
			Error: ErrNoAccess.Error(),
		})
	} else if task.Status != model.InProgress && task.Status != model.NotStarted {
		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{
			Error: ErrAlreadyFinished.Error(),
		})
	}

	if task.Next+1 < len(task.Coordinators) {
		task.Next = task.Next + 1
		task.Status = model.InProgress
	} else {
		task.Status = model.Approved
	}

	err = h.Db.UpdateTask(&task)
	if err != nil {
		h.log.Error(err, "unable to update task")

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if task.Status == model.Approved {

		for _, v := range task.Coordinators {
			sendReq := request2.SendMail{
				From:    validateResponse.Email,
				Subject: task.Description,
				To:      v,
				Type:    "info",
				Template: templates.Info{
					Body: "TASK VERIFIED!",
				},
			}

			SendEmail(ctx, sendReq)
		}

		return ctx.Status(fiber.StatusOK).JSON(response.Info{
			Message: "coordination end: approved",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response.Info{
		Message: fmt.Sprintf("you have approved this task: next coordinator: %v",
			task.Coordinators[task.Next]),
	})
}

// Decline task
// @Summary      Decline
// @Tags         Decline
// @Description  Decline task
// @ID           decline
// @Produce      json
// @Param        task_id        path      string  true  "Task ID"
// @Param        approvalLogin  path      string  true  "Approval login"
// @Success      200            {object}  handlers.TaskResponse
// @Failure      400,403,500        {object}  handlers.ErrorResponse
// @Router       /decline/:coordinator\:task_id [post]
func (h *TaskHandler) Decline(ctx *fiber.Ctx) error {
	validateRequest := &authService.ValidateRequest{
		AccessToken:  ctx.Cookies("accessToken"),
		RefreshToken: ctx.Cookies("refreshToken"),
	}

	// Check access token validity
	validateResponse, err := h.AuthService.Validate(ctx.Context(), validateRequest)
	if err != nil {
		h.log.Debug(err)

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{Error: err.Error()})
	}

	coordinator := ctx.Params("coordinator")
	taskID := ctx.Params("task_id")
	task, err := h.Db.GetTaskById(taskID)
	if err != nil {
		h.log.Debug(err)

		return ctx.Status(fiber.StatusBadRequest).JSON(response.Error{Error: err.Error()})
	} else if validateResponse.Email != task.Coordinators[task.Next] || validateResponse.Email != coordinator {
		h.log.Debug(ErrNoAccess)

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{
			Error: ErrNoAccess.Error(),
		})
	} else if task.Status != model.InProgress && task.Status != model.NotStarted {
		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{
			Error: ErrAlreadyFinished.Error(),
		})
	}

	task.Status = model.Declined

	err = h.Db.UpdateTask(&task)
	if err != nil {
		h.log.Error(err, "unable to update task")

		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.Status(fiber.StatusOK).JSON(response.Info{
		Message: "you have declined this task",
	})
}

// Run
// @Summary      Run
// @Tags         Run
// @Description  Run tasks
// @ID           run
// @Accept       json
// @Produce      json
// @Param        input    body      handlers.RunRequest  true  "Run info"
// @Success      200      {object}  handlers.TaskResponse
// @Failure      403,500  {object}  handlers.ErrorResponse
// @Router       /model/v1/tasks/run [post]
/*func (h *TaskHandler) Run(ctx *fiber.Ctx) error {
	var runRequest request.RunRequest
	if err := ctx.BodyParser(&runRequest); err != nil {
		log.Error().Err(err).Msg(err.Error())
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.Error{Error: err.Error()})
	}

	validateRequest := &authService.ValidateRequest{
		AccessToken:  ctx.Cookies("accessToken"),
		RefreshToken: ctx.Cookies("refreshToken"),
	}

	// Check access token validity
	validateResponse, err := h.AuthService.Validate(ctx.Context(), validateRequest)
	if err != nil {
		log.Error().Err(err).Msg(err.Error())

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{Error: err.Error()})
	}

	if validateResponse.Username != runRequest.Initiator {
		log.Debug().Msg("Unauthorized to perform RUN on this model")

		return ctx.Status(fiber.StatusForbidden).JSON(response.Error{
			Error: errors.New("unauthorized to perform RUN on this model").Error(),
		})
	}

	task, err := h.Db.AddTask(model.Task{
		Initiator:    runRequest.Initiator,
		Coordinators: runRequest.Coordinators,
	})
	if err != nil {
		log.Fatal().Stack().Msg(err.Error())
		return ctx.Status(fiber.StatusInternalServerError).JSON(response.Error{Error: err.Error()})
	}

	resp := fiber.Post("http://localhost:80/mail/v1/send")
	resp.Post()

	return ctx.Status(fiber.StatusOK).JSON(response.TaskResponse{
		ID:             task.ID,
		InitiatorLogin: task.InitiatorLogin,
		ApprovalLogins: task.ApprovalLogins,
	})
}*/
