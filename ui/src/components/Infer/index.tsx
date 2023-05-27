import { memo } from "react"
import { Info } from "../Icons";
import { useForm } from "../../utils/useForm";

const INDEX_RATIO_TOOLTIP =
    `This value determines how much of the index feature will be used in the model.`

const INFER_URL = '/v1/infer'

export const Infer = memo(function Infer() {
    const {
        status,
        errors,
        errorTimeout,
        popError,
        loading,
        onSubmit
    } = useForm(INFER_URL)

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
                    <button disabled={loading} type="submit">
                        {loading ? 'Loading...' : 'Infer'}
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