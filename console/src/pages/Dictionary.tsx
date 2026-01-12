import { useEffect, useState } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { getApiPath } from '@/lib/api';

import { Search, Loader2, Trash2, Plus, Check, X, ChevronLeft, ChevronRight, SquarePen, GitMerge } from 'lucide-react';

interface WordItem { word: string; freq: number; type: 'user' | 'base' | 'staging' }

export default function Dictionary() {
    const [words, setWords] = useState<WordItem[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);
    const [keyword, setKeyword] = useState('');
    const [selected, setSelected] = useState<Set<string>>(new Set());
    const [editingWord, setEditingWord] = useState<string | null>(null);
    const [editFreq, setEditFreq] = useState('');
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(20);
    const [showAddModal, setShowAddModal] = useState(false);
    const [newWord, setNewWord] = useState('');
    const [newFreq, setNewFreq] = useState('100');
    const [filterType, setFilterType] = useState<'all' | 'user' | 'base' | 'staging'>('user');

    const selectableWords = words.filter(w => w.type === 'user' || w.type === 'staging');
    const totalPages = Math.ceil(total / pageSize);

    const fetchWords = async (q = '', p = page, size = pageSize, type = filterType) => {
        setLoading(true);
        try {
            const url = q.trim()
                ? getApiPath(`words/search?q=${encodeURIComponent(q.trim())}`)
                : getApiPath(`words/list?page=${p}&size=${size}&type=${type}`);
            const res = await fetch(url);
            const data = await res.json();
            if (q.trim()) { // Changed condition to check if a search query was made
                setWords(data);
                setTotal(data.length);
            } else {
                setWords(data.items || []);
                setTotal(data.total || 0);
            }
            setSelected(new Set());
        } catch (error) {
            console.error("Failed to fetch words", error);
            setWords([]); setTotal(0); // Keep original behavior for error state
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchWords(keyword, page, pageSize, filterType); }, [page, pageSize, filterType]);

    const handleSearch = (e: React.FormEvent) => { e.preventDefault(); setPage(1); fetchWords(keyword, 1, pageSize, filterType); };

    const handleMerge = async () => {
        if (!confirm('确定要将所有暂存词合并到用户词库吗？这将立即生效到生产环境。')) return;
        try {
            const res = await fetch(getApiPath('dict/merge'), { method: 'POST' });
            if (res.ok) {
                alert('合并成功！');
                fetchWords(keyword, page, pageSize, filterType); // Keep original call with parameters
            } else {
                alert('合并失败');
            }
        } catch (e) { console.error(e); alert('合并请求出错'); }
    };

    const handleAdd = async () => {
        if (!newWord.trim()) return;
        try {
            await fetch(getApiPath('words'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ word: newWord.trim(), freq: parseInt(newFreq) || 100 }) });
            setNewWord(''); setNewFreq('100'); setShowAddModal(false);
            fetchWords(keyword, page, pageSize, filterType); // Keep original call with parameters
        } catch (e) { console.error(e); }
    };

    const handleDelete = async (word: string) => {
        if (!confirm(`确定删除 "${word}" 吗？`)) return; // Added confirmation as per instruction's spirit
        try {
            await fetch(getApiPath(`words/${encodeURIComponent(word)}`), { method: 'DELETE' });
            setWords(prev => prev.filter(w => w.word !== word)); // Keep original immediate UI update
            setSelected(prev => { prev.delete(word); return new Set(prev); }); // Keep original immediate UI update
            setTotal(prev => prev - 1); // Keep original immediate UI update
            // fetchWords(keyword, page, pageSize, filterType); // Re-fetching might be redundant if UI is updated immediately
        } catch (e) { console.error(e); }
    };

    const handleBatchDelete = async () => {
        if (!confirm(`确定删除选中的 ${selected.size} 个词吗？`)) return; // Added confirmation as per instruction's spirit
        try {
            for (const word of selected) await fetch(getApiPath(`words/${encodeURIComponent(word)}`), { method: 'DELETE' });
            const count = selected.size; // Keep original count for UI update
            setWords(prev => prev.filter(w => !selected.has(w.word))); // Keep original immediate UI update
            setSelected(new Set());
            setTotal(prev => prev - count); // Keep original immediate UI update
            // fetchWords(keyword, page, pageSize, filterType); // Re-fetching might be redundant if UI is updated immediately
        } catch (e) { console.error(e); }
    };

    const toggleSelect = (word: string) => setSelected(prev => { const next = new Set(prev); next.has(word) ? next.delete(word) : next.add(word); return next; });
    const selectAll = () => setSelected(new Set(selectableWords.map(w => w.word)));
    const selectNone = () => setSelected(new Set());
    const selectInverse = () => { const s = new Set(selectableWords.map(w => w.word)); setSelected(prev => { const next = new Set<string>(); s.forEach(w => { if (!prev.has(w)) next.add(w); }); return next; }); };

    const startEdit = (word: string, freq: number) => { setEditingWord(word); setEditFreq(String(freq)); };
    const cancelEdit = () => { setEditingWord(null); setEditFreq(''); };
    const saveEdit = async () => {
        if (!editingWord) return;
        const freq = parseInt(editFreq) || 100;
        await fetch(`/api/words/${encodeURIComponent(editingWord)}`, { method: 'DELETE' });
        await fetch('/api/words', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ word: editingWord, freq }) });
        setWords(prev => prev.map(w => w.word === editingWord ? { ...w, freq } : w));
        cancelEdit();
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h1 className="text-2xl font-bold text-foreground">词库管理</h1>
            </div>

            <Card>
                <CardContent className="pt-4 space-y-3">
                    {/* 顶部操作区 */}
                    <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                        {/* 筛选 Tabs */}
                        <div className="flex bg-muted p-1 rounded-lg self-start">
                            {[
                                { id: 'all', label: '全部' },
                                { id: 'user', label: '用户词库' },
                                { id: 'staging', label: '暂存区' },
                                { id: 'base', label: '基础词库' }
                            ].map(tab => (
                                <button
                                    key={tab.id}
                                    onClick={() => { setFilterType(tab.id as any); setPage(1); }}
                                    className={`px-3 py-1.5 text-sm font-medium rounded-md transition-all ${filterType === tab.id
                                        ? 'bg-card text-foreground shadow-sm'
                                        : 'text-muted-foreground hover:text-foreground'
                                        }`}
                                >
                                    {tab.label}
                                </button>
                            ))}
                        </div>

                        {/* 特殊操作按钮 */}
                        <div className="flex gap-2">
                            {filterType === 'staging' && (
                                <button onClick={handleMerge} className="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-white bg-amber-500 hover:bg-amber-600 rounded-lg cursor-pointer">
                                    <GitMerge className="w-4 h-4" /> 归档到生产
                                </button>
                            )}
                            <button onClick={() => setShowAddModal(true)} className="flex items-center gap-1 px-3 py-1.5 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 cursor-pointer">
                                <Plus className="w-4 h-4" /> 添加
                            </button>
                        </div>
                    </div>

                    {/* 搜索栏 */}
                    <form onSubmit={handleSearch} className="flex gap-2">
                        <div className="relative flex-1">
                            <Search className="absolute left-2.5 top-2 h-4 w-4 text-muted-foreground" />
                            <input type="text" placeholder="搜索词语..." className="w-full pl-8 pr-3 py-1.5 text-sm border border-border rounded-lg outline-none focus:ring-2 focus:ring-primary"
                                value={keyword} onChange={(e) => setKeyword(e.target.value)} />
                        </div>
                        <button type="submit" className="px-3 py-1.5 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 cursor-pointer">搜索</button>
                    </form>

                    {/* 批量操作工具栏 */}
                    <div className="flex items-center gap-2 text-xs">
                        <button onClick={selectAll} className="px-2 py-0.5 text-primary hover:bg-primary/10 dark:hover:bg-primary/20 rounded cursor-pointer">全选</button>
                        <button onClick={selectNone} className="px-2 py-0.5 text-muted-foreground hover:bg-muted rounded cursor-pointer">取消</button>
                        <button onClick={selectInverse} className="px-2 py-0.5 text-muted-foreground hover:bg-muted rounded cursor-pointer">反选</button>
                        {selected.size > 0 && (
                            <>
                                <span className="text-muted-foreground">|</span>
                                <span className="text-muted-foreground">已选 <b>{selected.size}</b></span>
                                <button onClick={handleBatchDelete} className="flex items-center gap-1 px-2 py-0.5 text-destructive hover:bg-destructive/10 rounded cursor-pointer">
                                    <Trash2 className="w-3 h-3" /> 删除
                                </button>
                            </>
                        )}
                    </div>

                    {/* 紧凑表格 */}
                    <div className="border border-border rounded-lg overflow-hidden">
                        <table className="w-full text-xs">
                            <thead className="bg-muted">
                                <tr>
                                    <th className="px-2 py-2 text-left font-medium text-muted-foreground w-8">
                                        <input type="checkbox" checked={selectableWords.length > 0 && selected.size === selectableWords.length} onChange={(e) => e.target.checked ? selectAll() : selectNone()} className="rounded cursor-pointer" />
                                    </th>
                                    <th className="px-2 py-2 text-left font-medium text-muted-foreground">词语</th>
                                    <th className="px-2 py-2 text-left font-medium text-muted-foreground w-20">词频</th>
                                    <th className="px-2 py-2 text-center font-medium text-muted-foreground w-16">类型</th>
                                    <th className="px-2 py-2 text-center font-medium text-muted-foreground w-16">操作</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-border bg-card">
                                {loading ? (
                                    <tr><td colSpan={5} className="px-2 py-6 text-center text-muted-foreground"><Loader2 className="h-4 w-4 animate-spin mx-auto" /></td></tr>
                                ) : words.length === 0 ? (
                                    <tr><td colSpan={5} className="px-2 py-6 text-center text-muted-foreground">没有找到词语</td></tr>
                                ) : words.map((item, idx) => (
                                    <tr key={idx} className={`hover:bg-muted ${selected.has(item.word) ? 'bg-primary/10 dark:bg-primary/20' : ''}`}>
                                        <td className="px-2 py-1.5">{(item.type === 'user' || item.type === 'staging') && <input type="checkbox" checked={selected.has(item.word)} onChange={() => toggleSelect(item.word)} className="rounded cursor-pointer" />}</td>
                                        <td className="px-2 py-1.5 font-medium">{item.word}</td>
                                        <td className="px-2 py-1.5 text-muted-foreground">
                                            {editingWord === item.word ? (
                                                <input type="number" value={editFreq} onChange={(e) => setEditFreq(e.target.value)} className="w-16 px-1 py-0.5 text-xs text-foreground border border-primary rounded bg-card focus:ring-1 focus:ring-primary outline-none" autoFocus />
                                            ) : (
                                                <span>{item.freq}</span>
                                            )}
                                        </td>
                                        <td className="px-2 py-1.5 text-center">
                                            <span className={`inline-flex px-1.5 py-0.5 text-xs rounded ${item.type === 'user'
                                                ? 'bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300'
                                                : item.type === 'staging'
                                                    ? 'bg-amber-100 dark:bg-amber-900/50 text-amber-700 dark:text-amber-300'
                                                    : 'bg-sky-100 dark:bg-sky-900/50 text-sky-700 dark:text-sky-300'
                                                }`}>
                                                {item.type === 'user' ? '用户' : item.type === 'staging' ? '暂存' : '基础'}
                                            </span>
                                        </td>
                                        <td className="px-2 py-1.5">
                                            {(item.type === 'user' || item.type === 'staging') && (
                                                editingWord === item.word ? (
                                                    <div className="flex gap-0.5 justify-around">
                                                        <button onClick={saveEdit} className="p-0.5 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/30 rounded cursor-pointer"><Check className="w-3 h-3" /></button>
                                                        <button onClick={cancelEdit} className="p-0.5 text-muted-foreground hover:bg-muted rounded cursor-pointer"><X className="w-3 h-3" /></button>
                                                    </div>
                                                ) : (
                                                    <div className="flex gap-0.5 justify-around">
                                                        <button onClick={() => startEdit(item.word, item.freq)} className="p-0.5 text-muted-foreground hover:text-primary hover:bg-primary/10 dark:hover:bg-primary/20 rounded cursor-pointer"><SquarePen className="w-3 h-3" /></button>
                                                        <button onClick={() => handleDelete(item.word)} className="p-0.5 text-destructive hover:bg-destructive/10 rounded cursor-pointer"><Trash2 className="w-3 h-3" /></button>
                                                    </div>
                                                )
                                            )}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>

                    {/* 分页 */}
                    <div className="flex items-center justify-between text-xs">
                        <div className="flex items-center gap-1 text-muted-foreground">
                            <span>共 {total} 条</span>
                            <select value={pageSize} onChange={(e) => { setPageSize(Number(e.target.value)); setPage(1); }} className="px-1 py-0.5 border border-border rounded cursor-pointer text-xs">
                                <option value={10}>10</option><option value={20}>20</option><option value={50}>50</option><option value={100}>100</option>
                            </select>
                            <span>条/页</span>
                        </div>
                        <div className="flex items-center gap-0.5">
                            <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page <= 1} className="p-1 rounded hover:bg-muted disabled:opacity-30 cursor-pointer"><ChevronLeft className="w-4 h-4" /></button>
                            <span className="px-2 text-muted-foreground">{page}/{totalPages || 1}</span>
                            <button onClick={() => setPage(p => Math.min(totalPages, p + 1))} disabled={page >= totalPages} className="p-1 rounded hover:bg-muted disabled:opacity-30 cursor-pointer"><ChevronRight className="w-4 h-4" /></button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 添加词语弹窗 */}
            {
                showAddModal && (
                    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setShowAddModal(false)}>
                        <div className="bg-card rounded-xl shadow-xl p-6 w-80 space-y-4" onClick={e => e.stopPropagation()}>
                            <h3 className="text-lg font-semibold text-foreground">添加新词</h3>
                            <div className="space-y-3">
                                <div>
                                    <label className="block text-sm font-medium text-foreground mb-1">词语</label>
                                    <input type="text" placeholder="输入词语" className="w-full px-3 py-2 text-sm border border-border rounded-lg outline-none focus:ring-2 focus:ring-primary"
                                        value={newWord} onChange={(e) => setNewWord(e.target.value)} autoFocus />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-foreground mb-1">词频</label>
                                    <input type="number" placeholder="100" className="w-full px-3 py-2 text-sm border border-border rounded-lg outline-none focus:ring-2 focus:ring-primary"
                                        value={newFreq} onChange={(e) => setNewFreq(e.target.value)} />
                                </div>
                            </div>
                            <div className="flex gap-2 justify-end">
                                <button onClick={() => setShowAddModal(false)} className="px-4 py-2 text-sm font-medium text-foreground bg-muted rounded-lg hover:bg-muted cursor-pointer">取消</button>
                                <button onClick={handleAdd} className="px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 cursor-pointer">添加</button>
                            </div>
                        </div>
                    </div>
                )
            }
        </div >
    );
}
