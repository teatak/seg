import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
    RefreshCw,
    Database,
    MessageSquare,
    Activity,
    Users,
    BarChart3,
} from 'lucide-react';
import { getApiPath } from '@/lib/api';

interface Stats {
    dictionary: { base_words: number; user_words: number; staging_words: number; total: number };
    feedback: { total: number; total_new_words: number };
    requests?: { total: number; since_last: number; mine_threshold: number };
}

export default function Dashboard() {
    const [stats, setStats] = useState<Stats | null>(null);
    const [loading, setLoading] = useState(false);

    const fetchStats = async () => {
        setLoading(true);
        try {
            const res = await fetch(getApiPath('stats'));
            setStats(await res.json());
            await new Promise(r => setTimeout(r, 300));
        } catch (e) { console.error(e); }
        finally { setLoading(false); }
    };

    useEffect(() => { fetchStats(); }, []);

    const statCards = [
        {
            title: '基础词库',
            value: stats?.dictionary?.base_words?.toLocaleString() || '0',
            icon: Database,
            gradient: 'from-blue-500/20 to-blue-600/10',
            iconColor: 'text-blue-600 dark:text-blue-400'
        },
        {
            title: '用户词库',
            value: stats?.dictionary?.user_words?.toLocaleString() || '0',
            icon: Users,
            gradient: 'from-emerald-500/20 to-emerald-600/10',
            iconColor: 'text-emerald-600 dark:text-emerald-400'
        },
        {
            title: '暂存词库',
            value: stats?.dictionary?.staging_words?.toLocaleString() || '0',
            icon: BarChart3,
            gradient: 'from-amber-500/20 to-amber-600/10',
            iconColor: 'text-amber-600 dark:text-amber-400'
        },
        {
            title: '用户反馈',
            value: stats?.feedback?.total?.toLocaleString() || '0',
            icon: MessageSquare,
            gradient: 'from-violet-500/20 to-violet-600/10',
            iconColor: 'text-violet-600 dark:text-violet-400'
        },
    ];

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">首页</h1>
                    <p className="text-muted-foreground mt-1">系统概览与统计数据</p>
                </div>
                <Button variant="outline" onClick={fetchStats} disabled={loading} size="sm">
                    <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
                    刷新
                </Button>
            </div>

            {/* Stats Row */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {statCards.map((card) => {
                    const Icon = card.icon;
                    return (
                        <Card key={card.title} className="overflow-hidden group cursor-pointer">
                            <CardHeader className="flex flex-row items-center justify-between pb-2 pt-4 px-4">
                                <CardTitle className="text-sm font-medium text-muted-foreground">
                                    {card.title}
                                </CardTitle>
                                <div className={`p-2 rounded-lg bg-gradient-to-br ${card.gradient} transition-transform group-hover:scale-110`}>
                                    <Icon className={`h-4 w-4 ${card.iconColor}`} />
                                </div>
                            </CardHeader>
                            <CardContent className="px-3 pb-3 md:px-4 md:pb-4">
                                <span className="text-2xl font-bold tracking-tight">{card.value}</span>
                            </CardContent>
                        </Card>
                    );
                })}
            </div>

            {/* Learning Status Card */}
            <Card>
                <CardHeader className="pb-3">
                    <CardTitle className="flex items-center gap-2 text-lg">
                        <Activity className="w-5 h-5 text-primary" />
                        自动学习状态
                    </CardTitle>
                </CardHeader>
                <CardContent className="p-3 md:p-6">
                    <div className="flex items-center gap-4">
                        <div className="flex-1 bg-muted rounded-full h-2.5 overflow-hidden">
                            <div
                                className="bg-gradient-to-r from-primary to-primary/60 h-full transition-all duration-500"
                                style={{ width: `${Math.min(100, ((stats?.requests?.since_last || 0) / (stats?.requests?.mine_threshold || 100)) * 100)}%` }}
                            />
                        </div>
                        <span className="text-sm text-muted-foreground whitespace-nowrap font-medium">
                            {stats?.requests?.since_last || 0} / {stats?.requests?.mine_threshold || 100}
                        </span>
                    </div>
                    <p className="text-sm text-muted-foreground mt-2">
                        距下次自动挖掘还需积累 <span className="font-semibold text-foreground">{Math.max(0, (stats?.requests?.mine_threshold || 100) - (stats?.requests?.since_last || 0))}</span> 条请求
                    </p>
                </CardContent>
            </Card>

            {/* Dictionary Explanation Cards */}
            <div className="grid gap-4 md:grid-cols-3">
                <Card className="group hover:border-blue-500/50 transition-colors">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base flex items-center gap-2">
                            <span className="text-xl">📚</span> 基础词库 (Base)
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="p-3 md:p-6 text-sm text-muted-foreground">
                        系统内置的通用核心词典，包含几十万标准中文词汇。它是分词的基础，只读不可修改。
                    </CardContent>
                </Card>
                <Card className="group hover:border-emerald-500/50 transition-colors">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base flex items-center gap-2">
                            <span className="text-xl">🌲</span> 用户词库 (User)
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="p-3 md:p-6 text-sm text-muted-foreground">
                        <span className="text-emerald-600 dark:text-emerald-400 font-medium">最高优先级</span>。包含您人工添加或从暂存区归档的词语。用于覆盖基础分词结果，修正特定领域的术语。
                    </CardContent>
                </Card>
                <Card className="group hover:border-amber-500/50 transition-colors">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-base flex items-center gap-2">
                            <span className="text-xl">🧪</span> 暂存词库 (Staging)
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="p-3 md:p-6 text-sm text-muted-foreground">
                        <span className="text-amber-600 dark:text-amber-400 font-medium">学习缓冲区</span>。自动挖掘和反馈学习产生的新词会先进入这里，用于评估效果。确认无误后可一键归档到用户词库。
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
