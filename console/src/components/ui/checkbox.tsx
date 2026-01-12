import * as React from "react"
import { Check } from "lucide-react"
import { cn } from "@/lib/utils"

const Checkbox = React.forwardRef<
    HTMLButtonElement,
    React.ButtonHTMLAttributes<HTMLButtonElement> & {
        checked?: boolean
        onCheckedChange?: (checked: boolean) => void
    }
>(({ className, checked, onCheckedChange, ...props }, ref) => {
    return (
        <button
            type="button"
            role="checkbox"
            aria-checked={checked}
            data-state={checked ? "checked" : "unchecked"}
            onClick={() => onCheckedChange?.(!checked)}
            ref={ref}
            className={cn(
                "peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground cursor-pointer flex items-center justify-center",
                className
            )}
            {...props}
        >
            {checked && <Check className="h-3 w-3 stroke-[3]" />}
        </button>
    )
})
Checkbox.displayName = "Checkbox"

export { Checkbox }
