package router

import (
	"context"
	"g4s-crm/api/internal/config"
	"g4s-crm/api/internal/handlers"
	"g4s-crm/api/internal/middleware"

	"g4s-crm/api/internal/services"
	"g4s-crm/api/migrations"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(nil))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "version": "1.0.0"})
	})

	r.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		sqlDB, err := db.DB()
		if err == nil {
			err = migrations.Ready(ctx, sqlDB)
		}
		if err != nil {
			c.JSON(503, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})

	// Initialize services
	authSvc, err := services.NewAuthService(db, &cfg.JWT)
	if err != nil {
		panic("failed to initialize auth service: " + err.Error())
	}

	// Initialize handlers
	authH := handlers.NewAuthHandler(authSvc)
	userH := handlers.NewUserHandler(db)
	teamH := handlers.NewTeamHandler(db)
	customerH := handlers.NewCustomerHandler(db)
	oppH := handlers.NewOpportunityHandler(db)
	mfrH := handlers.NewManufacturerHandler(db)
	productH := handlers.NewProductHandler(db)
	inventoryH := handlers.NewInventoryHandler(db)
	quoteH := handlers.NewQuoteHandler(db)
	projectH := handlers.NewProjectHandler(db)
	priceBookH := handlers.NewPriceBookHandler(db)
	fxH := handlers.NewExchangeRateHandler(db)
	contractH := handlers.NewContractHandler(db)
	catalogH := handlers.NewCatalogServiceHandler(db)
	recurringH := handlers.NewRecurringServiceHandler(db)
	procurementH := handlers.NewProcurementHandler(db)
	docH := handlers.NewDocumentHandler(db, cfg.Storage.Root)
	extractH := handlers.NewExtractHandler(db)
	dashH := handlers.NewDashboardHandler(db)

	api := r.Group("/api/v1")

	// ─── Auth (public) ───────────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.Refresh)
		auth.POST("/forgot-password", func(c *gin.Context) {
			c.JSON(501, gin.H{"success": false, "error": gin.H{"message": "Email password recovery is not configured; contact your administrator"}})
		})
		auth.POST("/reset-password", func(c *gin.Context) {
			c.JSON(501, gin.H{"success": false, "error": gin.H{"message": "Email password recovery is not configured; contact your administrator"}})
		})

		authProtected := auth.Group("")
		authProtected.Use(middleware.Auth(db))
		{
			authProtected.POST("/register", middleware.Authorize("users:create"), authH.Register)
			authProtected.POST("/logout", authH.Logout)
			authProtected.PATCH("/change-password", authH.ChangePassword)
			authProtected.GET("/me", authH.Me)
			authProtected.PATCH("/me", userH.Profile)
		}
	}

	// All routes below require authentication
	protected := api.Group("")
	protected.Use(middleware.Auth(db))

	// ─── Users ──────────────────────────────────────────────
	users := protected.Group("/users")
	{
		users.GET("", middleware.Authorize("users:read"), userH.List)
		users.GET("/lookup", userH.Lookup)
		users.GET("/:id", middleware.Authorize("users:read"), userH.Get)
		users.POST("", middleware.Authorize("users:create"), userH.Create)
		users.PATCH("/:id", middleware.Authorize("users:update"), userH.Update)
		users.PATCH("/:id/password", middleware.Authorize("users:update"), userH.ResetPassword)
		users.DELETE("/:id", middleware.Authorize("users:delete"), userH.Delete)
	}

	// ─── Teams ──────────────────────────────────────────────
	teams := protected.Group("/teams")
	{
		teams.GET("", middleware.Authorize("teams:read"), teamH.List)
		teams.GET("/:id", middleware.Authorize("teams:read"), teamH.Get)
		teams.POST("", middleware.Authorize("teams:create"), teamH.Create)
		teams.PATCH("/:id", middleware.Authorize("teams:update"), teamH.Update)
		teams.POST("/:id/members", middleware.Authorize("teams:update"), teamH.AddMember)
		teams.DELETE("/:id/members/:userId", middleware.Authorize("teams:update"), teamH.RemoveMember)
		teams.DELETE("/:id", middleware.Authorize("teams:delete"), teamH.Delete)
	}

	// ─── Customers ──────────────────────────────────────────
	customers := protected.Group("/customers")
	{
		customers.GET("", middleware.Authorize("customers:read"), customerH.List)
		customers.GET("/lookup", middleware.Authorize("customers:read"), customerH.Lookup)
		customers.GET("/:id", middleware.Authorize("customers:read"), customerH.Get)
		customers.POST("", middleware.Authorize("customers:create"), customerH.Create)
		customers.PATCH("/:id", middleware.Authorize("customers:update"), customerH.Update)
		customers.DELETE("/:id", middleware.Authorize("customers:delete"), customerH.Delete)

		customers.GET("/:id/sites", middleware.Authorize("customers:read"), customerH.ListSites)
		customers.POST("/:id/sites", middleware.Authorize("customers:update"), customerH.AddSite)
		customers.PATCH("/:id/sites/:siteId", middleware.Authorize("customers:update"), customerH.UpdateSite)
		customers.DELETE("/:id/sites/:siteId", middleware.Authorize("customers:update"), customerH.DeleteSite)

		customers.GET("/:id/contacts", middleware.Authorize("customers:read"), customerH.ListContacts)
		customers.POST("/:id/contacts", middleware.Authorize("customers:update"), customerH.AddContact)
		customers.PATCH("/:id/contacts/:contactId", middleware.Authorize("customers:update"), customerH.UpdateContact)
		customers.DELETE("/:id/contacts/:contactId", middleware.Authorize("customers:update"), customerH.DeleteContact)
	}

	// ─── Opportunities ──────────────────────────────────────
	opps := protected.Group("/opportunities")
	{
		opps.GET("", middleware.Authorize("opportunities:read"), oppH.List)
		opps.GET("/pipeline", middleware.Authorize("opportunities:read"), oppH.Pipeline)
		opps.GET("/:id", middleware.Authorize("opportunities:read"), oppH.Get)
		opps.POST("", middleware.Authorize("opportunities:create"), oppH.Create)
		opps.PATCH("/:id", middleware.Authorize("opportunities:update"), oppH.Update)
		opps.DELETE("/:id", middleware.Authorize("opportunities:delete"), oppH.Delete)
		opps.PATCH("/:id/stage", middleware.Authorize("opportunities:update"), oppH.UpdateStage)
	}

	// ─── Manufacturers ──────────────────────────────────────
	mfrs := protected.Group("/manufacturers")
	{
		mfrs.GET("", middleware.Authorize("manufacturers:read"), mfrH.List)
		mfrs.GET("/:id", middleware.Authorize("manufacturers:read"), mfrH.Get)
		mfrs.POST("", middleware.Authorize("manufacturers:create"), mfrH.Create)
		mfrs.PATCH("/:id", middleware.Authorize("manufacturers:update"), mfrH.Update)
		mfrs.DELETE("/:id", middleware.Authorize("manufacturers:delete"), mfrH.Delete)
		mfrs.GET("/:id/categories", middleware.Authorize("manufacturers:read"), mfrH.Categories)
		mfrs.POST("/:id/categories", middleware.Authorize("manufacturers:update"), mfrH.SaveCategory)
		mfrs.PATCH("/:id/categories/:categoryId", middleware.Authorize("manufacturers:update"), mfrH.SaveCategory)
		mfrs.DELETE("/:id/categories/:categoryId", middleware.Authorize("manufacturers:update"), mfrH.DeleteCategory)
	}

	// ─── Products ───────────────────────────────────────────
	products := protected.Group("/products")
	{
		products.GET("", middleware.Authorize("products:read"), productH.List)
		products.GET("/:id", middleware.Authorize("products:read"), productH.Get)
		products.POST("", middleware.Authorize("products:create"), productH.Create)
		products.PATCH("/:id", middleware.Authorize("products:update"), productH.Update)
		products.DELETE("/:id", middleware.Authorize("products:delete"), productH.Delete)
		products.POST("/recalculate-costs", middleware.Authorize("products:update"), productH.RecalculateCosts)
		products.POST("/import", middleware.Authorize("products:create"), middleware.Authorize("products:update"), productH.ImportProducts)
		products.GET("/:id/documents", middleware.Authorize("products:read"), productH.Documents)
		products.POST("/:id/documents", middleware.Authorize("products:update"), middleware.Authorize("documents:read"), productH.AddDocument)
		products.DELETE("/:id/documents/:documentId", middleware.Authorize("products:update"), productH.DeleteDocument)
		products.GET("/:id/vendors", middleware.Authorize("products:read"), productH.Vendors)
		products.POST("/:id/vendors", middleware.Authorize("products:update"), productH.SaveVendor)
		products.PATCH("/:id/vendors/:vendorId", middleware.Authorize("products:update"), productH.SaveVendor)
		products.DELETE("/:id/vendors/:vendorId", middleware.Authorize("products:update"), productH.DeleteVendor)
		products.GET("/:id/price-history", middleware.Authorize("products:read"), productH.PriceHistory)
		products.POST("/:id/price-history", middleware.Authorize("products:update"), productH.AddPrice)
	}

	// ─── Inventory ──────────────────────────────────────────
	inventory := protected.Group("/inventory")
	{
		inventory.GET("/stock", middleware.Authorize("inventory:read"), inventoryH.ListStock)
		inventory.GET("/stock/low-stock", middleware.Authorize("inventory:read"), inventoryH.LowStock)
		inventory.GET("/stock/by-warehouse/:location", middleware.Authorize("inventory:read"), inventoryH.ListStock)
		inventory.GET("/stock/by-product/:productId", middleware.Authorize("inventory:read"), inventoryH.ListStock)

		inventory.GET("/reservations", middleware.Authorize("inventory:read"), inventoryH.ListReservations)
		inventory.POST("/reservations", middleware.Authorize("inventory:create"), inventoryH.CreateReservation)
		inventory.PATCH("/reservations/:id/release", middleware.Authorize("inventory:update"), inventoryH.Release)
		inventory.PATCH("/reservations/:id/fulfill", middleware.Authorize("inventory:update"), inventoryH.Fulfill)
		inventory.POST("/movements/adjustment", middleware.Authorize("inventory:update"), inventoryH.Adjust)
		inventory.PATCH("/stock/:id/reorder-level", middleware.Authorize("inventory:update"), inventoryH.ReorderLevel)

		inventory.GET("/movements", middleware.Authorize("inventory:read"), inventoryH.ListMovements)
		inventory.POST("/movements/transfer", middleware.Authorize("inventory:create"), inventoryH.Transfer)
	}

	// ─── Quotes ─────────────────────────────────────────────
	quotes := protected.Group("/quotes")
	{
		quotes.GET("", middleware.Authorize("quotes:read"), quoteH.List)
		quotes.GET("/:id", middleware.Authorize("quotes:read"), quoteH.Get)
		quotes.POST("", middleware.Authorize("quotes:create"), quoteH.Create)
		quotes.PATCH("/:id", middleware.Authorize("quotes:update"), quoteH.Update)
		quotes.DELETE("/:id", middleware.Authorize("quotes:delete"), quoteH.Delete)
		quotes.PATCH("/:id/submit", middleware.Authorize("quotes:update"), quoteH.Submit)
		quotes.PATCH("/:id/approve", middleware.Authorize("quotes:approve"), quoteH.Approve)
		quotes.PATCH("/:id/reject", middleware.Authorize("quotes:approve"), quoteH.Reject)
		quotes.PATCH("/:id/send", middleware.Authorize("quotes:update"), quoteH.Send)
		quotes.PATCH("/:id/accept", middleware.Authorize("quotes:update"), quoteH.Accept)
		quotes.PATCH("/:id/decline", middleware.Authorize("quotes:update"), quoteH.Decline)
		quotes.POST("/:id/recalculate", middleware.Authorize("quotes:update"), quoteH.Recalculate)
		quotes.GET("/:id/activity", middleware.Authorize("quotes:read"), quoteH.Activity)
		quotes.GET("/:id/pdf", middleware.Authorize("quotes:read"), docH.ExportQuote)
		quotes.GET("/:id/builder", middleware.Authorize("quotes:read"), quoteH.Get)
		quotes.PUT("/:id/builder", middleware.Authorize("quotes:update"), quoteH.SaveBuilder)
		quotes.POST("/:id/duplicate", middleware.Authorize("quotes:create"), quoteH.Duplicate)
		quotes.POST("/:id/revisions", middleware.Authorize("quotes:create"), quoteH.Duplicate)
		quotes.POST("/:id/convert-to-contract", middleware.Authorize("contracts:create"), quoteH.ConvertToContract)
	}

	// ─── Projects ───────────────────────────────────────────
	projects := protected.Group("/projects")
	{
		projects.GET("", middleware.Authorize("projects:read"), projectH.List)
		projects.GET("/:id", middleware.Authorize("projects:read"), projectH.Get)
		projects.POST("", middleware.Authorize("projects:create"), projectH.Create)
		projects.PATCH("/:id", middleware.Authorize("projects:update"), projectH.Update)
		projects.DELETE("/:id", middleware.Authorize("projects:delete"), projectH.Delete)
	}

	// ─── Price Books ─────────────────────────────────────────
	priceBooks := protected.Group("/price-books")
	{
		priceBooks.GET("", middleware.Authorize("price-books:read"), priceBookH.List)
		priceBooks.GET("/:id", middleware.Authorize("price-books:read"), priceBookH.Get)
		priceBooks.POST("", middleware.Authorize("price-books:create"), priceBookH.Create)
		priceBooks.PATCH("/:id", middleware.Authorize("price-books:update"), priceBookH.Update)
		priceBooks.DELETE("/:id", middleware.Authorize("price-books:delete"), priceBookH.Delete)
	}

	priceBooks.GET("/:id/entries", middleware.Authorize("price-books:read"), priceBookH.Entries)
	priceBooks.PUT("/:id/entries", middleware.Authorize("price-books:update"), priceBookH.ReplaceEntries)
	priceBooks.POST("/:id/entries", middleware.Authorize("price-books:update"), priceBookH.SaveEntry)
	priceBooks.PATCH("/:id/entries/:entryId", middleware.Authorize("price-books:update"), priceBookH.SaveEntry)
	priceBooks.DELETE("/:id/entries/:entryId", middleware.Authorize("price-books:update"), priceBookH.DeleteEntry)
	// ─── Exchange Rates ──────────────────────────────────────
	fx := protected.Group("/exchange-rates")
	{
		fx.GET("", middleware.Authorize("exchange-rates:read"), fxH.List)
		fx.GET("/:id", middleware.Authorize("exchange-rates:read"), fxH.Get)
		fx.PATCH("/:id", middleware.Authorize("exchange-rates:update"), fxH.Update)
		fx.POST("/refresh", middleware.Authorize("exchange-rates:update"), fxH.Refresh)
	}

	// ─── Contracts ───────────────────────────────────────────
	contracts := protected.Group("/contracts")
	{
		contracts.GET("", middleware.Authorize("contracts:read"), contractH.List)
		contracts.GET("/expiring", middleware.Authorize("contracts:read"), contractH.Expiring)
		contracts.GET("/:id", middleware.Authorize("contracts:read"), contractH.Get)
		contracts.POST("", middleware.Authorize("contracts:create"), contractH.Create)
		contracts.PATCH("/:id", middleware.Authorize("contracts:update"), contractH.Update)
		contracts.DELETE("/:id", middleware.Authorize("contracts:delete"), contractH.Delete)
		contracts.PATCH("/:id/activate", middleware.Authorize("contracts:approve"), contractH.Activate)
		contracts.PATCH("/:id/terminate", middleware.Authorize("contracts:approve"), contractH.Terminate)
		contracts.POST("/:id/renew", middleware.Authorize("contracts:create"), contractH.Renew)
	}

	// ─── Recurring Services ──────────────────────────────────
	recurring := protected.Group("/recurring-services")
	{
		recurring.GET("", middleware.Authorize("recurring-services:read"), recurringH.List)
		recurring.GET("/:id", middleware.Authorize("recurring-services:read"), recurringH.Get)
		recurring.POST("", middleware.Authorize("recurring-services:create"), recurringH.Create)
		recurring.PATCH("/:id", middleware.Authorize("recurring-services:update"), recurringH.Update)
		recurring.DELETE("/:id", middleware.Authorize("recurring-services:delete"), recurringH.Delete)
	}

	// ─── Procurement ─────────────────────────────────────────
	proc := protected.Group("/procurement")
	{
		proc.GET("/supplier-items", middleware.Authorize("procurement:read"), procurementH.SupplierItems)
		proc.POST("/supplier-items", middleware.Authorize("procurement:create"), procurementH.SaveSupplierItem)
		proc.PATCH("/supplier-items/:id", middleware.Authorize("procurement:update"), procurementH.SaveSupplierItem)
		proc.DELETE("/supplier-items/:id", middleware.Authorize("procurement:delete"), procurementH.DeleteSupplierItem)
		pos := proc.Group("/purchase-orders")
		{
			pos.GET("", middleware.Authorize("procurement:read"), procurementH.ListPOs)
			pos.GET("/:id", middleware.Authorize("procurement:read"), procurementH.GetPO)
			pos.POST("", middleware.Authorize("procurement:create"), procurementH.CreatePO)
			pos.PATCH("/:id", middleware.Authorize("procurement:update"), procurementH.UpdatePO)
			pos.DELETE("/:id", middleware.Authorize("procurement:delete"), procurementH.DeletePO)
			pos.PATCH("/:id/approve", middleware.Authorize("procurement:approve"), procurementH.ApprovePO)
			pos.PATCH("/:id/status", middleware.Authorize("procurement:update"), procurementH.UpdatePOStatus)
		}

		sqs := proc.Group("/supplier-quotes")
		{
			sqs.GET("", middleware.Authorize("procurement:read"), procurementH.ListSQs)
			sqs.GET("/:id", middleware.Authorize("procurement:read"), procurementH.GetSQ)
			sqs.POST("", middleware.Authorize("procurement:create"), procurementH.CreateSQ)
			sqs.PATCH("/:id", middleware.Authorize("procurement:update"), procurementH.UpdateSQ)
			sqs.DELETE("/:id", middleware.Authorize("procurement:delete"), procurementH.DeleteSQ)
			sqs.PATCH("/:id/status", middleware.Authorize("procurement:update"), procurementH.UpdateSQStatus)
			sqs.POST("/:id/convert-to-po", middleware.Authorize("procurement:create"), procurementH.ConvertToPO)
		}

		grs := proc.Group("/goods-receipts")
		{
			grs.GET("", middleware.Authorize("procurement:read"), procurementH.ListGRs)
			grs.GET("/:id", middleware.Authorize("procurement:read"), procurementH.GetGR)
			grs.POST("", middleware.Authorize("procurement:create"), procurementH.CreateGR)
		}
	}

	// ─── Documents ───────────────────────────────────────────
	docs := protected.Group("/documents")
	{
		docs.GET("", middleware.Authorize("documents:read"), docH.List)
		docs.GET("/:id", middleware.Authorize("documents:read"), docH.Get)
		docs.POST("", middleware.Authorize("documents:create"), docH.Upload)
		docs.PATCH("/:id", middleware.Authorize("documents:update"), docH.Update)
		docs.GET("/:id/versions", middleware.Authorize("documents:read"), docH.Versions)
		docs.POST("/:id/versions", middleware.Authorize("documents:update"), docH.UploadVersion)
		docs.GET("/:id/versions/:versionId/download", middleware.Authorize("documents:read"), docH.Download)
		docs.POST("/:id/links", middleware.Authorize("documents:update"), docH.AddLink)
		docs.DELETE("/:id/links/:linkId", middleware.Authorize("documents:update"), docH.DeleteLink)
		docs.DELETE("/:id", middleware.Authorize("documents:delete"), docH.Delete)
		docs.GET("/:id/download", middleware.Authorize("documents:read"), docH.Download)
	}

	// ─── Dashboard ───────────────────────────────────────────
	dash := protected.Group("/dashboard")
	dash.Use(middleware.Authorize("dashboard:read"))
	{
		dash.GET("/recent-quotes", dashH.RecentQuotes)
		dash.GET("/expiring-quotes", dashH.ExpiringQuotes)
		dash.GET("/sales-performance", dashH.SalesPerformance)
		dash.GET("/kpis", dashH.KPIs)
		dash.GET("/pipeline", dashH.Pipeline)
		dash.GET("/recent-activity", dashH.RecentActivity)
		dash.GET("/top-customers", dashH.TopCustomers)
		dash.GET("/alerts", dashH.Alerts)
	}

	protected.GET("/catalog", middleware.Authorize("products:read"), productH.Catalog)
	// Reads an uploaded vendor PDF without storing it; used by both importers.
	protected.POST("/extract/pdf-tables", middleware.AuthorizeAny("products:create", "procurement:create"), extractH.PDFTables)
	catalog := protected.Group("/services")
	catalog.GET("", middleware.Authorize("products:read"), catalogH.List)
	catalog.GET("/:id", middleware.Authorize("products:read"), catalogH.Get)
	catalog.POST("", middleware.Authorize("products:create"), catalogH.Create)
	catalog.PATCH("/:id", middleware.Authorize("products:update"), catalogH.Update)
	catalog.DELETE("/:id", middleware.Authorize("products:delete"), catalogH.Delete)
	protected.POST("/exchange-rates", middleware.Authorize("exchange-rates:update"), fxH.Create)
	return r
}
