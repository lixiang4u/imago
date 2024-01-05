package utils

import "github.com/gofiber/fiber/v2"

const defaultLimit = 20

type Pager struct {
	Limit  int
	Offset int
	Page   int
	Total  int64
}

func ParsePage(ctx *fiber.Ctx) *Pager {
	var p = &Pager{
		Limit:  ctx.QueryInt("limit", defaultLimit),
		Offset: 0,
		Page:   ctx.QueryInt("page", 1),
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit > 100 {
		p.Page = 100
	}
	if p.Limit <= 0 {
		p.Limit = defaultLimit
	}
	p.Offset = (p.Page - 1) * p.Limit

	return p
}
