import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, RotateCcw } from 'lucide-react';
import { getApiPath } from '@/lib/api';

interface Token { text: string; type: string }

export default function Segment() {
    const [input, setInput] = useState('如今人工智能已经从一种令人惊叹的新兴技术，演变为像电力一样不可或缺的社会基础资源。');
    const [tokens, setTokens] = useState<Token[]>([]);
    const [originalTokens, setOriginalTokens] = useState<Token[]>([]);
    const [loading, setLoading] = useState(false);
    const [message, setMessage] = useState('');
    const [hoverMerge, setHoverMerge] = useState<number | null>(null);

    const segment = async () => {
        if (!input.trim()) return;
        setLoading(true);
        setMessage('');
        try {
            const res = await fetch(getApiPath('segment?mode=eval'), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ text: input })
            });
            const data = await res.json();
            const parsedTokens = (data.tokens || []).map((t: { word: string; type: string }) => ({
                text: t.word,
                type: t.type === 'word' ? 'dict' : t.type
            }));
            setTokens(parsedTokens);
            setOriginalTokens(parsedTokens);
        } catch (e) { console.error(e); }
        finally { setLoading(false); }
    };

    const mergeAt = (idx: number) => {
        if (idx >= tokens.length - 1) return;
        const merged = tokens[idx].text + tokens[idx + 1].text;
        const newTokens = [
            ...tokens.slice(0, idx),
            { text: merged, type: 'user' },
            ...tokens.slice(idx + 2)
        ];
        setTokens(newTokens);
        setHoverMerge(null);
    };

    // 按位置拆分词语
    const splitAt = (idx: number, charIndex: number) => {
        const text = tokens[idx].text;
        if (text.length <= 1 || charIndex <= 0 || charIndex >= text.length) return;
        const left = text.slice(0, charIndex);
        const right = text.slice(charIndex);
        const newTokens = [
            ...tokens.slice(0, idx),
            { text: left, type: left.length === 1 ? 'single' : 'user' },
            { text: right, type: right.length === 1 ? 'single' : 'user' },
            ...tokens.slice(idx + 1)
        ];
        setTokens(newTokens);
    };

    // 处理双击拆分
    const handleDoubleClick = (idx: number, e: React.MouseEvent<HTMLSpanElement>) => {
        const token = tokens[idx];
        if (token.type === 'alphanum' || token.type === 'punct' || token.text.length <= 1) return;

        const span = e.currentTarget;
        const rect = span.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const charWidth = rect.width / token.text.length;
        const charIndex = Math.round(x / charWidth);

        if (charIndex > 0 && charIndex < token.text.length) {
            splitAt(idx, charIndex);
        }
    };

    const reset = () => { setTokens([...originalTokens]); setMessage(''); setHoverMerge(null); };

    const submit = async () => {
        setLoading(true);
        try {
            await fetch(getApiPath('feedback'), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    text: input,
                    original_seg: originalTokens.map(t => t.text),
                    corrected_seg: tokens.map(t => t.text)
                })
            });
            await fetch(getApiPath('learn'), {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ type: 'feedback' })
            });
            setMessage('✅ 反馈已提交！');
        } catch { setMessage('❌ 提交失败'); }
        finally { setLoading(false); }
    };

    const getTokenClass = (type: string, idx: number) => {
        const base = "px-3 py-1.5 text-base font-medium select-none ";
        const baseSmall = "text-sm select-none flex items-center "; // 只读词用小字体，垂直居中
        const isHighlight = hoverMerge === idx || hoverMerge === idx - 1;
        if (isHighlight) return base + "bg-emerald-200 dark:bg-emerald-800 text-emerald-900 dark:text-emerald-100";
        switch (type) {
            case 'dict':
            case 'word': return base + "bg-sky-100 dark:bg-sky-900/50 text-sky-800 dark:text-sky-200";
            case 'user': return base + "bg-emerald-100 dark:bg-emerald-900/50 text-emerald-800 dark:text-emerald-200";
            case 'hmm': return base + "bg-amber-100 dark:bg-amber-900/50 text-amber-800 dark:text-amber-200";
            case 'single': return base + "bg-violet-100 dark:bg-violet-900/50 text-violet-800 dark:text-violet-200";
            case 'alphanum': return baseSmall + "text-muted-foreground";
            case 'punct': return baseSmall + "text-muted-foreground";
            default: return base + "bg-muted text-foreground";
        }
    };

    const getRoundedClass = (idx: number) => {
        const isLeft = hoverMerge === idx;
        const isRight = hoverMerge === idx - 1;
        if (isLeft) return "rounded-l rounded-r-none";
        if (isRight) return "rounded-r rounded-l-none";
        return "rounded";
    };

    return (
        <div className="space-y-6">
            <h1 className="text-2xl font-bold text-foreground">分词测试</h1>

            <Card>
                <CardHeader>
                    <CardTitle>交互式分词编辑</CardTitle>
                    <p className="text-sm text-muted-foreground">词语之间悬浮可合并，双击词语可在点击位置拆分</p>
                </CardHeader>
                <CardContent className="space-y-4">
                    <textarea
                        className="w-full p-3 border border-border dark:bg-card rounded-lg resize-none focus:ring-2 focus:ring-primary outline-none"
                        rows={3}
                        placeholder="请输入要分词的中文文本..."
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                    />
                    <div className="flex gap-2">
                        <button onClick={segment} disabled={loading} className="px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 cursor-pointer">
                            {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : '🔍 分词'}
                        </button>
                        {tokens.length > 0 && (
                            <button onClick={reset} className="flex items-center gap-1 px-4 py-2 text-sm font-medium text-secondary-foreground bg-secondary rounded-lg hover:bg-secondary/80 cursor-pointer">
                                <RotateCcw className="w-4 h-4" /> 重置
                            </button>
                        )}
                    </div>

                    {tokens.length > 0 && (
                        <>
                            <div className="flex flex-wrap items-stretch gap-y-2 p-4 bg-muted dark:bg-card rounded-lg min-h-[60px]">
                                {tokens.map((token, idx) => (
                                    <div key={idx} className="flex items-stretch">
                                        {/* 左侧间隔：只读词左边有非只读词时 */}
                                        {idx > 0 &&
                                            (token.type === 'alphanum' || token.type === 'punct') &&
                                            tokens[idx - 1].type !== 'alphanum' && tokens[idx - 1].type !== 'punct' && (
                                                <div className="w-1" />
                                            )}
                                        {/* 词语 */}
                                        <span
                                            className={`${getTokenClass(token.type, idx)} ${getRoundedClass(idx)} ${token.type !== 'alphanum' && token.type !== 'punct' && token.text.length > 1 ? 'cursor-pointer' : ''}`}
                                            onDoubleClick={(e) => handleDoubleClick(idx, e)}
                                        >
                                            {token.text}
                                        </span>
                                        {/* 右侧间隔 */}
                                        {idx < tokens.length - 1 && (() => {
                                            const isReadonly = token.type === 'alphanum' || token.type === 'punct';
                                            const nextIsReadonly = tokens[idx + 1].type === 'alphanum' || tokens[idx + 1].type === 'punct';
                                            // 两边都是可编辑词：显示可交互合并热区
                                            if (!isReadonly && !nextIsReadonly) {
                                                return (
                                                    <div
                                                        onClick={() => mergeAt(idx)}
                                                        onMouseEnter={() => setHoverMerge(idx)}
                                                        onMouseLeave={() => setHoverMerge(null)}
                                                        className={`w-2 cursor-pointer transition-colors ${hoverMerge === idx ? 'bg-emerald-200 dark:bg-emerald-800' : ''}`}
                                                    />
                                                );
                                            }
                                            // 只读词右边有非只读词：加间隔
                                            if (isReadonly && !nextIsReadonly) {
                                                return <div className="w-1" />;
                                            }
                                            // 只读词之间：小间隔
                                            if (isReadonly && nextIsReadonly) {
                                                return <div className="w-0.5" />;
                                            }
                                            // 其他情况
                                            return null;
                                        })()}
                                    </div>
                                ))}
                            </div>

                            <div className="flex gap-2">
                                <button onClick={submit} disabled={loading} className="px-3 py-1.5 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:opacity-50 cursor-pointer">
                                    ✅ 提交反馈
                                </button>
                            </div>

                            <div className="flex flex-wrap gap-3 text-xs text-muted-foreground">
                                <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-sky-200 dark:bg-sky-800"></span>词典词</span>
                                <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-amber-200 dark:bg-amber-800"></span>HMM推断</span>
                                <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-emerald-200 dark:bg-emerald-800"></span>用户词</span>
                                <span className="flex items-center gap-1"><span className="w-3 h-3 rounded bg-violet-200 dark:bg-violet-800"></span>单字</span>

                            </div>

                            <div className="text-xs text-muted-foreground bg-muted dark:bg-card rounded px-3 py-2">
                                💡 <b>操作说明：</b>悬浮词语间隙点击可<b>合并</b>，双击词语可在点击位置<b>拆分</b>
                            </div>

                            {message && <div className={`p-3 rounded-lg text-sm ${message.startsWith('✅') ? 'bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-300' : 'bg-destructive/10 text-destructive'}`}>{message}</div>}
                        </>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
