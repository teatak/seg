import { Link, useLocation, Outlet } from 'react-router-dom';
import { LayoutDashboard, Scissors, Library, Database } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ThemeToggle } from './ThemeToggle';

export default function Layout() {
    const location = useLocation();

    const navItems = [
        { name: '首页', path: '/', icon: LayoutDashboard },
        { name: '分词测试', path: '/segment', icon: Scissors },
        { name: '语料学习', path: '/corpus', icon: Library },
        { name: '词库管理', path: '/dictionary', icon: Database },
    ];

    return (
        <div className="flex h-screen bg-background">
            <aside className="w-60 bg-sidebar text-sidebar-foreground border-r border-sidebar-border flex flex-col">
                <div className="p-5 border-b border-sidebar-border">
                    <h1 className="text-lg font-bold bg-gradient-to-r from-primary to-cyan-500 bg-clip-text text-transparent">
                        🧠 自学习分词系统
                    </h1>
                </div>
                <nav className="flex-1 p-3 space-y-1">
                    {navItems.map((item) => {
                        const isActive = location.pathname === item.path;
                        const Icon = item.icon;
                        return (
                            <Link
                                key={item.path}
                                to={item.path}
                                className={cn(
                                    "flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all",
                                    isActive
                                        ? "bg-sidebar-primary text-sidebar-primary-foreground"
                                        : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                                )}
                            >
                                <Icon className="w-4 h-4" />
                                {item.name}
                            </Link>
                        );
                    })}
                </nav>
                <div className="p-3 border-t border-sidebar-border flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">v1.0.0</span>
                    <ThemeToggle />
                </div>
            </aside>

            <main className="flex-1 overflow-auto">
                <div className="p-6 max-w-6xl mx-auto">
                    <Outlet />
                </div>
            </main>
        </div>
    );
}
