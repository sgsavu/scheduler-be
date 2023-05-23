import { forwardRef } from "react"
import "./index.css"

export const DualRing = forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
    (props, fwdRef) => (
        <div ref={fwdRef} className="lds-dual-ring" {...props}></div>
    )
)