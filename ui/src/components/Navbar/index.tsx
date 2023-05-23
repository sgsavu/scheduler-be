import { memo } from 'react'
import './index.css'

interface NavbarProps {
    items: Array<JSX.Element>
    onSelectedTabChange: (tab: number) => void
    selectedTab: number
}

export const Navbar = memo<NavbarProps>(function Navbar({
    items,
    onSelectedTabChange,
    selectedTab
}) {
    

    return (
        <div className="navbar">
            {items.map((item, i) => {
                const className = i === selectedTab ? 'navbar-item navbar-item-selected' : 'navbar-item'
                return (
                    <div
                        className={className}
                        onClick={() => onSelectedTabChange(i)}
                        key={i}
                    >
                        {item}
                    </div>
                )
            })}
        </div>
    )
})