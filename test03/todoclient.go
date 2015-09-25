type TodoClient struct {
    Jobs chan Job
}

func (c *TodoClient) GetTodos() ([]Todo, error) {
    arr := make([]Todo, 0)
    job := NewReadTodosJob()
    c.Jobs <- job
    if err := <-job.ExitChan(); err != nil {
        retrun arr, err
    }
    todos <-job.todos
    for _, value := range todos {
         arr = append(arr, value)
    }
    return arr, nil
}
