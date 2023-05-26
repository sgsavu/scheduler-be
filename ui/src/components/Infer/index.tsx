import { memo, useState, useCallback, FormEventHandler } from "react"
import { useTimeoutMessageQueue } from "../../utils/useTimeoutMessageQueue";
import { Info } from "../Icons";

const INDEX_RATIO_TOOLTIP =
    `This value determines how much of the index feature will be used in the model.`

export const Infer = memo(function Infer() {
    const [status, setStatus] = useState<string>('')
    const [errors, errorTimeout, pushError, popError] = useTimeoutMessageQueue()

    const onSubmit = useCallback<FormEventHandler<HTMLFormElement>>(e => {
        e.preventDefault()

        const formData = new FormData(e.target as HTMLFormElement)

        for (const [key, value] of formData.entries()) {
            if (key === 'model' || key === 'input') {
                const dataset = value as File

                if (dataset.size === 0) {
                    pushError('No files selected')
                    return
                }
                continue
            }

            if (value === '') {
                pushError(`Missing value for "${key}"`)
                return
            }
        }

        fetch('/v1/infer', {
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

    }, [pushError])

    return (
        <>
            <h1>Infer vocals</h1>
            <form onSubmit={onSubmit} className="form">
                <div className="row">
                    <label>Task name</label>
                    <input defaultValue="test" name="name"></input>
                </div>

                <label>Input files</label>
                <input
                    accept="audio/wav"
                    className="dropzone"
                    id="input"
                    multiple
                    name="input"
                    type="file">
                </input>

                <label>Model files</label>
                <input
                    accept="application/*,.pth,.index"
                    className="dropzone"
                    id="model"
                    multiple
                    name="model"
                    type="file">
                </input>

                <div className="row">
                    <div
                        aria-label={INDEX_RATIO_TOOLTIP}
                        className="info"
                        data-microtip-position="top"
                        role="tooltip"
                    >
                        <label>Index ratio</label>
                        <Info className="info-icon" />
                    </div>
                    <input
                        defaultValue={0.5}
                        name="indexRatio"
                        max={1.0}
                        min={0}
                        step={0.1}
                        type="number"
                    />
                </div>

                <div className="row">
                    <div
                        className="info"
                        data-microtip-position="top"
                        role="tooltip"
                    >
                        <label>Pitch</label>
                        <Info className="info-icon"></Info>
                    </div>
                    <input
                        defaultValue={4}
                        min={1}
                        name="pitch"
                        type="number"
                    />
                </div>

                <div className="submit-button-container">
                    <button type="submit">
                        Infer
                    </button>
                </div>

            </form>
            {Boolean(errors.length) && (
                <div className="error-bar">
                    <p>({errors.length}) {errors[0]}</p>
                    <button onClick={popError}>Close ({errorTimeout})</button>
                </div>
            )}
            {status}
        </>
    )
})