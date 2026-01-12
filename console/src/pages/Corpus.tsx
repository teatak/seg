import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, Sparkles, Zap } from 'lucide-react';
import { getApiPath } from '@/lib/api';

interface Candidate { word: string; freq: number; score: number }
interface MineResult { added_words?: string[]; mined_candidates?: number; source_texts?: number; candidates?: Candidate[] }

export default function Corpus() {
    const [corpus, setCorpus] = useState('');
    const [loading, setLoading] = useState(false);
    const [result, setResult] = useState<MineResult | null>(null);
    const [autoLoading, setAutoLoading] = useState(false);
    const [autoResult, setAutoResult] = useState<MineResult | null>(null);

    const mineCorpus = async () => {
        if (!corpus.trim()) return;
        setLoading(true);
        setResult(null);
        try {
            const texts = corpus.split('\n').filter(t => t.trim());
            const res = await fetch(getApiPath('learn'), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: 'corpus', texts })
            });
            setResult(await res.json());
        } catch (e) { console.error(e); }
        finally { setLoading(false); }
    };

    const triggerAutoLearn = async () => {
        setAutoLoading(true);
        setAutoResult(null);
        try {
            const res = await fetch(getApiPath('learn/requests'), { method: 'POST' });
            setAutoResult(await res.json());
        } catch (e) { console.error(e); }
        finally { setAutoLoading(false); }
    };

    const ResultDisplay = ({ data, title }: { data: MineResult | null; title: string }) => {
        if (!data) return null;
        return (
            <div className="mt-4 p-4 bg-muted rounded-lg space-y-3">
                <h4 className="font-medium text-foreground">{title}</h4>
                <div className="text-sm text-muted-foreground">
                    {data.source_texts !== undefined && <p>📊 分析了 <b>{data.source_texts}</b> 条文本</p>}
                    <p>🔍 发现 <b>{data.mined_candidates || 0}</b> 个候选词</p>
                    {data.added_words && data.added_words.length > 0 ? (
                        <p className="text-emerald-600 dark:text-emerald-400">🆕 新增词语: <b>{data.added_words.join(', ')}</b></p>
                    ) : (
                        <p className="text-muted-foreground">💡 未发现符合条件的新词</p>
                    )}
                </div>
                {data.candidates && data.candidates.length > 0 && (
                    <div className="mt-3">
                        <p className="text-xs text-muted-foreground mb-2">候选词详情：</p>
                        <div className="max-h-40 overflow-auto">
                            <table className="w-full text-xs">
                                <thead className="bg-muted">
                                    <tr>
                                        <th className="px-2 py-1 text-left">词语</th>
                                        <th className="px-2 py-1 text-center">频率</th>
                                        <th className="px-2 py-1 text-center">得分</th>
                                        <th className="px-2 py-1 text-center">状态</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {data.candidates.map((c, i) => (
                                        <tr key={i} className="border-t border-border">
                                            <td className="px-2 py-1">{c.word}</td>
                                            <td className="px-2 py-1 text-center">{c.freq}</td>
                                            <td className="px-2 py-1 text-center">{c.score.toFixed(2)}</td>
                                            <td className="px-2 py-1 text-center">
                                                {data.added_words?.includes(c.word) ? (
                                                    <span className="text-emerald-600 dark:text-emerald-400">✅</span>
                                                ) : c.score < 1.0 ? (
                                                    <span className="text-amber-500 dark:text-amber-400">得分低</span>
                                                ) : (
                                                    <span className="text-muted-foreground">-</span>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                )}
            </div>
        );
    };

    return (
        <div className="space-y-6">
            <h1 className="text-2xl font-bold text-foreground">语料学习</h1>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><Sparkles className="w-5 h-5 text-purple-500" /> 语料新词挖掘</CardTitle>
                        <p className="text-sm text-muted-foreground">输入多行文本，自动发现新词</p>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <textarea
                            className="w-full p-3 border border-border bg-card rounded-lg resize-none focus:ring-2 focus:ring-primary outline-none"
                            rows={6}
                            placeholder="每行一段文本..."
                            value={corpus}
                            onChange={(e) => setCorpus(e.target.value)}
                        />
                        <button onClick={mineCorpus} disabled={loading || !corpus.trim()} className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 cursor-pointer">
                            {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4" />} 自动发现新词
                        </button>
                        <ResultDisplay data={result} title="挖掘结果" />
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><Zap className="w-5 h-5 text-amber-500" /> 请求语料学习</CardTitle>
                        <p className="text-sm text-muted-foreground">从历史分词请求中自动挖掘新词</p>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <p className="text-sm text-muted-foreground">系统会自动记录用户的分词请求，积累到一定量后可触发新词挖掘。</p>
                        <button onClick={triggerAutoLearn} disabled={autoLoading} className="flex items-center gap-2 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 cursor-pointer">
                            {autoLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Zap className="w-4 h-4" />} 立即从请求学习
                        </button>
                        <ResultDisplay data={autoResult} title="学习结果" />
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
