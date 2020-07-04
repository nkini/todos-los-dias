package main

import (
	"database/sql"
	"fmt"
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

func insertTask(db *sql.DB, task string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := db.Exec("CREATE TABLE IF NOT EXISTS task (name varchar(500), create_time timestamp)"); err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error creating database table: %q", err))
			log.Fatalf("Error creating database table: %q", err)
			return
		}

		if _, err := db.Exec("INSERT INTO task VALUES (?, now())", task); err != nil {
			c.String(http.StatusInternalServerError,
				fmt.Sprintf("Error incrementing tick: %q", err))
			log.Fatalf("Error incrementing tick: %q", err)
			return
		}
	}
}

func getTasks(db *sql.DB) (tasks []Task) {

	rows, err := db.Query("SELECT name, create_time FROM task")
	if err != nil {
		log.Fatalf("Error reading tasks: %q", err)
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
		log.Print(name)
		log.Printf("%s", create_time)
	}
	return
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	log.Printf("Database URL: %s", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.POST("/", func(c *gin.Context) {
		insertTask(db, c.PostForm("task"))
		c.HTML(http.StatusOK, "index.tmpl.html", getTasks(db))
	})

	router.Run(":" + port)
}
