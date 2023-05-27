import { memo, useEffect, useState, useRef, useCallback } from "react"
import { Task } from "../Task";
import { dbGetAll, dbSet } from "../../state/localStorage";
import { Task as TaskInterface } from "../../const";
import "./index.css"

export const Tasks = memo(function Tasks() {
    const [tasks, setTasks] = useState<Record<string, TaskInterface>>({})
    const [status, setStatus] = useState<Record<string, string>>({})
    const eventSource = useRef<EventSource>()

    useEffect(() => {
        dbGetAll()
            .then(tasks => {
                const tasksFromStorage = tasks.reduce((acc: Record<string, TaskInterface>, curr: TaskInterface) => {
                    acc[curr.ID] = curr
                    return acc
                }, {})
                setTasks(tasksFromStorage)
            })

        eventSource.current = new EventSource(`/v1/tasks`)
        eventSource.current.addEventListener('onChange', ({ data }: MessageEvent<string>) => {
            const parsedData = JSON.parse(data) as Record<string, TaskInterface>
            const entries = Object.entries(parsedData)

            setTasks(prev => {
                const newTasks = { ...prev }
                entries.forEach(([key, value]) => {
                    newTasks[key] = value
                    dbSet(value)
                })

                return newTasks
            })

        })

        eventSource.current.addEventListener('onStatus', ({ data }: MessageEvent<string>) => {
            const parsedData = JSON.parse(data) as Record<string, string>
            const entries = Object.entries(parsedData)

            setStatus(prevStatus => {
                const newStatus = { ...prevStatus }

                entries.forEach(([key, value]) => {
                    newStatus[key] = value
                })

                return newStatus
            })
        })

        return () => { eventSource.current?.close() }
    }, [])

    const onDelete = useCallback((id: string) => {
        setTasks(prev => {
            const newTasks = { ...prev }
            delete newTasks[id]
            return newTasks
        })
    }, [])

    return (
        <div className="container">
            <h1>Tasks</h1>
            <div className="tasks">
                {Object.entries(tasks).map(([key, task]) =>
                    <Task
                        id={key}
                        key={key}
                        data={task}
                        status={status[key]}
                        onDelete={onDelete}
                    />
                )}
            </div>
        </div>
    )
});