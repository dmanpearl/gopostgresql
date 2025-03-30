package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type todo struct {
	Item string
}

func main() {
	// Add line numbers and filenames to logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	connStr := "postgresql://postgres:DaVinci@localhost/todos?sslmode=disable"
	log.Printf("connStr: %q", connStr)
	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("main - failed to open db: %v", err)
	}
	log.Println("db opened")

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		log.Println("Get...")
		return indexHandler(c, db)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		log.Println("Post...")
		return postHandler(c, db)
	})

	app.Put("/update", func(c *fiber.Ctx) error {
		log.Println("Put...")
		return putHandler(c, db)
	})

	app.Delete("/delete", func(c *fiber.Ctx) error {
		log.Println("Delete...")
		return deleteHandler(c, db)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Static("/", "./public")
	log.Fatalln(app.Listen(fmt.Sprintf(":%v", port)))
}

func indexHandler(c *fiber.Ctx, db *sql.DB) error {
	var res string
	var todos []string
	rows, err := db.Query("SELECT * FROM todos")
	defer rows.Close()
	if err != nil {
		log.Fatalf("indexHandler - db.Query failed: %v", err)
		c.JSON("An error occured")
	}
	for rows.Next() {
		rows.Scan(&res)
		todos = append(todos, res)
	}
	log.Printf("indexHandler - todos: %q", todos)
	return c.Render("index", fiber.Map{
		"Todos": todos,
	})
}

func postHandler(c *fiber.Ctx, db *sql.DB) error {
	newTodo := todo{}
	if err := c.BodyParser(&newTodo); err != nil {
		log.Printf("An error occured: %v", err)
		return c.SendString(err.Error())
	}
	log.Printf("%v", newTodo)
	if newTodo.Item != "" {
		if _, err := db.Exec("INSERT into todos VALUES ($1)", newTodo.Item); err != nil {
			log.Fatalf("postHandler - db.Exec failed to insert %q: %v", newTodo.Item, err)
		}
	}
	return c.Redirect("/")
}

func putHandler(c *fiber.Ctx, db *sql.DB) error {
	olditem := c.Query("olditem")
	newitem := c.Query("newitem")
	log.Printf("putHandler - olditem: %q, newitem: %q", olditem, newitem)
	if _, err := db.Exec("UPDATE todos SET item=$1 WHERE item=$2", newitem, olditem); err != nil {
		log.Fatalf("putHandler - db.Exec failed to update: %v", err)
	}
	log.Println("putHandler - update complete")
	return c.SendString("renamed")
}

func deleteHandler(c *fiber.Ctx, db *sql.DB) error {
	todoToDelete := c.Query("item")
	log.Printf("deleteHandler - todoToDelete: %q", todoToDelete)
	if _, err := db.Exec("DELETE from todos WHERE item=$1", todoToDelete); err != nil {
		log.Fatalf("deleteHandler - db.Exec failed to delete %q: %v", todoToDelete, err)
	}
	log.Println("deleteHandler - delete complete")
	return c.SendString("deleted")
}
