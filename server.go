package main

import (
	"context"
	"github.com/Mohagames205/Golileo/skin"
	"github.com/Mohagames205/Golileo/util"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func main() {
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		AppName: "Golileo",
		Views:   engine,
	})

	no_cache_group := app.Group("/api/:username/img/no-cache", func(ctx *fiber.Ctx) error {
		return ctx.Next()
	})

	app.Use(cors.New())

	_ = godotenv.Load()

	app.Static("/cdn", "./public")

	if os.Getenv("AUTH") == "true" {
		app.Use(basicauth.New(basicauth.Config{
			Users: map[string]string{
				os.Getenv("USERNAME"): os.Getenv("PASSWORD"),
			},
			Unauthorized: func(c *fiber.Ctx) error {
				return c.Render("unauthorized", fiber.Map{})
			},
		}))
	}

	util.InitDatabase()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("3dtesting", fiber.Map{})
	})

	app.Post("/api/skin", func(ctx *fiber.Ctx) error {

		payload := skin.Skin{
			Username: `json:"username"`,
			Skin:     `json:"skin"`,
		}

		if err := ctx.BodyParser(&payload); err != nil {
			return err
		}

		skinCollection := util.Database().Collection("skins")

		opts := options.Update().SetUpsert(true)
		filter := bson.D{
			{"username", payload.Username},
		}
		update := bson.D{{"$set", bson.D{
			{Key: "username", Value: payload.Username},
			{Key: "skinstring", Value: payload.Skin}},
		}}

		/*
		 * We create the skin struct again using the special S function because the function also calculates the skinsize for us
		 * This is needed for the skinhead, without the calculation the head image saving will not work
		 */
		skinStruct := skin.S(payload.Username, payload.Skin)

		skinStruct.SaveFullImage("full")
		skinStruct.SaveHeadImage("head")

		_, err := skinCollection.UpdateOne(context.TODO(), filter, update, opts)

		if err != nil {
			log.Println(err)
		}

		return ctx.SendString("Skin has been inserted into the util")
	})

	app.Get("/api/:username/raw", func(ctx *fiber.Ctx) error {

		skinCollection := util.Database().Collection("skins")

		var skinResult bson.M
		err := skinCollection.FindOne(context.TODO(), bson.M{"username": ctx.Params("username")}).Decode(&skinResult)
		if err != nil {
			return fiber.NewError(404, "Username not found")
		}

		return ctx.JSON(fiber.Map{"username": ctx.Params("username"), "skinstring": skinResult["skinstring"]})
	})

	no_cache_group.Get("/head", func(ctx *fiber.Ctx) error {

		skinCollection := util.Database().Collection("skins")

		var skinResult bson.M
		err := skinCollection.FindOne(context.TODO(), bson.M{"username": ctx.Params("username")}).Decode(&skinResult)
		if err != nil {
			return fiber.NewError(404, "Username not found")
		}

		skinStruct := skin.S(skinResult["username"].(string), skinResult["skinstring"].(string))

		uuid := skin.PseudoUuid()
		skinStruct.SaveHeadImage(uuid)

		return ctx.JSON(fiber.Map{"url": "/cdn/images/" + skinStruct.Username + "-" + uuid + ".png"})
	})

	no_cache_group.Get("/full", func(ctx *fiber.Ctx) error {

		skinCollection := util.Database().Collection("skins")

		var skinResult bson.M
		err := skinCollection.FindOne(context.TODO(), bson.M{"username": ctx.Params("username")}).Decode(&skinResult)
		if err != nil {
			return fiber.NewError(404, "Username not found")
		}

		skinStruct := skin.S(skinResult["username"].(string), skinResult["skinstring"].(string))

		uuid := skin.PseudoUuid()
		skinStruct.SaveFullImage(uuid)

		return ctx.JSON(fiber.Map{"url": "/cdn/images/" + skinStruct.Username + "-" + uuid + ".png"})
	})

	app.Get("/api/:username/img/full", func(ctx *fiber.Ctx) error {
		skinCollection := util.Database().Collection("skins")

		var skinResult bson.M
		err := skinCollection.FindOne(context.TODO(), bson.M{"username": ctx.Params("username")}).Decode(&skinResult)
		if err != nil {
			return fiber.NewError(404, "Username not found")
		}

		skinStruct := skin.S(skinResult["username"].(string), skinResult["skinstring"].(string))

		workingDirectory, _ := os.Getwd()
		return ctx.SendFile(workingDirectory + "/public/images/" + skinStruct.Username + "-full.png")
	})

	app.Get("/api/:username/img/head", func(ctx *fiber.Ctx) error {
		skinCollection := util.Database().Collection("skins")

		var skinResult bson.M
		err := skinCollection.FindOne(context.TODO(), bson.M{"username": ctx.Params("username")}).Decode(&skinResult)
		if err != nil {
			return fiber.NewError(404, "Username not found")
		}

		skinStruct := skin.S(skinResult["username"].(string), skinResult["skinstring"].(string))

		workingDirectory, _ := os.Getwd()
		return ctx.SendFile(workingDirectory + "/public/images/" + skinStruct.Username + "-head.png")
	})

	log.Fatal(app.Listen(":3000"))
}
