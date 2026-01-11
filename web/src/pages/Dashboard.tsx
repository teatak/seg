import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { RefreshCw, Database, BookOpen, MessageSquare, Activity } from 'lucide-react';

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
            const res = await fetch('/api/stats');
            setStats(await res.json());
            await new Promise(r => setTimeout(r, 300)); // 延迟让 loading 效果更明显
        } catch (e) { console.error(e); }
        finally { setLoading(false); }
    };

    useEffect(() => { fetchStats(); }, []);

    const StatCard = ({ title, value, icon: Icon, color }: { title: string; value: number; icon: React.ElementType; color: string }) => (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">{title}</CardTitle>
                <Icon className={`h-4 w-4 ${color}`} />
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">{value.toLocaleString()}</div>
            </CardContent>
        </Card>
    );

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-foreground">首页</h1>
                <button onClick={fetchStats} disabled={loading} className="flex items-center gap-2 px-3 py-2 text-sm bg-card border border-border rounded-lg hover:bg-muted cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed">
                    <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} /> 刷新
                </button>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
                <StatCard title="基础词库" value={stats?.dictionary.base_words || 0} icon={Database} color="text-primary" />
                <StatCard title="用户词库" value={stats?.dictionary.user_words || 0} icon={BookOpen} color="text-emerald-500 dark:text-emerald-400" />
                <StatCard title="暂存词库" value={stats?.dictionary.staging_words || 0} icon={Database} color="text-amber-500 dark:text-amber-400" />
                <StatCard title="用户反馈" value={stats?.feedback.total || 0} icon={MessageSquare} color="text-blue-500 dark:text-blue-400" />
                <StatCard title="请求记录" value={stats?.requests?.total || 0} icon={Activity} color="text-purple-500 dark:text-purple-400" />
            </div>

            <Card>
                <CardHeader><CardTitle>自动学习状态</CardTitle></CardHeader>
                <CardContent>
                    <p className="text-sm text-muted-foreground">
                        距下次自动挖掘还需积累 <span className="font-bold text-foreground">{Math.max(0, (stats?.requests?.mine_threshold || 100) - (stats?.requests?.since_last || 0))}</span> 条请求
                    </p>
                </CardContent>
            </Card>

            <div className="grid gap-4 md:grid-cols-3">
                <Card>
                    <CardHeader className="pb-3"><CardTitle className="text-base">📚 基础词库 (Base)</CardTitle></CardHeader>
                    <CardContent className="text-sm text-muted-foreground">
                        系统内置的通用核心词典，包含几十万标准中文词汇。它是分词的基础，只读不可修改。
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="pb-3"><CardTitle className="text-base">🌲 用户词库 (User)</CardTitle></CardHeader>
                    <CardContent className="text-sm text-muted-foreground">
                        <span className="text-emerald-600 dark:text-emerald-400 font-medium">最高优先级</span>。包含您人工添加或从暂存区归档的词语。用于覆盖基础分词结果，修正特定领域的术语。
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="pb-3"><CardTitle className="text-base">🧪 暂存词库 (Staging)</CardTitle></CardHeader>
                    <CardContent className="text-sm text-muted-foreground">
                        <span className="text-amber-600 dark:text-amber-400 font-medium">学习缓冲区</span>。自动挖掘和反馈学习产生的新词会先进入这里，用于评估效果。确认无误后可一键归档到用户词库。
                    </CardContent>
                </Card>
            </div>
        </div >
    );
}
