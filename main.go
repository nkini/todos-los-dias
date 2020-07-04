package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
)

type Task struct {
	Name       string
	CreateTime time.Time
}

func insertTask(db *sql.DB, task string) {
	sqlStatement := `
		INSERT INTO task (name, create_time) 
		VALUES ($1, now()) 
	`
	if _, err := db.Exec(sqlStatement, task); err != nil {
		log.Fatalf("Error inserting into task: %q", err)
	}
	log.Print("task inserted successfully")
}

func getTasks(db *sql.DB) (tasks []Task) {

	rows, err := db.Query("SELECT name, create_time FROM task")
	if err != nil {
		log.Fatalf("Error reading task: %q", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var create_time time.Time

		err := rows.Scan(&name, &create_time)
		if err != nil {
			log.Fatalf("Error scanning row: %q", err)
			return
		}

		tasks = append(tasks, Task{Name: name, CreateTime: create_time})
	}
	return
}

func dbSetup(db *sql.DB) {
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS task (name varchar(500), create_time timestamp)"); err != nil {
		log.Fatalf("Error creating database table: %q", err)
		return
	}
	log.Print("dbSetup completed successfully")
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	dbSetup(db)

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", getTasks(db))
	})

	router.POST("/", func(c *gin.Context) {
		insertTask(db, c.PostForm("task"))
		c.HTML(http.StatusOK, "index.tmpl.html", getTasks(db))
	})

	router.Run(":" + port)
}
