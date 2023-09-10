import { Task } from "../Task";
import { dbGetAll, dbSet } from "../../state/localStorage";
import { Task as TaskInterface } from "../../const";
import "./index.css"
import { createEffect, createSignal } from "solid-js";

export const Tasks = () => {
    const [getTasks, setTasks] = createSignal<Record<string, TaskInterface>>({})
    const [getStatus, setStatus] = createSignal<Record<string, string>>({})

    createEffect(() => {
        let eventSource = new EventSource(`/v1/tasks`)

        dbGetAll()
            .then(tasks => {
                const tasksFromStorage = tasks.reduce((acc: Record<string, TaskInterface>, curr: TaskInterface) => {
                    acc[curr.ID] = curr
                    return acc
                }, {})
                setTasks(tasksFromStorage)
            })

        eventSource.addEventListener('onChange', ({ data }: MessageEvent<string>) => {
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

        eventSource.addEventListener('onStatus', ({ data }: MessageEvent<string>) => {
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

        return () => { eventSource.close() }
    }, [])

    const onDelete = (id: string) => {
        setTasks(prev => {
            const newTasks = { ...prev }
            delete newTasks[id]
            return newTasks
        })
    }

    return (
        <div class="container">
            <h1>Tasks</h1>
            <div class="tasks">
                {Object.entries(getTasks()).map(([key, task]) =>
                    <Task
                        id={key}
                        data={task}
                        status={getStatus()[key]}
                        onDelete={onDelete}
                    />
                )}
            </div>
        </div>
    )
}