package main

import (
	"database/sql"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type todo struct {
	Item     string
	Priority int // Range 0..9, 0=highest, 9=lowest
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

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("main - db.Ping failed: %v", err)
	}
	log.Println("db connection verified")

	if err := dbInit(db); err != nil {
		log.Fatalf("main - dbInit failed: %v", err)
	}

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
	var res todo
	var todos []todo
	rows, err := db.Query("SELECT * FROM todos")
	defer rows.Close()
	if err != nil {
		log.Fatalf("indexHandler - db.Query failed: %v", err)
		c.JSON("An error occured")
	}
	for rows.Next() {
		rows.Scan(&res.Item, &res.Priority)
		todos = append(todos, res)
	}
	log.Printf("indexHandler - todos: %q", todos)
	return c.Render("index", fiber.Map{
		"Todos": todos,
	})
}

func postHandler(c *fiber.Ctx, db *sql.DB) error {
	type todor struct{ Item, Priority string }
	rawTodo := todor{}
	if err := c.BodyParser(&rawTodo); err != nil {
		log.Printf("An error occured: %v", err)
		return c.SendString(err.Error())
	}
	log.Printf("postHandler - rawTodo: %v", rawTodo)
	if rawTodo.Item != "" {
		priority := 9
		if num, err := strconv.Atoi(rawTodo.Priority); err == nil && num >= 0 && num <= 9 {
			// TODO: abort on illegal priority
			priority = num
		}
		newTodo := todo{Item: rawTodo.Item, Priority: priority}
		log.Printf("postHandler - newTodo: %v", newTodo)
		if _, err := db.Exec("INSERT into todos (item, priority) VALUES ($1, $2)", newTodo.Item, newTodo.Priority); err != nil {
			log.Fatalf("postHandler - db.Exec failed to insert %q: %v", newTodo, err)
		}
	} else {
		log.Printf("postHandler - newTodo.Item is empty")
	}
	return c.Redirect("/")
}

// TODO: Implement priority update capability as putHandler currently only updates Item
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

func dbInit(db *sql.DB) error {
	columQueryMap := map[string]string{
		"item":     "ALTER TABLE todos ADD COLUMN item VARCHAR(255);",
		"priority": "ALTER TABLE todos ADD COLUMN priority INTEGER DEFAULT 9;",
	}
	columnNames := slices.Collect(maps.Keys(columQueryMap))
	columnCount := len(columnNames)
	existCount := 0
	createdCount := 0

	for _, columnName := range columnNames {
		exists, err := checkColumnExistsPostgreSQL(db, "todos", columnName)
		if err != nil {
			log.Printf("dbInit - checkColumnExistsPostgreSQL failed for columnName %q: %v", columnName, err)
			return err
		}
		if exists {
			log.Printf("dbInit - columnName %q (exists)", columnName)
			existCount++
			continue
		}
		queryText := columQueryMap[columnName]
		log.Printf("dbInit - queryText %q", queryText)
		if _, err := db.Exec(queryText); err != nil {
			log.Printf("dbInit - db.Exec failed for columnName %q: %v", columnName, err)
			return err
		}
		createdCount++
		log.Printf("dbInit - columnName %q - (created)", columnName)
	}

	log.Printf("dbInit - column counts - total: %d, exists: %d, created: %d", columnCount, existCount, createdCount)
	return nil
}

func checkColumnExistsPostgreSQL(db *sql.DB, tableName, columnName string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = $1 AND COLUMN_NAME = $2`

	var count int
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
