import { useState, useCallback, useEffect, useRef } from "react"

const DEFAULT_ERROR_TIMEOUT = 10

export const useTimeoutMessageQueue = () => {
    const [errors, setErrors] = useState<Array<string>>([])
    const [errorTimeout, setErrorTimeout] = useState<number>(DEFAULT_ERROR_TIMEOUT)
    const timeout = useRef<number | undefined>(undefined)

    useEffect(() => {
        if (errorTimeout === 0) {
            pop()
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [errorTimeout])

    const push = useCallback((message: string) => {
        setErrors(prev => [...prev, message])
        if (!timeout.current) {
            timeout.current = setInterval(() => setErrorTimeout(prev => prev - 1), 1000)
        }
    }, [])

    const pop = useCallback(() => {
        setErrors(([, ...rest]) => rest)
        if (errors.length >= 1) {
            setErrorTimeout(DEFAULT_ERROR_TIMEOUT)
        }
        if (errors.length === 1) {
            clearInterval(timeout.current)
            timeout.current = undefined
        }
    }, [errors.length])

    return [errors, errorTimeout, push, pop] as const
}