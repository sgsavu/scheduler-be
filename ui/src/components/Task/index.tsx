import { memo, useState, useEffect, useCallback, MouseEventHandler, useMemo } from "react";
import { Task as TaskInterface, TaskStatus, TaskType } from "../../const";
import { dbGet, dbRemove, dbSet } from "../../state/localStorage";
import { downloadBlob } from "../../utils/downloadBlob";
import { formatDate } from "../../utils/formatDate";
import { Cross, Delete, Download, Tick, Train } from "../Icons";
import { Infer } from "../Icons";
import { DualRing } from "../Spinners";
import './index.css'

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

export const Task = memo<TaskProps>(function Task({
    id,
    data,
    status = 'Unknown status...',
    onDelete,
}) {
    const {
        ExpiryTime,
        FailureReason,
        Name,
        Status,
        result: _,
        ...expandedData
    } = data

    const [expanded, setExpanded] = useState<boolean>(false)
    const [elapsedTime, setElapsedTime] = useState<string>('')
    const [isDownloaded, setIsDownloaded] = useState<boolean>(false)

    const toggleExpanded = useCallback<MouseEventHandler>(() => { setExpanded(prev => !prev) }, [])

    useEffect(() => {
        dbGet(id)
            .then(({ result }) => {
                if (result) {
                    setIsDownloaded(true)
                }
            })
    }, [id])

    useEffect(() => {
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
    }, [data.CreationTime, data.TerminationTime])

    const onDownload = useCallback(() => {
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
    }, [id])

    const onLocalDelete = useCallback(() => {
        fetch(`v1/tasks/${id}`, { method: 'DELETE' })
        dbRemove(id)
        onDelete?.(id)
    }, [id, onDelete])

    const expired = useMemo(() => {
        if (!ExpiryTime || isDownloaded) { return false }

        return new Date(ExpiryTime).getTime() < Date.now()
    }, [ExpiryTime, isDownloaded])

    const colorTheme = expired ? 'expired' : statusToColor[Status]

    const arrowClassName = expanded ? 'arrow arrow-invert' : 'arrow'
    const expandedClassName = 'task-expanded ' + colorTheme
    const taskClassName = 'task-pill ' + colorTheme

    const taskStatus = useMemo(() => {
        if (Status === TaskStatus.FAILED) {
            return FailureReason
        }
        if (expired) {
            return 'Download has expired'
        }
        if (Status === TaskStatus.DONE) {
            return 'Download available'
        }
        return status
    }, [FailureReason, Status, expired, status])

    return (
        <div className="task">
            <button className={taskClassName} onClick={toggleExpanded}>
                <div className="task-pill-header">
                    {statusToIcon[Status]}
                    {taskTypeToIcon[data.Type]}
                    <span>{Name}</span>
                </div>
                <span className={arrowClassName}></span>
            </button>
            {expanded && (
                <div className={expandedClassName}>
                    {Object.entries(expandedData).map(([key, value]) => {
                        let formattedValue = value

                        if (key === 'CreationTime' || key === 'TerminationTime') {
                            formattedValue = formatDate(new Date(value))
                        }

                        return (
                            <div className="task-row">
                                {key}:
                                <span>{formattedValue}</span>
                            </div>
                        )
                    })}
                    <div>
                        <div className="task-row">
                            Elapsed time:
                            <span>
                                {elapsedTime}
                            </span>
                        </div>
                    </div>
                    <div className="terminal">
                        <div className="top">
                            <div className="btns">
                                <span className="circle red"></span>
                                <span className="circle yellow"></span>
                                <span className="circle green"></span>
                            </div>
                            <div className="title">Status: {Status}</div>
                        </div>
                        <pre className="body">
                            {taskStatus}
                        </pre>
                    </div>
                    <div className="task-buttons">
                        {Status === TaskStatus.DONE && (
                            <button
                                className="download-icon"
                                disabled={expired}
                                onClick={onDownload}
                            >
                                <Download />
                            </button>
                        )}
                        <button className="delete-icon" onClick={onLocalDelete}>
                            <Delete />
                        </button>
                    </div>
                </div>
            )}
        </div>
    )
})