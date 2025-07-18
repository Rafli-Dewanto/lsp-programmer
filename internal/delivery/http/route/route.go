package route

import (
	"cakestore/internal/constants"
	http "cakestore/internal/delivery/http"
	"cakestore/internal/middleware"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/sirupsen/logrus"
)

type RouteConfig struct {
	App                   *fiber.App
	MenuController        *http.MenuController
	CustomerController    *http.CustomerController
	CartController        *http.CartController
	OrderController       *http.OrderController
	WishlistController    *http.WishListController
	PaymentController     http.PaymentController
	ReservationController *http.ReservationController
	InventoryController   *http.InventoryController
	TableController       *http.TableController
	JWTSecret             string
	Log                   *logrus.Logger
}

func (c *RouteConfig) Setup() {
	c.SetupRoute()
}

func (c *RouteConfig) SetupRoute() {
	c.App.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PATCH,PUT,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-App-Role, User-Agent",
	}))
	c.App.Use(pprof.New())
	c.App.Use(middleware.LogMiddleware(c.Log))
	c.App.Static("/docs", "./docs")
	cfg := swagger.Config{
		FilePath: "./docs/swagger.json",
		Path:     "docs",
		Title:    "Swagger API Docs",
		BasePath: "/api/v1/",
	}
	c.App.Use(swagger.New(cfg))

	// Public routes
	c.App.Post("/register", c.CustomerController.Register)
	c.App.Post("/login", c.CustomerController.Login)
	// Midtrans notification webhook
	c.App.Post("/payment/notification/", c.PaymentController.GetTransactionStatus)
	// menus
	c.App.Get("/menus", c.MenuController.GetAllMenus)
	c.App.Get("/menus/:id", c.MenuController.GetMenuByID)

	// Protected routes
	protectedRoutes := c.App.Group("/api/v1", middleware.AuthMiddleware(c.JWTSecret))

	// Customer routes
	protectedRoutes.Get("/authorize", c.CustomerController.Authorize)
	protectedRoutes.Get("/customers/me", c.CustomerController.GetCustomerByID)
	protectedRoutes.Put("/customers/:id", c.CustomerController.UpdateProfile)

	// employee routes
	employeeRoutes := protectedRoutes.Group("/employees")
	employeeRoutes.Get("/", c.CustomerController.GetEmployees)
	employeeRoutes.Get("/:id", c.CustomerController.GetEmployeeByID)
	employeeRoutes.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CustomerController.UpdateEmployee)
	employeeRoutes.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin), c.CustomerController.DeleteEmployee)

	// Menu routes
	menus := protectedRoutes.Group("/menus")
	menus.Post("/", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.MenuController.CreateMenu)
	menus.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.MenuController.UpdateMenu)
	menus.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier, constants.RoleKitchen, constants.RoleWaitress), c.MenuController.DeleteMenu)

	// Cart routes
	carts := protectedRoutes.Group("/carts")
	carts.Post("/", c.CartController.AddCart)
	carts.Get("/customer", c.CartController.GetCartByCustomerID)
	carts.Get("/:id", c.CartController.GetCartByID)
	carts.Delete("/:id", c.CartController.RemoveCart)
	carts.Delete("/", c.CartController.ClearCart)
	carts.Post("/bulk", c.CartController.BulkDeleteCart)

	// Order routes
	orders := protectedRoutes.Group("/orders")
	orders.Get("/customers", c.OrderController.GetAllOrders)
	orders.Post("/", c.OrderController.CreateOrder)
	orders.Get("/", c.OrderController.GetCustomerOrders)
	orders.Get("/:id", c.OrderController.GetOrderByID)
	orders.Patch("/:id/food-status", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.OrderController.UpdateFoodStatus)

	// payment routes
	payment := protectedRoutes.Group("/payments")
	payment.Get("/:id", c.PaymentController.GetPaymentURL)

	// Wishlist routes
	wishlist := protectedRoutes.Group("/wishlists")
	wishlist.Get("/", c.WishlistController.GetWishListByCustomerID)
	wishlist.Post("/:menuId", c.WishlistController.CreateWishList)
	wishlist.Delete("/:menuId", c.WishlistController.DeleteWishList)

	// Reservation routes
	reservation := protectedRoutes.Group("/reservations")
	reservation.Post("/", c.ReservationController.CreateReservation)
	reservation.Get("/", c.ReservationController.GetAllReservations)
	reservation.Get("/admin", middleware.RoleMiddleware(constants.RoleAdmin), c.ReservationController.AdminGetAllCustomerReservations)
	reservation.Get("/:id", c.ReservationController.GetReservationByID)
	reservation.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleWaitress), c.ReservationController.UpdateReservation)
	reservation.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleWaitress), c.ReservationController.DeleteReservation)

	// Ingredient routes
	inventory := protectedRoutes.Group("/inventories")
	inventory.Get("/", c.InventoryController.GetAllInventories)
	inventory.Get("/low-stock", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.GetLowStockInventories)
	// temporary fix for conflicting route (/low-stock)
	inventory.Get("/by-id/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.GetInventoryByID)
	inventory.Post("/", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.CreateInventory)
	inventory.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.UpdateInventory)
	inventory.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.DeleteInventory)
	inventory.Put("/:id/stock", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleKitchen), c.InventoryController.UpdateInventoryStock)

	// Table routes
	tables := protectedRoutes.Group("/tables")
	tables.Get("/", c.TableController.GetAllTables)
	tables.Get("/:id", c.TableController.GetTableByID)
	tables.Post("/", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.CreateTable)
	tables.Put("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.UpdateTable)
	tables.Patch("/:id/availability", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.UpdateTableAvailability)
	tables.Delete("/:id", middleware.RoleMiddleware(constants.RoleAdmin, constants.RoleCashier), c.TableController.DeleteTable)
}
