# seg

[English](README.md)

一个轻量级、高性能的 Go 语言中文分词库。

## 特性

- **基于词典的分词**：支持加载自定义词典进行精确切分。
- **高效算法**：使用有向无环图（DAG）生成候选路径，利用动态规划（DP）寻找最大概率路径。经过 Slice 优化，具备极高的处理性能。
- **CRF 模型支持**：集成线性链条件随机场（CRF），具备强大的新词（未登录词）识别能力，擅长处理人名、复杂地名及各种新术语。
- **数字/英文保护**：自动识别并保护连续的数字和英文字符序列（如 "7天", "iPhone15", "PKU"），防止被错误切碎。
- **简洁的 API**：统一的接口设计，支持可选的分词模式。

## 安装

```bash
go get github.com/teatak/seg
```

## 快速开始

```go
package main

import (
	"fmt"
	"github.com/teatak/seg/crf"
	"github.com/teatak/seg/dictionary"
	"github.com/teatak/seg/segmenter"
)

func main() {
	// 1. 加载词典
	dict := dictionary.NewDictionary()
	dict.Load("data/dictionary.txt")

	// 2. 初始化分词器
	seg := segmenter.NewSegmenter(dict)

	// (可选) 加载 CRF 模型以启用智能识别
	model := crf.NewModel()
	if err := model.Load("data/crf_model.txt"); err == nil {
		seg.CRFModel = model
	}

	text := "四川省成都市大邑县晋原镇顺兴路25号"

	// 3. 基础用法 (默认使用 ModeDAG)
	parts := seg.Cut(text)
	fmt.Println("DAG 模式:", parts) 

	// 4. 使用 CRF 模式
	crfParts := seg.Cut(text, segmenter.ModeCRF)
	fmt.Println("CRF 模式:", crfParts)

	// 5. 搜索引擎模式 (细粒度切分)
	searchParts := seg.CutSearch(text)
	fmt.Println("搜索模式:", searchParts)
}
```

## 分词模式

- **`segmenter.ModeDAG` (默认)**：适用于通用的词典基础分词，速度极快。
- **`segmenter.ModeCRF`**：适用于人名、复杂地址、以及词典未收录新词的识别场景。

## 命令行工具

你可以通过命令行快速测试分词效果：

```bash
# 默认模式
go run cmd/seg/main.go "你好世界"

# 使用 CRF 模式
go run cmd/seg/main.go -mode crf "顺兴路25号"
```

## 许可证

MIT License
