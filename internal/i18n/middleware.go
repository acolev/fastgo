package i18n

import "github.com/gofiber/fiber/v3"

func Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {

		lang := c.Get("X-Lang")

		if lang == "" {
			lang = c.Get("Accept-Language")
		}

		if len(lang) >= 2 {
			lang = lang[:2]
		}

		if lang == "" {
			lang = defaultLang
		}

		c.Locals("lang", lang)

		return c.Next()
	}
}
