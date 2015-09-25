type TodoHandlers struct {
    Client *TodoClient
}

func (h *TodoHandlers) GetTodos(C *gin.Context) {
    todos, err := h.Client.GetTodos()
    if err != nil {
        log.Print(err)
        c.JSON(500, "Internal error inside the serverrrr")
        return
    }
    c.JSON(200, todos)
}
