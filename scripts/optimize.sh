#!/bin/bash

# optimize.sh - 自动化词典更新与模型重训脚本
# 使用方法: ./optimize.sh <新词文件>
# 示例: ./optimize.sh new_words.txt

set -e # 遇到错误立即停止

# 路径配置
DATA_DIR="data"
DICT_FILE="$DATA_DIR/dictionary.txt"
DICT_CLEAN_FILE="$DATA_DIR/dictionary_clean.txt"
TEXT_FILE="$DATA_DIR/text.txt"
CORPUS_FILE="$DATA_DIR/corpus.txt"
MODEL_FILE="$DATA_DIR/crf_model.txt"

NEW_WORDS_FILE=$1

# 1. 检查输入
if [ -z "$NEW_WORDS_FILE" ]; then
    echo "Usage: ./optimize.sh <path_to_new_words_file>"
    echo "Example: echo '人工智能 1000' > new.txt && ./optimize.sh new.txt"
    exit 1
fi

if [ ! -f "$NEW_WORDS_FILE" ]; then
    echo "Error: File $NEW_WORDS_FILE not found."
    exit 1
fi

echo "=== 开始优化流程 ==="
echo "时间: $(date)"

# 2. 备份当前词典 (防止搞坏)
echo "[1/5] 备份当前词典..."
cp "$DICT_FILE" "${DICT_FILE}.bak"

# 3. 合并新词
# 假设新词文件格式可以是 "单词" 或者 "单词 频率"
# 我们统一处理：如果是纯单词，默认给个高频 (如 10000)
# 如果已有频率，则保留。
# 为了简单，我们先简单追加，然后让 sort/uniq 处理去重（注意：这里简单的 uniq 不足以合并频率，但 clean_dict 会处理）
echo "[2/5] 合并新词到词典..."
# Check formatting. If just word, append default freq.
while read -r line; do
    if [[ "$line" =~ ^[[:space:]]*$ ]]; then continue; fi
    
    # Check if line has a number
    if [[ "$line" =~ [0-9]+$ ]]; then
        echo "$line" >> "$DICT_FILE"
    else
        echo "$line 10000" >> "$DICT_FILE"
    fi
done < "$NEW_WORDS_FILE"

echo "已追加新词。开始清洗..."

# 4. 清洗词典
# 使用我们写的 Go 工具进行清洗和去重
go run cmd/clean_dict/main.go -input "$DICT_FILE" -output "$DICT_CLEAN_FILE"
mv "$DICT_CLEAN_FILE" "$DICT_FILE"
echo "词典清洗完成。"

# 5. 重新生成语料
echo "[3/5] 基于新词典重新切分训练语料..."
# 注意：这步比较耗时，取决于 text.txt 大小
go run cmd/batch_seg/main.go -input "$TEXT_FILE" -output "$CORPUS_FILE" -dict "$DICT_FILE"

# 6. 重新训练模型
echo "[4/5] 重新训练 CRF 模型 (Iter=10)..."
go run cmd/train_crf/main.go -input "$CORPUS_FILE" -output "$MODEL_FILE" -iter 10

# 7. 验证
echo "[5/5] 验证核心 Case..."
# 可以把一些必须要对的测试用例写死在这里，或者从 test_cases.txt 读取
# 这里简单测试一下 newly added words 如果在 new_words_file 里
echo "抽取新词文件中的第一个词进行测试:"
FIRST_WORD=$(head -n 1 "$NEW_WORDS_FILE" | awk '{print $1}')
if [ ! -z "$FIRST_WORD" ]; then
    echo "测试词: $FIRST_WORD"
    # 构造一个简单的句子: "测试一下[WORD]的效果"
    TEST_SENT="我正在测试${FIRST_WORD}的效果"
    echo "输入: $TEST_SENT"
    OUTPUT=$(go run cmd/seg/main.go -mode hybrid "$TEST_SENT")
    echo "输出: $OUTPUT"
    
    # 简单检查输出是否包含该词 (包含即说明没被切开)
    if [[ "$OUTPUT" == *"$FIRST_WORD"* ]]; then
        echo "✅ 验证通过: 新词被正确识别。"
    else
        echo "⚠️ 警告: 新词似乎仍被切分，请检查频率设置或冲突词。"
    fi
fi

echo "=== 优化流程结束 ==="
