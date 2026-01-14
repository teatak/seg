import * as React from "react"
import { cn } from "@/lib/utils"

const ScrollArea = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement> & {
        orientation?: "vertical" | "horizontal"
    }
>(({ className, children, orientation = "vertical", ...props }, ref) => (
    <div
        ref={ref}
        className={cn("relative overflow-hidden", className)}
        {...props}
    >
        <div
            className={cn(
                "h-full w-full overflow-auto scrollbar-thin scrollbar-thumb-border scrollbar-track-transparent",
                orientation === "horizontal" ? "overflow-x-auto" : "overflow-y-auto"
            )}
            style={{
                scrollbarWidth: "thin",
                scrollbarColor: "var(--border) transparent",
            }}
        >
            {children}
        </div>
    </div>
))
ScrollArea.displayName = "ScrollArea"

const ScrollBar = React.forwardRef<
    HTMLDivElement,
    React.HTMLAttributes<HTMLDivElement> & {
        orientation?: "vertical" | "horizontal"
    }
>(({ className, orientation = "vertical", ...props }, ref) => (
    <div
        ref={ref}
        className={cn(
            "flex touch-none select-none transition-colors",
            orientation === "vertical" &&
            "h-full w-2.5 border-l border-l-transparent p-px",
            orientation === "horizontal" &&
            "h-2.5 flex-col border-t border-t-transparent p-px",
            className
        )}
        {...props}
    />
))
ScrollBar.displayName = "ScrollBar"

export { ScrollArea, ScrollBar }
