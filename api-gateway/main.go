package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "gym-membership/proto/asset"
	membership "gym-membership/proto/membership"
)

func getContextWithMetadata(c *gin.Context) context.Context {
	authHeader := c.GetHeader("Authorization")
	md := metadata.New(nil)
	if authHeader != "" {
		md.Set("authorization", authHeader)
	}
	role := c.GetHeader("X-User-Role")
	if role != "" {
		md.Set("x-user-role", role)
	}
	return metadata.NewOutgoingContext(c.Request.Context(), md)
}

func main() {
	assetServiceAddr := os.Getenv("ASSET_SERVICE_ADDR")
	if assetServiceAddr == "" {
		assetServiceAddr = "localhost:50051"
	}

	membershipServiceAddr := os.Getenv("MEMBERSHIP_SERVICE_ADDR")
	if membershipServiceAddr == "" {
		membershipServiceAddr = "localhost:50052"
	}

	// Connect to Asset Service
	conn, err := grpc.Dial(assetServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to asset-service: %v", err)
	}
	defer conn.Close()
	assetClient := pb.NewAssetServiceClient(conn)

	// Connect to Membership Service
	mConn, err := grpc.Dial(membershipServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to membership-service: %v", err)
	}
	defer mConn.Close()
	membershipClient := membership.NewMembershipServiceClient(mConn)

	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-Role")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// === Service A: Asset Service Endpoints (12 Strict Endpoints) ===

	// 1. GET /assets -> ListAvailableAssets
	r.GET("/assets", func(c *gin.Context) {
		assetType := c.Query("type")
		resp, err := assetClient.ListAvailableAssets(getContextWithMetadata(c), &pb.ListRequest{Type: assetType})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp.Assets)
	})

	// 2. GET /assets/:id -> GetAsset
	r.GET("/assets/:id", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := assetClient.GetAsset(getContextWithMetadata(c), &pb.GetAssetRequest{Id: id})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 3. GET /assets/:id/health -> GetHealthScore
	r.GET("/assets/:id/health", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := assetClient.GetHealthScore(getContextWithMetadata(c), &pb.GetAssetRequest{Id: id})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"health_score": resp.HealthScore})
	})

	// 4. PATCH /assets/:id/status -> UpdateAssetStatus (admin only)
	r.PATCH("/assets/:id/status", adminOnly(), func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
			return
		}

		resp, err := assetClient.UpdateAssetStatus(getContextWithMetadata(c), &pb.UpdateStatusRequest{
			Id:     id,
			Status: input.Status,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 5. GET /assets/check -> CheckAvailability
	r.GET("/assets/check", func(c *gin.Context) {
		id := c.Query("id")
		startTime := c.Query("start_time")
		endTime := c.Query("end_time")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		resp, err := assetClient.CheckAvailability(getContextWithMetadata(c), &pb.CheckRequest{
			Id:        id,
			StartTime: startTime,
			EndTime:   endTime,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"available": resp.Available})
	})

	// 6. POST /assets -> CreateAsset (admin only)
	r.POST("/assets", adminOnly(), func(c *gin.Context) {
		var input struct {
			Name        string `json:"name" binding:"required"`
			Type        string `json:"type" binding:"required"`
			Status      string `json:"status"`
			HealthScore int32  `json:"health_score"`
			Location    string `json:"location"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := assetClient.CreateAsset(getContextWithMetadata(c), &pb.CreateAssetRequest{
			Name:        input.Name,
			Type:        input.Type,
			Status:      input.Status,
			HealthScore: input.HealthScore,
			Location:    input.Location,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, resp)
	})

	// 7. PUT /assets/:id -> UpdateAsset (admin only)
	r.PUT("/assets/:id", adminOnly(), func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Name     string `json:"name" binding:"required"`
			Type     string `json:"type" binding:"required"`
			Location string `json:"location"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp, err := assetClient.UpdateAsset(getContextWithMetadata(c), &pb.UpdateAssetRequest{
			Id:       id,
			Name:     input.Name,
			Type:     input.Type,
			Location: input.Location,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 8. DELETE /assets/:id -> DeleteAsset (admin only)
	r.DELETE("/assets/:id", adminOnly(), func(c *gin.Context) {
		id := c.Param("id")
		resp, err := assetClient.DeleteAsset(getContextWithMetadata(c), &pb.DeleteAssetRequest{
			Id: id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 9. POST /assets/:id/damage -> ReportDamage
	r.POST("/assets/:id/damage", func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Amount int32 `json:"amount" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "amount is required"})
			return
		}
		resp, err := assetClient.ReportDamage(getContextWithMetadata(c), &pb.ReportDamageRequest{
			Id:           id,
			DamageAmount: input.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 10. POST /assets/:id/maintenance/resolve -> ResolveMaintenance (admin only)
	r.POST("/assets/:id/maintenance/resolve", adminOnly(), func(c *gin.Context) {
		id := c.Param("id")
		resp, err := assetClient.ResolveMaintenance(getContextWithMetadata(c), &pb.GetAssetRequest{
			Id: id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 11. GET /assets/all -> ListAllAssets
	r.GET("/assets/all", func(c *gin.Context) {
		assetType := c.Query("type")
		resp, err := assetClient.ListAllAssets(getContextWithMetadata(c), &pb.ListAllRequest{Type: assetType})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp.Assets)
	})

	// 12. POST /assets/batch -> BatchCreateAssets (admin only)
	r.POST("/assets/batch", adminOnly(), func(c *gin.Context) {
		var input struct {
			Assets []struct {
				Name        string `json:"name" binding:"required"`
				Type        string `json:"type" binding:"required"`
				Status      string `json:"status"`
				HealthScore int32  `json:"health_score"`
				Location    string `json:"location"`
			} `json:"assets" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var protoAssets []*pb.CreateAssetRequest
		for _, a := range input.Assets {
			protoAssets = append(protoAssets, &pb.CreateAssetRequest{
				Name:        a.Name,
				Type:        a.Type,
				Status:      a.Status,
				HealthScore: a.HealthScore,
				Location:    a.Location,
			})
		}
		resp, err := assetClient.BatchCreateAssets(getContextWithMetadata(c), &pb.BatchCreateRequest{
			Assets: protoAssets,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, resp.Assets)
	})

	// === Service B: Membership/Credit Service Endpoints (12 Strict Endpoints) ===

	// 1. POST /bookings -> CreateBooking
	r.POST("/bookings", func(c *gin.Context) {
		var input struct {
			UserID    string `json:"user_id" binding:"required"`
			AssetID   string `json:"asset_id" binding:"required"`
			StartTime string `json:"start_time" binding:"required"`
			EndTime   string `json:"end_time" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := membershipClient.CreateBooking(getContextWithMetadata(c), &membership.CreateBookingRequest{
			UserId:    input.UserID,
			AssetId:   input.AssetID,
			StartTime: input.StartTime,
			EndTime:   input.EndTime,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, resp)
	})

	// 2. DELETE /bookings/:id -> CancelBooking
	r.DELETE("/bookings/:id", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := membershipClient.CancelBooking(getContextWithMetadata(c), &membership.CancelBookingRequest{
			BookingId: id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 3. POST /bookings/:id/return -> ReturnBooking
	r.POST("/bookings/:id/return", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := membershipClient.ReturnBooking(getContextWithMetadata(c), &membership.ReturnBookingRequest{
			BookingId: id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 4. GET /users/:id/credits -> GetUserCredits
	r.GET("/users/:id/credits", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := membershipClient.GetUserCredits(getContextWithMetadata(c), &membership.GetUserCreditsRequest{
			UserId: id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 5. POST /users/:id/credits/deduct -> DeductCredits
	r.POST("/users/:id/credits/deduct", func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Amount int32 `json:"amount" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := membershipClient.DeductCredits(getContextWithMetadata(c), &membership.DeductCreditsRequest{
			UserId: id,
			Amount: input.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 6. POST /users/:id/credits/add -> AddCredits
	r.POST("/users/:id/credits/add", func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Amount int32 `json:"amount" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := membershipClient.AddCredits(getContextWithMetadata(c), &membership.AddCreditsRequest{
			UserId: id,
			Amount: input.Amount,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 7. GET /users/:id/bookings -> GetUserBookings
	r.GET("/users/:id/bookings", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := membershipClient.GetUserBookings(getContextWithMetadata(c), &membership.GetUserBookingsRequest{
			UserId: id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp.Bookings)
	})

	// 8. POST /users -> CreateUser
	r.POST("/users", func(c *gin.Context) {
		var input struct {
			Name            string `json:"name" binding:"required"`
			Email           string `json:"email" binding:"required"`
			StartingCredits int32  `json:"starting_credits"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := membershipClient.CreateUser(getContextWithMetadata(c), &membership.CreateUserRequest{
			Name:            input.Name,
			Email:           input.Email,
			StartingCredits: input.StartingCredits,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, resp)
	})

	// 9. GET /users/:id -> GetUser
	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := membershipClient.GetUser(getContextWithMetadata(c), &membership.GetUserRequest{
			UserId: id,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 10. PUT /users/:id -> UpdateUser
	r.PUT("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Name  string `json:"name" binding:"required"`
			Email string `json:"email" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := membershipClient.UpdateUser(getContextWithMetadata(c), &membership.UpdateUserRequest{
			UserId: id,
			Name:   input.Name,
			Email:  input.Email,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 11. GET /users/:id/membership -> GetUserMembership
	r.GET("/users/:id/membership", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := membershipClient.GetUserMembership(getContextWithMetadata(c), &membership.GetUserMembershipRequest{
			UserId: id,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// 12. GET /users/:id/transactions -> GetCreditTransactions
	r.GET("/users/:id/transactions", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := membershipClient.GetCreditTransactions(getContextWithMetadata(c), &membership.GetCreditTransactionsRequest{
			UserId: id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp.Transactions)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func adminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetHeader("X-User-Role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
