import { Task as TaskInterface, TaskStatus, TaskType } from "../../const";
import { dbGet, dbRemove, dbSet } from "../../state/localStorage";
import { downloadBlob } from "../../utils/downloadBlob";
import { formatDate } from "../../utils/formatDate";
import { Cross, Delete, Download, Tick, Train } from "../Icons";
import { Infer } from "../Icons";
import { DualRing } from "../Spinners";
import './index.css'
import { createEffect, createSignal } from "solid-js";

interface TaskProps {
    id: string
    data: TaskInterface
    status?: string
    onDelete?: (id: string) => void
}

const taskTypeToIcon = {
    [TaskType.TRAINING]: <Train />,
    [TaskType.INFERRING]: <Infer />
}

const statusToIcon = {
    [TaskStatus.WORKING]: <DualRing />,
    [TaskStatus.DONE]: <Tick />,
    [TaskStatus.FAILED]: <Cross />
}

const statusToColor = {
    [TaskStatus.WORKING]: 'working',
    [TaskStatus.DONE]: 'done',
    [TaskStatus.FAILED]: 'failed'
}

const getTimeElapsed = (start: string, stop?: string): string => {
    const stopInMs = stop ? new Date(stop).getTime() : Date.now()
    const startInMs = new Date(start).getTime()
    const difInMs = stopInMs - startInMs

    const seconds = Math.floor(difInMs / 1000)
    const minutes = Math.floor(seconds / 60)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    const daysString = days > 0 ? `${days}d ` : ''
    const hoursString = hours > 0 ? `${hours % 24}h ` : ''
    const minutesString = minutes > 0 ? `${minutes % 60}m ` : ''
    const secondsString = seconds > 0 ? `${seconds % 60}s` : ''

    return daysString + hoursString + minutesString + secondsString
}

export const Task = ({
    id,
    data,
    status = 'Unknown status...',
    onDelete,
}: TaskProps) => {
    const {
        ExpiryTime,
        FailureReason,
        Name,
        Status,
        result: _,
        ...expandedData
    } = data

    const [getExpanded, setExpanded] = createSignal<boolean>(false)
    const [getElapsedTime, setElapsedTime] = createSignal<string>('')
    const [isDownloaded, setIsDownloaded] = createSignal<boolean>(false)

    const toggleExpanded = () => { setExpanded(prev => !prev) }

    createEffect(() => {
        dbGet(id)
            .then(({ result }) => {
                if (result) {
                    setIsDownloaded(true)
                }
            })
    })

    createEffect(() => {
        let interval: number

        if (data.TerminationTime) {
            setElapsedTime(getTimeElapsed(data.CreationTime, data.TerminationTime))
        } else {
            interval = setInterval(() => {
                setElapsedTime(getTimeElapsed(data.CreationTime))
            }, 1000)
        }

        return () => {
            if (interval) { clearInterval(interval) }
        }
    })

    const onDownload = () => {
        dbGet(id)
            .then(task => {
                const { result } = task

                if (result) {
                    downloadBlob(result, id)
                    return
                }

                fetch(`v1/tasks/${id}/result`)
                    .then(response => {
                        if (!response.ok) {
                            throw new Error(response.status + ' ' + response.statusText)
                        }
                        return response.blob()
                    })
                    .then(blob => {
                        dbGet(id)
                            .then(task => {
                                dbSet({ ...task, result: blob })
                                    .then(() => {
                                        fetch(`v1/tasks/${id}`, { method: 'DELETE' })
                                    })
                            })
                        downloadBlob(blob, id)
                    })
                    .catch(error => {
                        alert('Failed to download.' + error)
                    })
            })
    }

    const onLocalDelete = () => {
        fetch(`v1/tasks/${id}`, { method: 'DELETE' })
        dbRemove(id)
        onDelete?.(id)
    }

    const expired = () => {
        if (!ExpiryTime || isDownloaded()) { return false }

        return new Date(ExpiryTime).getTime() < Date.now()
    }

    const colorTheme = expired() ? 'expired' : statusToColor[Status]

    const arrowClassName = getExpanded() ? 'arrow arrow-invert' : 'arrow'
    const expandedClassName = 'task-expanded ' + colorTheme
    const taskClassName = 'task-pill ' + colorTheme

    const taskStatus = () => {
        if (Status === TaskStatus.FAILED) {
            return FailureReason
        }
        if (expired()) {
            return 'Download has expired'
        }
        if (Status === TaskStatus.DONE) {
            return 'Download available'
        }
        return status
    }

    return (
        <div class="task">
            <button class={taskClassName} onClick={toggleExpanded}>
                <div class="task-pill-header">
                    {statusToIcon[Status]}
                    {taskTypeToIcon[data.Type]}
                    <span>{Name}</span>
                </div>
                <span class={arrowClassName}></span>
            </button>
            {getExpanded() && (
                <div class={expandedClassName}>
                    {Object.entries(expandedData).map(([key, value]) => {
                        let formattedValue = value

                        if (key === 'CreationTime' || key === 'TerminationTime') {
                            formattedValue = formatDate(new Date(value))
                        }

                        return (
                            <div class="task-row">
                                {key}:
                                <span>{formattedValue}</span>
                            </div>
                        )
                    })}
                    <div>
                        <div class="task-row">
                            Elapsed time:
                            <span>
                                {getElapsedTime()}
                            </span>
                        </div>
                    </div>
                    <div class="terminal">
                        <div class="top">
                            <div class="btns">
                                <span class="circle red"></span>
                                <span class="circle yellow"></span>
                                <span class="circle green"></span>
                            </div>
                            <div class="title">Status: {Status}</div>
                        </div>
                        <pre class="body">
                            {taskStatus()}
                        </pre>
                    </div>
                    <div class="task-buttons">
                        {Status === TaskStatus.DONE && (
                            <button
                                class="download-icon"
                                disabled={expired()}
                                onClick={onDownload}
                            >
                                <Download />
                            </button>
                        )}
                        <button class="delete-icon" onClick={onLocalDelete}>
                            <Delete />
                        </button>
                    </div>
                </div>
            )}
        </div>
    )
}