import { useEffect, useState, useRef } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';

import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
    Pagination,
    PaginationContent,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
    PaginationEllipsis,
} from '@/components/ui/pagination';
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { getApiPath } from '@/lib/api';

import { Search, Loader2, Trash2, Plus, Check, X, SquarePen, GitMerge } from 'lucide-react';

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

    // 删除确认对话框状态
    const [deleteTarget, setDeleteTarget] = useState<{ type: 'single' | 'batch'; word?: string } | null>(null);
    const [mergeConfirmOpen, setMergeConfirmOpen] = useState(false);

    const selectableWords = words.filter(w => w.type === 'user' || w.type === 'staging');
    const totalPages = Math.ceil(total / pageSize);

    // 计算 checkbox 状态
    const getCheckboxState = () => {
        if (selectableWords.length === 0) return 'none';
        if (selected.size === 0) return 'none';
        if (selected.size === selectableWords.length) return 'all';
        return 'indeterminate';
    };

    const checkboxRef = useRef<HTMLButtonElement>(null);

    useEffect(() => {
        if (checkboxRef.current) {
            const state = getCheckboxState();
            checkboxRef.current.dataset.state = state === 'indeterminate' ? 'indeterminate' : state === 'all' ? 'checked' : 'unchecked';
        }
    }, [selected, selectableWords]);

    const fetchWords = async (q = '', p = page, size = pageSize, type = filterType) => {
        setLoading(true);
        try {
            const url = q.trim()
                ? getApiPath(`words/search?q=${encodeURIComponent(q.trim())}`)
                : getApiPath(`words/list?page=${p}&size=${size}&type=${type}`);
            const res = await fetch(url);
            const data = await res.json();
            if (q.trim()) {
                setWords(data);
                setTotal(data.length);
            } else {
                setWords(data.items || []);
                setTotal(data.total || 0);
            }
            setSelected(new Set());
        } catch (error) {
            console.error("Failed to fetch words", error);
            setWords([]); setTotal(0);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchWords(keyword, page, pageSize, filterType); }, [page, pageSize, filterType]);

    const handleSearch = (e: React.FormEvent) => { e.preventDefault(); setPage(1); fetchWords(keyword, 1, pageSize, filterType); };

    const handleMerge = async () => {
        setMergeConfirmOpen(false);
        try {
            const res = await fetch(getApiPath('dict/merge'), { method: 'POST' });
            if (res.ok) {
                fetchWords(keyword, page, pageSize, filterType);
            }
        } catch (e) { console.error(e); }
    };

    const handleAdd = async () => {
        if (!newWord.trim()) return;
        try {
            await fetch(getApiPath('words'), { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ word: newWord.trim(), freq: parseInt(newFreq) || 100 }) });
            setNewWord(''); setNewFreq('100'); setShowAddModal(false);
            fetchWords(keyword, page, pageSize, filterType);
        } catch (e) { console.error(e); }
    };

    const handleDelete = async (word: string) => {
        try {
            await fetch(getApiPath(`words/${encodeURIComponent(word)}`), { method: 'DELETE' });
            setWords(prev => prev.filter(w => w.word !== word));
            setSelected(prev => { prev.delete(word); return new Set(prev); });
            setTotal(prev => prev - 1);
        } catch (e) { console.error(e); }
        setDeleteTarget(null);
    };

    const handleBatchDelete = async () => {
        try {
            for (const word of selected) await fetch(getApiPath(`words/${encodeURIComponent(word)}`), { method: 'DELETE' });
            const count = selected.size;
            setWords(prev => prev.filter(w => !selected.has(w.word)));
            setSelected(new Set());
            setTotal(prev => prev - count);
        } catch (e) { console.error(e); }
        setDeleteTarget(null);
    };

    const toggleSelect = (word: string) => setSelected(prev => { const next = new Set(prev); next.has(word) ? next.delete(word) : next.add(word); return next; });
    const selectAll = () => setSelected(new Set(selectableWords.map(w => w.word)));
    const selectNone = () => setSelected(new Set());

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
            <div>
                <h1 className="text-2xl font-bold text-foreground">词库管理</h1>
                <p className="text-sm text-muted-foreground mt-1">管理用户词库、暂存词库和基础词库</p>
            </div>

            <Card>
                <CardContent className="pt-4 space-y-3">
                    {/* 顶部操作区 */}
                    <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                        {/* 筛选 Tabs */}
                        <div className="flex bg-muted p-1 rounded-xl self-start">
                            {[
                                { id: 'all', label: '全部' },
                                { id: 'user', label: '用户词库' },
                                { id: 'staging', label: '暂存区' },
                                { id: 'base', label: '基础词库' }
                            ].map(tab => (
                                <button
                                    key={tab.id}
                                    onClick={() => { setFilterType(tab.id as any); setPage(1); }}
                                    className={`px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${filterType === tab.id
                                        ? 'bg-background text-foreground shadow-sm'
                                        : 'text-muted-foreground hover:text-foreground hover:bg-background/50'
                                        }`}
                                >
                                    {tab.label}
                                </button>
                            ))}
                        </div>

                        {/* 特殊操作按钮 */}
                        <div className="flex gap-2">
                            {filterType === 'staging' && (
                                <Button onClick={() => setMergeConfirmOpen(true)} className="bg-amber-500 hover:bg-amber-600 text-white">
                                    <GitMerge className="w-4 h-4 mr-1" /> 归档到生产
                                </Button>
                            )}
                        </div>
                    </div>

                    {/* 搜索栏 */}
                    <form onSubmit={handleSearch} className="flex gap-2">
                        <div className="relative flex-1">
                            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="搜索词语..."
                                className="pl-8"
                                value={keyword}
                                onChange={(e) => setKeyword(e.target.value)}
                            />
                        </div>
                        <Button type="submit">搜索</Button>
                        <Button variant="outline" onClick={() => setShowAddModal(true)}>
                            <Plus className="w-4 h-4 mr-1" /> 添加
                        </Button>
                    </form>

                    {/* 批量操作工具栏 - checkbox 与表格内对齐 */}
                    <div className="flex items-center text-xs h-7 px-2 gap-2">
                        <span className="text-muted-foreground">已选 <b>{selected.size}</b></span>
                        {selected.size > 0 && (
                            <Button variant="ghost" size="sm" onClick={() => setDeleteTarget({ type: 'batch' })} className="h-7 text-destructive hover:text-destructive hover:bg-destructive/10">
                                <Trash2 className="w-3 h-3 mr-1" /> 删除
                            </Button>
                        )}
                    </div>

                    {/* 紧凑表格 */}
                    <div className="border border-border rounded-lg overflow-hidden">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="w-[40px] px-2 text-center bg-muted">
                                        <Checkbox
                                            ref={checkboxRef}
                                            checked={getCheckboxState() === 'all'}
                                            // @ts-ignore
                                            indeterminate={getCheckboxState() === 'indeterminate'}
                                            onCheckedChange={(checked) => checked ? selectAll() : selectNone()}
                                            disabled={selectableWords.length === 0}
                                        />
                                    </TableHead>
                                    <TableHead className="bg-muted">词语</TableHead>
                                    <TableHead className="w-[100px] bg-muted">词频</TableHead>
                                    <TableHead className="w-[80px] text-center bg-muted">类型</TableHead>
                                    <TableHead className="w-[100px] text-center bg-muted">操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {loading ? (
                                    <TableRow>
                                        <TableCell colSpan={5} className="h-24 text-center">
                                            <Loader2 className="h-4 w-4 animate-spin mx-auto" />
                                        </TableCell>
                                    </TableRow>
                                ) : words.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={5} className="h-24 text-center">
                                            没有找到词语
                                        </TableCell>
                                    </TableRow>
                                ) : words.map((item, idx) => (
                                    <TableRow key={idx} className={selected.has(item.word) ? 'bg-primary/10 dark:bg-primary/20' : ''}>
                                        <TableCell className="px-2 py-2 text-center">
                                            {(item.type === 'user' || item.type === 'staging') &&
                                                <Checkbox
                                                    checked={selected.has(item.word)}
                                                    onCheckedChange={() => toggleSelect(item.word)}
                                                />
                                            }
                                        </TableCell>
                                        <TableCell className="font-medium py-2">{item.word}</TableCell>
                                        <TableCell className="py-2">
                                            {editingWord === item.word ? (
                                                <Input
                                                    type="number"
                                                    value={editFreq}
                                                    onChange={(e) => setEditFreq(e.target.value)}
                                                    className="w-20 h-8 text-xs"
                                                    autoFocus
                                                />
                                            ) : (
                                                <span className="text-muted-foreground">{item.freq}</span>
                                            )}
                                        </TableCell>
                                        <TableCell className="text-center py-2">
                                            <span className={`inline-flex px-1.5 py-0.5 text-xs rounded border ${item.type === 'user'
                                                ? 'bg-emerald-100 dark:bg-emerald-900/50 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800'
                                                : item.type === 'staging'
                                                    ? 'bg-amber-100 dark:bg-amber-900/50 text-amber-700 dark:text-amber-300 border-amber-200 dark:border-amber-800'
                                                    : 'bg-sky-100 dark:bg-sky-900/50 text-sky-700 dark:text-sky-300 border-sky-200 dark:border-sky-800'
                                                }`}>
                                                {item.type === 'user' ? '用户' : item.type === 'staging' ? '暂存' : '基础'}
                                            </span>
                                        </TableCell>
                                        <TableCell className="text-center py-2">
                                            {(item.type === 'user' || item.type === 'staging') && (
                                                editingWord === item.word ? (
                                                    <div className="flex gap-1 justify-center">
                                                        <Button variant="ghost" size="icon" className="h-7 w-7 text-emerald-600 hover:text-emerald-700 hover:bg-emerald-100" onClick={saveEdit}><Check className="w-3.5 h-3.5" /></Button>
                                                        <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:bg-muted" onClick={cancelEdit}><X className="w-3.5 h-3.5" /></Button>
                                                    </div>
                                                ) : (
                                                    <div className="flex gap-1 justify-center">
                                                        <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-primary" onClick={() => startEdit(item.word, item.freq)}><SquarePen className="w-3.5 h-3.5" /></Button>
                                                        <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive hover:text-destructive hover:bg-destructive/10" onClick={() => setDeleteTarget({ type: 'single', word: item.word })}><Trash2 className="w-3.5 h-3.5" /></Button>
                                                    </div>
                                                )
                                            )}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>

                    {/* 分页 */}
                    <div className="flex items-center justify-end text-xs pt-2">
                        <Pagination className="w-auto mx-0">
                            <PaginationContent>
                                <PaginationItem>
                                    <span className="text-muted-foreground mr-2">共 {total} 条</span>
                                </PaginationItem>
                                <PaginationItem>
                                    <Select value={String(pageSize)} onValueChange={(value) => { setPageSize(Number(value)); setPage(1); }}>
                                        <SelectTrigger size="sm" className="h-7 text-xs px-2 min-w-[60px]">
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="10">10</SelectItem>
                                            <SelectItem value="20">20</SelectItem>
                                            <SelectItem value="50">50</SelectItem>
                                            <SelectItem value="100">100</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </PaginationItem>
                                <PaginationItem>
                                    <span className="text-muted-foreground mr-2">条/页</span>
                                </PaginationItem>
                                <PaginationItem>
                                    <PaginationPrevious
                                        onClick={(e) => { e.preventDefault(); if (page > 1) setPage(p => p - 1); }}
                                        className={page <= 1 ? "pointer-events-none opacity-50" : "cursor-pointer"}
                                        aria-disabled={page <= 1}
                                    />
                                </PaginationItem>
                                {/* 第一页 */}
                                {page > 2 && (
                                    <PaginationItem>
                                        <PaginationLink onClick={(e) => { e.preventDefault(); setPage(1); }} className="cursor-pointer">
                                            1
                                        </PaginationLink>
                                    </PaginationItem>
                                )}
                                {/* 左省略 */}
                                {page > 3 && (
                                    <PaginationItem>
                                        <PaginationEllipsis />
                                    </PaginationItem>
                                )}
                                {/* 上一页 */}
                                {page > 1 && (
                                    <PaginationItem>
                                        <PaginationLink onClick={(e) => { e.preventDefault(); setPage(page - 1); }} className="cursor-pointer">
                                            {page - 1}
                                        </PaginationLink>
                                    </PaginationItem>
                                )}
                                {/* 当前页 */}
                                <PaginationItem>
                                    <PaginationLink isActive className="cursor-default">
                                        {page}
                                    </PaginationLink>
                                </PaginationItem>
                                {/* 下一页 */}
                                {page < totalPages && (
                                    <PaginationItem>
                                        <PaginationLink onClick={(e) => { e.preventDefault(); setPage(page + 1); }} className="cursor-pointer">
                                            {page + 1}
                                        </PaginationLink>
                                    </PaginationItem>
                                )}
                                {/* 右省略 */}
                                {page < totalPages - 2 && (
                                    <PaginationItem>
                                        <PaginationEllipsis />
                                    </PaginationItem>
                                )}
                                {/* 最后一页 */}
                                {page < totalPages - 1 && totalPages > 1 && (
                                    <PaginationItem>
                                        <PaginationLink onClick={(e) => { e.preventDefault(); setPage(totalPages); }} className="cursor-pointer">
                                            {totalPages}
                                        </PaginationLink>
                                    </PaginationItem>
                                )}
                                <PaginationItem>
                                    <PaginationNext
                                        onClick={(e) => { e.preventDefault(); if (page < totalPages) setPage(p => p + 1); }}
                                        className={page >= totalPages ? "pointer-events-none opacity-50" : "cursor-pointer"}
                                        aria-disabled={page >= totalPages}
                                    />
                                </PaginationItem>
                            </PaginationContent>
                        </Pagination>
                    </div>
                </CardContent>
            </Card>

            {/* 添加新词对话框 */}
            <Dialog open={showAddModal} onOpenChange={setShowAddModal}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>添加新词</DialogTitle>
                        <DialogDescription>
                            添加到用户词库(User)中，优先级最高。
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-3 py-2">
                        <div className="space-y-1">
                            <label className="text-sm font-medium">词语</label>
                            <Input
                                placeholder="输入词语"
                                value={newWord}
                                onChange={(e) => setNewWord(e.target.value)}
                                autoFocus
                            />
                        </div>
                        <div className="space-y-1">
                            <label className="text-sm font-medium">词频</label>
                            <Input
                                type="number"
                                placeholder="100"
                                value={newFreq}
                                onChange={(e) => setNewFreq(e.target.value)}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="secondary" onClick={() => setShowAddModal(false)}>取消</Button>
                        <Button onClick={handleAdd}>添加</Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* 删除确认对话框 */}
            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => !open && setDeleteTarget(null)}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认删除</AlertDialogTitle>
                        <AlertDialogDescription>
                            {deleteTarget?.type === 'single'
                                ? `确定要删除词语 "${deleteTarget.word}" 吗？此操作无法撤销。`
                                : `确定要删除选中的 ${selected.size} 个词语吗？此操作无法撤销。`
                            }
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction
                            onClick={() => deleteTarget?.type === 'single' && deleteTarget.word
                                ? handleDelete(deleteTarget.word)
                                : handleBatchDelete()
                            }
                            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            删除
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>

            {/* 合并确认对话框 */}
            <AlertDialog open={mergeConfirmOpen} onOpenChange={setMergeConfirmOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>确认归档</AlertDialogTitle>
                        <AlertDialogDescription>
                            确定要将所有暂存词合并到用户词库吗？这将立即生效到生产环境。
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel>取消</AlertDialogCancel>
                        <AlertDialogAction onClick={handleMerge}>
                            确认归档
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </div>
    );
}
