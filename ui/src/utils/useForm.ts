import { FormEventHandler, useCallback, useState } from "react"
import { useTimeoutMessageQueue } from "./useTimeoutMessageQueue"

export const useForm = (url: string) => {
    const [status, setStatus] = useState<string>('')
    const [errors, errorTimeout, pushError, popError] = useTimeoutMessageQueue()
    const [loading, setLoading] = useState<boolean>(false)

    const onSubmit = useCallback<FormEventHandler<HTMLFormElement>>(e => {
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

    }, [pushError, url])

    return {
        status,
        errors,
        errorTimeout,
        loading,
        onSubmit,
        popError
    }
}