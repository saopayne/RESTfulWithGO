package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	Id         int
	First_Name string
	Last_Name  string
	Username   string
	Email      string
}

func userHandler(db *sql.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var (
			user   User
			result gin.H
		)
		id := c.Param("id")

		row := db.QueryRow("select id, first_name, last_name, username, email from user where id = ?", id)
		err := row.Scan(&user.Id, &user.First_Name, &user.Last_Name, &user.Username, &user.Email)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		result = gin.H{
			"user": user,
		}
		c.JSON(http.StatusOK, result)
	}
}

func usersHandler(db *sql.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var (
			user  User
			users []User
		)

		rows, err := db.Query("select id, first_name, last_name, username, email from user")
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&user.Id, &user.First_Name, &user.Last_Name, &user.Username, &user.Email)
			users = append(users, user)
			if err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"count": len(users),
		})
	}
}

func newUserHandler(stmt *sql.Stmt) func(*gin.Context) {
	return func(c *gin.Context) {
		first_name := c.PostForm("first_name")
		last_name := c.PostForm("last_name")
		username := c.PostForm("username")
		email := c.PostForm("email")

		_, err := stmt.Exec(first_name, last_name, username, email)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		//Appending strings via buffer , fast enough?
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%s %s successfully created", first_name, last_name),
		})
	}
}

func updateUserHandler(stmt *sql.Stmt) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Query("id")
		first_name := c.PostForm("first_name")
		last_name := c.PostForm("last_name")
		username := c.PostForm("username")
		email := c.PostForm("email")

		_, err := stmt.Exec(first_name, last_name, username, email, id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully updated to %s %s", first_name, last_name),
		})
	}
}

func deleteUserHandler(stmt *sql.Stmt) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Query("id")

		_, err := stmt.Exec(id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully deleted user: %s", id),
		})
	}
}

func main() {
	//xxxx = mysql username
	//yyyy = mysql password
	db, err := sql.Open("mysql", "root:saopayne@tcp(127.0.0.1:3306)/gosample")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// make sure connection is available
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	// Check if the connection is active
	_, err = db.Exec("SELECT 1")
	if err != nil {
		log.Fatalln(err)
	}

	// Reuse prepared statements in order to speed up code execution in the database
	newUserStmt, err := db.Prepare("insert into user (first_name, last_name, username, email) values(?,?,?,?)")
	if err != nil {
		log.Fatalln(err)
	}
	defer newUserStmt.Close()

	updateUserStmt, err := db.Prepare("update user set first_name= ?, last_name= ?, username=?, email=? where id= ?")
	if err != nil {
		log.Fatalln(err)
	}
	defer updateUserStmt.Close()

	deleteUserStmt, err := db.Prepare("delete from user where id= ?")
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer deleteUserStmt.Close()

	router := gin.Default()
	// Add API handlers here

	// GET individual user detail which includes {id, lastname, firstname, username and email}
	router.GET("/user/:id", userHandler(db))

	// GET all users stored
	router.GET("/users", usersHandler(db))

	// POST new user details
	router.POST("/user", newUserHandler(newUserStmt))

	// PUT - update a user details
	router.PUT("/user", updateUserHandler(updateUserStmt))

	// Delete resources
	router.DELETE("/user", deleteUserHandler(deleteUserStmt))

	log.Fatalln(router.Run(":3000"))
}
