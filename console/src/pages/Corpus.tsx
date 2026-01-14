import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea'
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
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
                        <div className="max-h-40 overflow-auto border rounded-md bg-background">
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="w-[100px]">词语</TableHead>
                                        <TableHead className="text-center">频率</TableHead>
                                        <TableHead className="text-center">得分</TableHead>
                                        <TableHead className="text-center">状态</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {data.candidates.map((c, i) => (
                                        <TableRow key={i}>
                                            <TableCell className="font-medium">{c.word}</TableCell>
                                            <TableCell className="text-center">{c.freq}</TableCell>
                                            <TableCell className="text-center">{c.score.toFixed(2)}</TableCell>
                                            <TableCell className="text-center">
                                                {data.added_words?.includes(c.word) ? (
                                                    <span className="text-emerald-600 dark:text-emerald-400">✅</span>
                                                ) : c.score < 1.0 ? (
                                                    <span className="text-amber-500 dark:text-amber-400">得分低</span>
                                                ) : (
                                                    <span className="text-muted-foreground">-</span>
                                                )}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        </div>
                    </div>
                )}
            </div>
        );
    };

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-bold text-foreground">语料学习</h1>
                <p className="text-sm text-muted-foreground mt-1">从语料和请求中自动发现新词</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="flex items-center gap-2 text-base"><Sparkles className="w-4 h-4 text-purple-500" /> 语料新词挖掘</CardTitle>
                    </CardHeader>
                    <CardContent className="p-3 md:p-6 space-y-4">
                        <Textarea
                            rows={6}
                            placeholder="每行一段文本..."
                            value={corpus}
                            onChange={(e) => setCorpus(e.target.value)}
                        />
                        <Button onClick={mineCorpus} disabled={loading || !corpus.trim()}>
                            {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Sparkles className="w-4 h-4 mr-2" />} 自动发现新词
                        </Button>
                        <ResultDisplay data={result} title="挖掘结果" />
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="pb-2">
                        <CardTitle className="flex items-center gap-2 text-base"><Zap className="w-4 h-4 text-amber-500" /> 请求语料学习</CardTitle>
                    </CardHeader>
                    <CardContent className="p-3 md:p-6 space-y-4">
                        <p className="text-sm text-muted-foreground">系统会自动记录用户的分词请求，积累到一定量后可触发新词挖掘。</p>
                        <Button onClick={triggerAutoLearn} disabled={autoLoading}>
                            {autoLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Zap className="w-4 h-4 mr-2" />} 立即从请求学习
                        </Button>
                        <ResultDisplay data={autoResult} title="学习结果" />
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}
