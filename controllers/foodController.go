package controller

import (
	"context"
	"golang-restaurant-management/database"
	"golang-restaurant-management/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
		projectStage := bson.D{
			{
				Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "total_count", Value: 1},
					{Key: "food_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
				},
			},
		}

		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while listing the food item!"})
			return
		}

		var allFoods []bson.M
		if err = result.All(ctx, &allFoods); err != nil {
			log.Fatal(err)
			return
		}

		c.JSON(http.StatusOK, allFoods[0])

	}
}

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		foodId := c.Param("food_id")
		var food models.Food

		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while fetching the food item"})
			return
		}

		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		}

		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Menu was not found!"})
		}
		defer cancel()

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()

		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Food item was not created!"})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, result)
	}
}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		foodId := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key:"name", Value: food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key:"price", Value: food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{Key:"food_image", Value: food.Food_image})
		}

		if food.Menu_id != nil {
			err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			defer cancel()

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Message : Menu was not found!"})
				return
			}

			updateObj = append(updateObj, bson.E{Key: "menu_id", Value: food.Menu_id})
		}

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})

		filter := bson.M{"food_id": foodId}
		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key:"$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Food item update failed!"})
		}

		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
