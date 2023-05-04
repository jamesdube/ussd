package session

import "github.com/gofiber/fiber/v2"

type Repository interface {
	GetSession(id string) (*Session, error)
	Save(s *Session) error
	Delete(id string)
}

type FiberRepository interface {
	GetSession(ctx *fiber.Ctx, id string) (*Session, error)
	Save(ctx *fiber.Ctx, s *Session)
	Delete(ctx *fiber.Ctx, id string)
}
