
# Skill: create_http_feature

Creates a new HTTP feature following FastGo architecture.

Target directory:

internal/http/<feature>/

Structure:

handlers/
services/
models/
routes.go

Example:

internal/http/users/

handlers/
    users_handler.go

services/
    users_service.go

models/
    user.go

routes.go

Handler template:

type Handler struct {
    service *services.Service
}

func New() *Handler {
    return &Handler{
        service: services.New(),
    }
}

Routes template:

func RegisterRoutes(r fiber.Router)
