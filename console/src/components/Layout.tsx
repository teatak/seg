import { Link, useLocation, Outlet } from 'react-router-dom';
import { LayoutDashboard, Scissors, Library, Database } from 'lucide-react';
import { ThemeToggle } from './ThemeToggle';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarProvider,
    SidebarTrigger,
    SidebarInset,
    SidebarMenu,
    SidebarMenuItem,
    SidebarMenuButton,
    SidebarGroup,
    SidebarGroupContent,
} from "@/components/ui/sidebar"

export default function Layout() {
    const location = useLocation();

    const navItems = [
        { name: '首页', path: '/', icon: LayoutDashboard },
        { name: '分词测试', path: '/segment', icon: Scissors },
        { name: '语料学习', path: '/corpus', icon: Library },
        { name: '词库管理', path: '/dictionary', icon: Database },
    ];

    return (
        <SidebarProvider>
            <Sidebar>
                <SidebarHeader className="border-b border-sidebar-border p-5">
                    <h1 className="text-lg font-bold bg-gradient-to-r from-primary to-cyan-500 bg-clip-text text-transparent">
                        🧠 自学习分词系统
                    </h1>
                </SidebarHeader>
                <SidebarContent>
                    <SidebarGroup>
                        <SidebarGroupContent>
                            <SidebarMenu>
                                {navItems.map((item) => {
                                    const isActive = location.pathname === item.path;
                                    const Icon = item.icon;
                                    return (
                                        <SidebarMenuItem key={item.path}>
                                            <SidebarMenuButton
                                                asChild
                                                isActive={isActive}
                                                tooltip={item.name}
                                                size="lg"
                                            >
                                                <Link to={item.path}>
                                                    <Icon />
                                                    <span>{item.name}</span>
                                                </Link>
                                            </SidebarMenuButton>
                                        </SidebarMenuItem>
                                    );
                                })}
                            </SidebarMenu>
                        </SidebarGroupContent>
                    </SidebarGroup>
                </SidebarContent>
                <SidebarFooter className="border-t border-sidebar-border p-3">
                    <div className="flex items-center justify-between">
                        <span className="text-xs text-muted-foreground ml-2">v1.0.0</span>
                        <ThemeToggle />
                    </div>
                </SidebarFooter>
            </Sidebar>

            <SidebarInset>
                <header className="flex h-16 items-center gap-2 border-b px-4 md:hidden">
                    <SidebarTrigger />
                    <span className="font-semibold">自学习分词系统</span>
                </header>
                <div className="flex-1 overflow-auto p-6">
                    <div className="max-w-6xl mx-auto">
                        <Outlet />
                    </div>
                </div>
            </SidebarInset>
        </SidebarProvider>
    );
}
