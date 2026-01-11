# 自学习中文分词系统

一个具备**自我学习**和**持续进化**能力的中文分词系统。

## ✨ 特性

- **基础分词**：双向最大匹配算法 + HMM 处理未登录词
- **三层词库体系**：基础词库 + 用户词库 + 暂存词库，支持灰度发布新词
- **新词发现**：基于互信息和左右熵的自动新词挖掘
- **用户反馈**：收集用户纠正，自动提取新词
- **动态更新**：词典实时热更新，无需重启
- **Web 界面**：内置现代化 React 前端，支持可视化分词测试与管理

## 🚀 快速开始

```bash
# 启动服务
go run cmd/server/main.go

# 访问 Web 界面
open http://localhost:8080
```
## 📦 组件调用 (Go)

如果您希望在现有的 Go 项目中直接集成：

```go
package main

import (
	"fmt"

	"github.com/teatak/seg/pkg/dict"
	"github.com/teatak/seg/pkg/seg"
)

func main() {
	// 1. 初始化词典 (基础词典 + 用户词典 + 暂存词典)
	d := dict.NewDictionary("./data/dict/base.txt", "./data/dict/user.txt", "./data/dict/staging.txt")
	if err := d.Load(); err != nil {
		panic(err)
	}

	// 2. 初始化 HMM 模型 (用于新词识别)
	hmm := seg.NewDefaultHMM()

	// 3. 创建分词器
	// useStaging=true 表示启用暂存词典 (适合开发/评估模式)
	segmenter := seg.NewSegmenter(d, hmm, true)

	// 4. 执行分词
	text := "人工智能正在改变世界"
	tokens := segmenter.Segment(text)

	// 5. 输出结果
	for _, token := range tokens {
		fmt.Printf("%s\t[%s]\n", token.Word, token.Type)
	}
}
```
## 📡 API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/segment` | POST | 分词 |
| `/api/segment/search` | POST | 搜索模式分词 |
| `/api/feedback` | POST/GET | 提交/查询反馈 |
| `/api/words` | GET/POST | 查询/添加词语 |
| `/api/words/{word}` | GET/DELETE | 查询/删除词语 |
| `/api/learn` | POST | 触发学习 |
| `/api/stats` | GET | 系统统计 |

### 分词示例

```bash
curl -X POST http://localhost:8080/api/segment \
  -H "Content-Type: application/json" \
  -d '{"text": "人工智能正在改变世界"}'
```

响应：
```json
{
  "text": "人工智能正在改变世界",
  "tokens": [
    {"word": "人工智能", "start": 0, "end": 4, "type": "word"},
    {"word": "正在", "start": 4, "end": 6, "type": "word"},
    {"word": "改变", "start": 6, "end": 8, "type": "word"},
    {"word": "世界", "start": 8, "end": 10, "type": "word"}
  ],
  "words": ["人工智能", "正在", "改变", "世界"]
}
```

### 提交反馈

```bash
curl -X POST http://localhost:8080/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "text": "机器学习",
    "original_seg": ["机", "器", "学", "习"],
    "corrected_seg": ["机器学习"]
  }'
```

### 从语料学习新词

```bash
curl -X POST http://localhost:8080/api/learn \
  -H "Content-Type: application/json" \
  -d '{
    "type": "corpus",
    "texts": ["深度学习是人工智能的重要分支", "机器学习算法"]
  }'
```

## 📁 项目结构

```
seg/
├── cmd/server/main.go    # HTTP 服务入口
├── pkg/
│   ├── dict/             # 词典模块
│   │   ├── trie.go       # Trie 树
│   │   └── dict.go       # 词典管理
│   ├── seg/              # 分词模块
│   │   ├── segmenter.go  # 分词器
│   │   └── hmm.go        # HMM 模型
│   └── learn/            # 自学习模块
│       ├── miner.go      # 新词发现
│       ├── feedback.go   # 反馈处理
│       └── updater.go    # 词典更新
└── data/
    └── dict/base.txt     # 基础词典
```

## 🧠 核心算法

### 1. 双向最大匹配
结合正向和逆向最大匹配，选择词数最少、单字最少的结果。

### 2. HMM 未登录词处理
使用 Viterbi 算法，状态集 {B, M, E, S}，识别词典外的词语。

### 3. 新词发现
- **互信息 (MI)**：衡量词语内部凝聚度
- **左右熵**：衡量词语边界自由度

### 4. 自学习流程
```
用户反馈 → 新词提取 → 置信度评估 → 动态入库 → 持续优化
```

## 📄 License

MIT
