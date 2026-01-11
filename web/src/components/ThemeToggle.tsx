import { useTheme } from './ThemeProvider';
import { Sun, Moon, Monitor } from 'lucide-react';

export function ThemeToggle() {
    const { theme, setTheme } = useTheme();

    const options = [
        { value: 'light' as const, icon: Sun, label: '浅色' },
        { value: 'dark' as const, icon: Moon, label: '深色' },
        { value: 'system' as const, icon: Monitor, label: '系统' },
    ];

    return (
        <div className="flex items-center gap-1 p-1 bg-muted rounded-lg">
            {options.map(({ value, icon: Icon, label }) => (
                <button
                    key={value}
                    onClick={() => setTheme(value)}
                    className={`p-1.5 rounded transition-colors cursor-pointer ${theme === value
                        ? 'bg-card dark:bg-muted text-primary shadow-sm'
                        : 'text-muted-foreground dark:text-muted-foreground hover:text-foreground dark:hover:text-muted-foreground'
                        }`}
                    title={label}
                >
                    <Icon className="w-4 h-4" />
                </button>
            ))}
        </div>
    );
}
