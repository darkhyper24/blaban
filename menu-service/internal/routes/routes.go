package routes

import (
	"github.com/darkhyper24/blaban/menu-service/internal/db"
	"github.com/darkhyper24/blaban/menu-service/services"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, menuDB db.MenuDBOperations) {
	menuService := services.NewMenuService(menuDB)
	app.Get("/api/categories", menuService.HandleGetCategories)
	app.Get("/api/menu", menuService.HandleGetMenu)
	app.Get("/api/menu/search", menuService.HandleSearchItems)
	app.Get("/api/menu/filter", menuService.HandleFilterItems)
	app.Get("/api/menu/:id", menuService.HandleGetMenuItem)
	app.Post("/api/menu", menuService.HandleCreateMenuItem)
	app.Patch("/api/menu/:id", menuService.HandleUpdateMenuItem)
	app.Delete("/api/menu/:id", menuService.HandleDeleteMenuItem)
	app.Post("/api/menu/:id/discount", menuService.HandleAddDiscount)
}
