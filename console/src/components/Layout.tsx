import { Link, useLocation, Outlet } from 'react-router-dom';
import {
    Home,
    Type,
    BookOpen,
    FileText,
    ChevronRight,
} from 'lucide-react';
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
    SidebarGroupLabel,
    useSidebar,
} from "@/components/ui/sidebar"

const navItems = [
    { name: '首页', path: '/', icon: Home },
    { name: '分词测试', path: '/segment', icon: Type },
    { name: '语料学习', path: '/corpus', icon: BookOpen },
    { name: '词库管理', path: '/dictionary', icon: FileText },
];

function NavMenu() {
    const location = useLocation();
    const { setOpenMobile, isMobile } = useSidebar();

    const handleNavClick = () => {
        if (isMobile) {
            setOpenMobile(false);
        }
    };

    return (
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
                            <Link to={item.path} onClick={handleNavClick}>
                                <Icon className="h-4 w-4" />
                                <span>{item.name}</span>
                            </Link>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                );
            })}
        </SidebarMenu>
    );
}

export default function Layout() {
    const location = useLocation();

    const getPageTitle = () => {
        const item = navItems.find(i => i.path === location.pathname);
        return item?.name || '首页';
    };

    return (
        <SidebarProvider>
            <Sidebar className="border-r border-sidebar-border">
                {/* Header */}
                <SidebarHeader className="border-b border-sidebar-border p-4">
                    <div className="flex items-center gap-3">
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-primary/80 text-primary-foreground text-sm font-semibold">
                            🧠
                        </div>
                        <div>
                            <p className="text-sm font-semibold leading-none">自学习分词系统</p>
                            <p className="text-xs text-muted-foreground mt-0.5">v1.0.0</p>
                        </div>
                    </div>
                </SidebarHeader>

                <SidebarContent>
                    <SidebarGroup>
                        <SidebarGroupLabel className="text-xs text-muted-foreground px-3 py-2">
                            导航菜单
                        </SidebarGroupLabel>
                        <SidebarGroupContent>
                            <NavMenu />
                        </SidebarGroupContent>
                    </SidebarGroup>
                </SidebarContent>

                {/* Footer */}
                <SidebarFooter className="border-t border-sidebar-border p-3">
                    <div className="flex items-center justify-center">
                        <span className="text-xs text-muted-foreground">© 2026</span>
                    </div>
                </SidebarFooter>
            </Sidebar>

            <SidebarInset className="flex flex-col">
                {/* Fixed Header */}
                <header className="fixed top-0 right-0 left-0 md:left-[var(--sidebar-width)] z-10 flex h-[65px] items-center gap-4 border-b border-sidebar-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4 md:peer-data-[state=collapsed]:left-[var(--sidebar-width-icon)] transition-[left] duration-200 ease-linear">
                    <SidebarTrigger className="md:hidden" />

                    {/* Breadcrumb */}
                    <nav className="hidden md:flex items-center gap-1.5 text-sm">
                        <span className="text-muted-foreground">自学习分词系统</span>
                        <ChevronRight className="h-4 w-4 text-muted-foreground" />
                        <span className="font-medium">{getPageTitle()}</span>
                    </nav>

                    <span className="font-semibold md:hidden">{getPageTitle()}</span>

                    {/* Spacer */}
                    <div className="flex-1" />

                    {/* Theme Toggle */}
                    <ThemeToggle />
                </header>

                {/* Main Content */}
                <div className="flex-1 overflow-auto p-3 md:p-6 pt-[calc(65px+0.75rem)] md:pt-[calc(65px+1.5rem)]">
                    <div className="max-w-7xl mx-auto">
                        <Outlet />
                    </div>
                </div>
            </SidebarInset>
        </SidebarProvider>
    );
}
