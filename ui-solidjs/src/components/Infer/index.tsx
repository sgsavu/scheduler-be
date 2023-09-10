import { useForm } from "../../utils/useForm";
import { Info } from "../Icons";

const INDEX_RATIO_TOOLTIP =
    `This value determines how much of the index feature will be used in the model.`

const INFER_URL = '/v1/infer'

export const Infer = () => {
     const {
         getErrors,
         getTimer,
         getStatus,
         getLoading,
         onSubmit,
         popError
    } = useForm(INFER_URL)

    return (
        <>
            <h1>Infer vocals</h1>
            <form onSubmit={onSubmit} class="form">
                <div class="row">
                    <label>Task name</label>
                    <input value="test" name="name"></input>
                </div>

                <label>Input files</label>
                <input
                    accept="audio/wav"
                    class="dropzone"
                    id="input"
                    multiple
                    name="input"
                    type="file">
                </input>

                <label>Model files</label>
                <input
                    accept="application/*,.pth,.index"
                    class="dropzone"
                    id="model"
                    multiple
                    name="model"
                    type="file">
                </input>

                <div class="row">
                    <div
                        aria-label={INDEX_RATIO_TOOLTIP}
                        class="info"
                        data-microtip-position="top"
                        role="tooltip"
                    >
                        <label>Index ratio</label>
                        <Info class="info-icon" />
                    </div>
                    <input
                        value={0.5}
                        name="indexRatio"
                        max={1.0}
                        min={0}
                        step={0.1}
                        type="number"
                    />
                </div>

                <div class="row">
                    <div
                        class="info"
                        data-microtip-position="top"
                        role="tooltip"
                    >
                        <label>Pitch</label>
                        <Info class="info-icon"/>
                    </div>
                    <input
                        value={4}
                        min={1}
                        name="pitch"
                        type="number"
                    />
                </div>

                <div class="submit-button-container">
                    <button disabled={getLoading()} type="submit">
                        {getLoading() ? 'Loading...' : 'Infer'}
                    </button>
                </div>

            </form>
             {Boolean(getErrors().length) && (
                <div class="error-bar">
                    <p>({getErrors().length}) {getErrors()[0]}</p>
                    <button onClick={popError}>Close ({getTimer()})</button>
                </div>
            )}
            {getStatus()}
        </>
    )
}