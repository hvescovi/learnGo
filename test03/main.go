package main

db := "./db.json"
jobs := make(chan Job)
go ProcessJobs(jobs, db)
client := &TodoClient{Jobs: jobs}
handlers := &TodoHandlers{Client: client}
r := gin.Default()
r.GET("/todo", handlers.GetTodos)
r.Run(":8080")

