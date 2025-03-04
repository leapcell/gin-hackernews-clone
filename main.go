package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Post represents a post in the Hacker News clone
type Post struct {
	ID           int
	Title        string
	Link         string
	Host         string
	Content      string
	CreatedAt    time.Time
	CommentCount int
	Comments     []Comment
}

// Comment represents a comment on a post
type Comment struct {
	ID        int
	Content   string
	PostID    int
	CreatedAt time.Time
}

// createTable encapsulates the logic to create a table
// It checks if the table exists and creates it if not.
func createTable(db *sql.DB, tableName, createQuery string) error {
	var exists bool
	// SQL query to check if the table exists in the 'public' schema
	err := db.QueryRow(`
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = $1
        );
    `, tableName).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		// Create the table if it doesn't exist
		_, err := db.Exec(createQuery)
		if err != nil {
			return err
		}
		fmt.Printf("%s table created.\n", tableName)
	}
	return nil
}

// createTables creates all necessary tables
func createTables(db *sql.DB) error {
	// SQL query to create the 'posts' table
	postsTableQuery := `
        CREATE TABLE posts (
            id SERIAL PRIMARY KEY, -- Auto - incrementing primary key
            title VARCHAR(255) NOT NULL, -- Post title
            link VARCHAR(255) NOT NULL DEFAULT '', -- Post link
            content TEXT NOT NULL, -- Post content
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Creation time
        );
    `
	// SQL query to create the 'comments' table
	commentsTableQuery := `
        CREATE TABLE comments (
            id SERIAL PRIMARY KEY, -- Auto - incrementing primary key
            content TEXT NOT NULL, -- Comment content
            post_id INTEGER NOT NULL, -- ID of the related post
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Creation time
            FOREIGN KEY (post_id) REFERENCES posts(id) -- Foreign key referencing 'posts' table
        );
    `
	if err := createTable(db, "posts", postsTableQuery); err != nil {
		return err
	}
	if err := createTable(db, "comments", commentsTableQuery); err != nil {
		return err
	}
	return nil
}

// renderTemplate encapsulates the template rendering logic
func renderTemplate(c *gin.Context, tmplPath string, data interface{}) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func main() {
	// Database connection configuration
	// Use DSN from environment variable
	dsn := os.Getenv("PG_DSN")
	// Connect to the database using DSN
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		log.Fatal(err)
	}

	// Set up Gin router
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Define routes
	// Route to display the list of posts
	r.GET("/", func(c *gin.Context) {
		// SQL query to select posts ordered by creation time in descending order
		rows, err := db.Query("SELECT id, title, link, content, created_at FROM posts ORDER BY created_at DESC")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var posts []Post
		for rows.Next() {
			var post Post
			if err := rows.Scan(
				&post.ID,
				&post.Title,
				&post.Link,
				&post.Content,
				&post.CreatedAt,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			u, _ := url.Parse(post.Link)
			post.Host = u.Host

			// SQL query to count comments for each post
			var commentCount int
			if err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE post_id = $1", post.ID).Scan(&commentCount); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			post.CommentCount = commentCount

			posts = append(posts, post)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		renderTemplate(c, "templates/index.html", map[string]interface{}{
			"Posts": posts,
		})
	})

	// Route to add a new post
	r.POST("/new", func(c *gin.Context) {
		title := c.PostForm("title")
		content := c.PostForm("content")
		link := c.PostForm("link")
		// SQL query to insert a new post into the 'posts' table
		if _, err := db.Exec("INSERT INTO posts (title, content, link, created_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP)",
			title, content, link); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, "/")
	})

	// Route to display a single post and its comments
	r.GET("/post/:id", func(c *gin.Context) {
		id := c.Param("id")
		var post Post
		// SQL query to select a single post by ID
		if err := db.QueryRow("SELECT id, title, link, content, created_at FROM posts WHERE id = $1", id).Scan(
			&post.ID,
			&post.Title,
			&post.Link,
			&post.Content,
			&post.CreatedAt,
		); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// SQL query to select comments for a post ordered by creation time in descending order
		rows, err := db.Query("SELECT id, content, created_at FROM comments WHERE post_id = $1 ORDER BY created_at DESC", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var comments []Comment
		for rows.Next() {
			var comment Comment
			if err := rows.Scan(&comment.ID, &comment.Content, &comment.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			comment.PostID = post.ID
			comments = append(comments, comment)
		}
		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		post.Comments = comments

		renderTemplate(c, "templates/post_detail.html", map[string]interface{}{
			"Post": post,
		})
	})

	// Route to add a comment to a post
	r.POST("/post/:id/comment", func(c *gin.Context) {
		id := c.Param("id")
		content := c.PostForm("content")
		// SQL query to insert a new comment into the 'comments' table
		if _, err := db.Exec("INSERT INTO comments (content, post_id, created_at) VALUES ($1, $2, CURRENT_TIMESTAMP)", content, id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, "/post/"+id)
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server started on port %s", port)
	r.Run(":" + port)
}
