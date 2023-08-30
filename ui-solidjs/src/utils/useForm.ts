import { createSignal } from "solid-js"
import { useTimerMessageQueue } from "./useTimeoutMessageQueue"

export const useForm = (url: string) => {
    const [getStatus, setStatus] = createSignal('')
    const [getLoading, setLoading] = createSignal(false)
    const [getErrors, getTimer, pushError, popError] = useTimerMessageQueue()

    const onSubmit = (e: Event) => {
        e.preventDefault()
        setLoading(true)

        const formData = new FormData(e.target as HTMLFormElement)

        for (const [key, value] of formData.entries()) {
            if (key === 'dataset') {
                const dataset = value as File

                if (dataset.size === 0) {
                    pushError('No files selected')
                    setLoading(false)
                    return
                }
                continue
            }

            if (value === '') {
                pushError(`Missing value for "${key}"`)
                setLoading(false)
                return
            }
        }

        fetch(url, {
            method: 'POST',
            body: formData
        })
            .then(res => {
                const { ok, status, statusText } = res
                if (!ok) {
                    throw new Error(`${status}: ${statusText}`)
                }
                return res.text()
            })
            .then(data => setStatus(`Task successfully scheduled with id: ${data}`))
            .catch(err => pushError(err.message))
            .finally(() => setLoading(false))

    }

    return {
        getStatus,
        getLoading,
        onSubmit,
        getErrors,
        getTimer,
        popError
    }
}