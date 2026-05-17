package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "gym-membership/proto/asset"
)

func main() {
	assetServiceAddr := os.Getenv("ASSET_SERVICE_ADDR")
	if assetServiceAddr == "" {
		assetServiceAddr = "localhost:50051"
	}

	// Connect to Asset Service
	conn, err := grpc.Dial(assetServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to asset-service: %v", err)
	}
	defer conn.Close()
	assetClient := pb.NewAssetServiceClient(conn)

	r := gin.Default()

	// Phase 5: API Gateway implementation
	
	// GET /assets -> ListAvailableAssets
	r.GET("/assets", func(c *gin.Context) {
		assetType := c.Query("type")
		resp, err := assetClient.ListAvailableAssets(context.Background(), &pb.ListRequest{Type: assetType})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp.Assets)
	})

	// GET /assets/:id -> GetAsset
	r.GET("/assets/:id", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := assetClient.GetAsset(context.Background(), &pb.GetAssetRequest{Id: id})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	// GET /assets/:id/health -> returns health_score for Frontend dashboard
	r.GET("/assets/:id/health", func(c *gin.Context) {
		id := c.Param("id")
		resp, err := assetClient.GetHealthScore(context.Background(), &pb.GetAssetRequest{Id: id})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"health_score": resp.HealthScore})
	})

	// PATCH /assets/:id/status -> UpdateAssetStatus (admin only)
	r.PATCH("/assets/:id/status", adminOnly(), func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
			return
		}

		resp, err := assetClient.UpdateAssetStatus(context.Background(), &pb.UpdateStatusRequest{
			Id:     id,
			Status: input.Status,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
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
		// Simple role check from header
		role := c.GetHeader("X-User-Role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}
