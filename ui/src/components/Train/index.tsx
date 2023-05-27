import { memo, useState, useCallback, FormEventHandler } from "react"
import { useTimeoutMessageQueue } from "../../utils/useTimeoutMessageQueue";
import { Info } from "../Icons";
import './index.css'

const EPOCH_TOOLTIP =
    `The number of times the model will see the entire dataset.
It is not recommended to set this value too high, as it will 
cause the model to overfit and/or the returns will not be 
worth the processing power expended. The recommended 
range for a good model is between 100 and 1000.`
const BATCH_SIZE_TOOLTIP =
    `The number of samples that will be propagated through the network at once. 
This is a tradeoff between speed and accuracy. A larger batch size will result 
in faster training, but a smaller batch size will result in more accurate training. 
As a rule of thumb you should divide your GPU's VRAM by 1.2 and the resulting 
whole number is your batch size. Example: 12GB VRAM / 1.2 = 10.`

export const Train = memo(function Train() {
    const [status, setStatus] = useState<string>('')
    const [errors, errorTimeout, pushError, popError] = useTimeoutMessageQueue()
    const [loading, setLoading] = useState<boolean>(false)

    const onTrain = useCallback<FormEventHandler<HTMLFormElement>>(e => {
        e.preventDefault()
        setLoading(true)

        const formData = new FormData(e.target as HTMLFormElement)

        for (const [key, value] of formData.entries()) {
            if (key === 'dataset') {
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

        fetch('/v1/train', {
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

    }, [pushError])

    return (
        <>
            <h1>Train a model</h1>
            <form onSubmit={onTrain} className="form">
                <div className="row">
                    <label>Task name</label>
                    <input defaultValue="test" name="name"></input>
                </div>

                <label>Trainset</label>
                <input
                    accept="audio/wav"
                    className="dropzone"
                    id="dataset"
                    multiple
                    name="dataset"
                    type="file">
                </input>

                <div className="row">
                    <label>Sample rate</label>
                    <div>
                        <select name="sampleRate">
                            <option value={32000}>32K</option>
                            <option value={40000}>40K</option>
                            <option value={48000}>48K</option>
                        </select>
                    </div>
                </div>

                <div className="row">
                    <div
                        aria-label={EPOCH_TOOLTIP}
                        className="info"
                        data-microtip-position="top"
                        role="tooltip"
                    >
                        <label>Epochs</label>
                        <Info className="info-icon" />
                    </div>
                    <input
                        defaultValue={100}
                        name="epochs"
                        min={1}
                        type="number"
                    />
                </div>

                <div className="row">
                    <div
                        aria-label={BATCH_SIZE_TOOLTIP}
                        className="info"
                        data-microtip-position="top"
                        role="tooltip"
                    >
                        <label>Batch size</label>
                        <Info className="info-icon"></Info>
                    </div>
                    <input
                        defaultValue={4}
                        min={1}
                        name="batchSize"
                        type="number"
                    />
                </div>

                <div className="submit-button-container">
                    <button disabled={loading} type="submit">
                        {loading ? 'Loading...' : 'Train'}
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