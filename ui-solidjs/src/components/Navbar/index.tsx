import { Accessor, Setter } from 'solid-js'
import {
    Infer as InferIcon,
    Train as TrainIcon,
    Tasks as TasksIcon
} from '../Icons'
import './index.css'

const NAV_ITEMS = [
    (
        <>
            <InferIcon />
            <span>Infer</span>
        </>
    ),
    (
        <>
            <TrainIcon />
            <span>Train</span>
        </>
    ),
    (
        <>
            <TasksIcon />
            <span>Tasks</span>
        </>
    ),
]

interface NavbarProps {
    setSelected?: Setter<number>
    getSelected?: Accessor<number>
}

export const Navbar = ({
    setSelected,
    getSelected
}: NavbarProps) =>
    <div class="navbar">
        {NAV_ITEMS.map((item, i) => {
            const className = i === getSelected?.() ? 'navbar-item navbar-item-selected' : 'navbar-item'
            return (
                <div
                    class={className}
                    onClick={() => setSelected?.(i)}
                >
                    {item}
                </div>
            )
        })}
    </div>