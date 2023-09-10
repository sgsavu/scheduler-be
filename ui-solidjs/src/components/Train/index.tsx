import { useForm } from "../../utils/useForm";
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

const TRAIN_URL = '/v1/train'

export const Train = () => {
    const {
        getErrors,
        getTimer,
        getLoading,
        onSubmit,
        popError
    } = useForm(TRAIN_URL)

    const errors = getErrors()
    const timer = getTimer()
    const loading = getLoading()
    const loadingText = loading ? 'Loading...' : 'Train'

    return (
        <>
            <h1>Train a model</h1>
            <form onSubmit={onSubmit} class="form">
                <div class="row">
                    <label>Task name</label>
                    <input value="test" name="name"></input>
                </div>

                <label>Trainset</label>
                <input
                    accept="audio/wav"
                    class="dropzone"
                    id="dataset"
                    multiple
                    name="dataset"
                    type="file">
                </input>

                <div class="row">
                    <label>Sample rate</label>
                    <div>
                        <select name="sampleRate">
                            <option value={32000}>32K</option>
                            <option value={40000}>40K</option>
                            <option value={48000}>48K</option>
                        </select>
                    </div>
                </div>

                <div class="row">
                    <div
                        class="info"
                    >
                        <label>Epochs</label>
                        <Info class="info-icon" />
                    </div>
                    <div popover id="wow">{EPOCH_TOOLTIP}</div>
                    <input
                        value={100}
                        name="epochs"
                        min={1}
                        type="number"
                        popovertarget="wow"

                    />
                </div>

                <div class="row">
                    <div
                        aria-label={BATCH_SIZE_TOOLTIP}
                        class="info"
                        data-microtip-position="top"
                        role="tooltip"
                    >
                        <label>Batch size</label>
                        <Info class="info-icon"></Info>
                    </div>
                    <input
                        value={4}
                        min={1}
                        name="batchSize"
                        type="number"
                    />
                </div>

                <div class="submit-button-container">
                    <button disabled={loading} type="submit">
                        {loadingText}
                    </button>
                </div>

            </form>
            {Boolean(errors.length) && (
                <div class="error-bar">
                    <p>({errors.length}) {errors[0]}</p>
                    <button onClick={popError}>Close ({timer})</button>
                </div>
            )}
            {status}
        </>
    )
}