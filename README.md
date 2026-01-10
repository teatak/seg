# Seg (Self-Evolving Segmentation System)

> **高性能、自进化、端到端中文分词引擎**

Seg 是一款结合了传统词图（DAG）与现代机器学习（CRF）优势的中文分词系统。它最大的特色在于引入了 **“闭环自进化”** 体系，能够根据用户交互和原始语料自动发现新词、重构词库并重新训练模型，实现真正的生产级自适应能力。

---

## 🌟 核心特性

- **🚀 自进化管道 (Self-Evolution)**: 一键触发“新词发现 -> 语料洗牌 -> 差量合并 -> 模型重训”。
- **🛡️ 品牌保护与抗干扰**: 内置专为品牌识别优化的算法，杜绝“美城希尔顿”等分词碎片的产生。
- **🧩 分层字典体系**: 
  - `Base`: 顶级规则词库，手动维护，优先级最高。
  - `Core`: 核心统计词库，由系统扫描全量语料自动生成。
  - `User`: 用户反馈补丁，通过 UI 交互实时沉淀。
- **🎯 混合动力引擎**: 同时支持高效率的 DAG 匹配和高精度的 CRF 序列标注。
- **📊 交互式修正界面**: 可视化调整分词结果，点击“缝隙”即可拆分或合并词语。

---

## 🚀 快速开始

### 1. 启动服务
```bash
go run cmd/server/main.go
```
启动后访问 `http://localhost:8080` 进入可视化交互面板。

### 2. 命令行交互 (CLI)
```bash
# 单次分词
go run cmd/seg/main.go "北京信息科技大学，简称信息科大，北信，是中华人民共和国首都北京市的一所全日制公办本科大学。由原机械部所属北京机械工业学院和原电子部所属北京信息工程学院合并组建，是北京市重点支持建设的信息学科较为齐全的高校。"

# 搜索引擎模式 (长词再切分)
go run cmd/seg/main.go -func=search "北京信息科技大学"
# 输出: 北京 / 信息 / 科技 / 大学 / 科技大学 / 北京信息科技大学
```

### 3. 使用 Makefile (推荐)
```bash
make run    # 启动 Web 服务
make cli    # 启动命令行分词工具
make build  # 编译生成 bin/seg 和 bin/server
make test   # 运行单元测试
make clean  # 清理临时文件
```

---

## 📖 API 文档

### 1. 核心分词接口 `/segment`
**Method**: `POST` | **Endpoint**: `/segment`

| 字段 | 说明 |
| :--- | :--- |
| `text` | 待分词的原始文本 |
| `algorithm` | `hybrid` (推荐), `crf`, `dag` |
| `function` | `standard` (默认), `search` (搜索引擎模式) |

**CURL 示例**:
```bash
curl -X POST http://localhost:8080/segment \
  -H "Content-Type: application/json" \
  -d '{"text": "希尔顿欢朋酒店北京市朝阳区", "algorithm": "hybrid"}'
```

### 2. 用户反馈接口 `/feedback`
当用户在 UI 上点击“确认修正”时，会通过此接口提交纠错信息。
**Method**: `GET` | **Endpoint**: `/feedback?word=北京市`

---

## 💻 开发者集成 (Go Library)

```go
import (
    "github.com/teatak/seg/dictionary"
    "github.com/teatak/seg/segmenter"
    "github.com/teatak/seg/crf"
)

func main() {
    // 1. 初始化分层词典
    dict := dictionary.NewDictionary()
    dict.Load("data/dict_base.txt") // 核心品牌
    dict.Load("data/dict_core.txt") // 语料基础
    dict.Load("data/dict_user.txt") // 用户补丁

    // 2. 构造分词器并加载模型
    seg := segmenter.NewSegmenter(dict)
    model := crf.NewModel()
    model.Load("data/model.crf")
    seg.CRFModel = model

    // 3. 执行分词
    // 标准模式 (Standard)
    tokens := seg.Cut("北京信息科技大学", segmenter.ModeHybrid)
    // 结果: [北京信息科技大学]

    // 搜索引擎模式 (Search Mode)
    // 会对长词进行细粒度切分，提高召回率
    searchTokens := seg.CutSearch("北京信息科技大学", segmenter.ModeHybrid)
    // 结果: [北京, 信息, 科技, 大学, 科技大学, 北京信息科技大学]
}
```

---

## ⚙️ 进化流水线 (Self-Evolution Pipeline)

当你在界面点击 **「确认修正并启动自进化训练」** 或 **「触发自动进化」** 时，后台会依次执行：
1. **反馈吸收**：将当前纠错写入 `dict_user.txt`。
2. **潜在新词挖掘**：扫描 `text.txt` 原始语料，基于 N-Gram 统计发现高频重复模式。
3. **全局语料洗牌 (Back-Washing)**：利用当前最新的词库对全量语料重新分词，纠正模型偏见。
4. **CRF 模型重构**：基于洗出的语料全量重新训练 `model.crf`。
5. **热加载**：无需重启服务，模型和词典即时切换。

---

## 🛠 项目架构

```text
.
├── cmd/           # 工具入口 (server, seg, train_crf)
├── optimizer/     # 核心优化模块 (新词发现, 语料洗理, 训练调度)
├── segmenter/     # 分词逻辑核心 (DAG & Hybrid)
├── dictionary/    # 词典管理 (双向序列化, 优先级覆盖)
├── crf/           # CRF 模型算法实现
├── data/          # 数据资产 (词典、语料、模型)
└── static/        # 可视化 UI 资源
```

---

## 📜 开源协议
MIT License
