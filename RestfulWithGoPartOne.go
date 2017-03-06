package main

import (
	"bytes"
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
		row := db.QueryRow("select id, first_name, last_name, username, email from user where id = ?;", id)
		err := row.Scan(&user.Id, &user.First_Name, &user.Last_Name, &user.Username, &user.Email)
		if err != nil {
			// If no results send null
			result = gin.H{
				"result": nil,
				"count":  0,
			}
		} else {
			result = gin.H{
				"result": user,
				"count":  1,
			}
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
		rows, err := db.Query("select id, first_name, last_name, username, email from user;")
		if err != nil {
			fmt.Print(err.Error())
		}
		for rows.Next() {
			err = rows.Scan(&user.Id, &user.First_Name, &user.Last_Name, &user.Username, &user.Email)
			users = append(users, user)
			if err != nil {
				fmt.Print(err.Error())
			}
		}
		defer rows.Close()
		c.JSON(http.StatusOK, gin.H{
			"result": users,
			"count":  len(users),
		})
	}
}

func newUserHandler(db *sql.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var buffer bytes.Buffer
		first_name := c.PostForm("first_name")
		last_name := c.PostForm("last_name")
		username := c.PostForm("username")
		email := c.PostForm("email")
		stmt, err := db.Prepare("insert into user (first_name, last_name, username, email) values(?,?,?,?);")
		if err != nil {
			fmt.Print(err.Error())
		}
		_, err = stmt.Exec(first_name, last_name, username, email)

		if err != nil {
			fmt.Print(err.Error())
		}

		//Appending strings via buffer , fast enough?
		buffer.WriteString(first_name)
		buffer.WriteString(" ")
		buffer.WriteString(last_name)
		defer stmt.Close()
		name := buffer.String()
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf(" %s successfully created", name),
		})
	}
}

func updateUserHandler(db *sql.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		var buffer bytes.Buffer
		id := c.Query("id")
		first_name := c.PostForm("first_name")
		last_name := c.PostForm("last_name")
		username := c.PostForm("username")
		email := c.PostForm("email")
		stmt, err := db.Prepare("update user set first_name= ?, last_name= ?, username=?, email=? where id= ?;")
		if err != nil {
			fmt.Print(err.Error())
		}
		_, err = stmt.Exec(first_name, last_name, username, email, id)
		if err != nil {
			fmt.Print(err.Error())
		}

		// Fastest way to append strings
		buffer.WriteString(first_name)
		buffer.WriteString(" ")
		buffer.WriteString(last_name)
		defer stmt.Close()
		name := buffer.String()
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Successfully updated to %s", name),
		})
	}
}

func deleteUserHandler(db *sql.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Query("id")
		stmt, err := db.Prepare("delete from user where id= ?;")
		if err != nil {
			fmt.Print(err.Error())
		}
		_, err = stmt.Exec(id)
		if err != nil {
			fmt.Print(err.Error())
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

	router := gin.Default()
	// Add API handlers here

	// GET individual user detail which includes {id, lastname, firstname, username and email}
	router.GET("/user/:id", userHandler(db))

	// GET all users stored
	router.GET("/users", usersHandler(db))

	// POST new user details
	router.POST("/user", newUserHandler(db))

	// PUT - update a user details
	router.PUT("/user", updateUserHandler(db))

	// Delete resources
	router.DELETE("/user", deleteUserHandler(db))

	router.Run(":3000")
}
