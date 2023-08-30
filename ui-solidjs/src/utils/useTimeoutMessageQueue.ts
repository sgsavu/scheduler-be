import { createEffect, createSignal, onCleanup } from "solid-js"

const DEFAULT_ERROR_TIMEOUT = 10

export const useTimerMessageQueue = () => {
    const [getMessages, setMessages] = createSignal<Array<string>>([])
    const [getTimer, setTimer] = createSignal(DEFAULT_ERROR_TIMEOUT)

    let timeout: number | null = null
    onCleanup(() => {
        if (timeout !== null) {
            clearInterval(timeout)
        }
    });

    createEffect(() => {
        if (getTimer() === 0) {
            pop()
        }
    });

    const push = (message: string) => {
        setMessages(prev => [...prev, message])
        setTimer(DEFAULT_ERROR_TIMEOUT)

        if (timeout === null) {
            timeout = setInterval(() => setTimer(prev => prev - 1), 1000)
        }
    }

    const pop = () => {
        setMessages(([, ...rest]) => rest)
        setTimer(DEFAULT_ERROR_TIMEOUT)

        const messages = getMessages()
        if (messages.length === 0 && timeout !== null) {
            clearInterval(timeout)
            timeout = null
        }
    }

    return [getMessages, getTimer, push, pop] as const
}