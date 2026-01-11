package seg

import (
	"encoding/json"
	"math"
	"os"
)

// HMM 隐马尔可夫模型
// 状态集: B(词首), M(词中), E(词尾), S(单字词)
type HMM struct {
	// 初始状态概率 P(state)
	InitProb map[string]float64 `json:"init_prob"`
	// 状态转移概率 P(state2|state1)
	TransProb map[string]map[string]float64 `json:"trans_prob"`
	// 发射概率 P(char|state)
	EmitProb map[string]map[rune]float64 `json:"emit_prob"`
}

// 状态常量
const (
	StateB = "B" // 词首
	StateM = "M" // 词中
	StateE = "E" // 词尾
	StateS = "S" // 单字词
)

// 合法状态转移
var validTrans = map[string][]string{
	StateB: {StateM, StateE},
	StateM: {StateM, StateE},
	StateE: {StateB, StateS},
	StateS: {StateB, StateS},
}

// 起始状态
var startStates = []string{StateB, StateS}

// 结束状态
var endStates = map[string]bool{StateE: true, StateS: true}

// NewHMM 创建 HMM 模型
func NewHMM() *HMM {
	return &HMM{
		InitProb:  make(map[string]float64),
		TransProb: make(map[string]map[string]float64),
		EmitProb:  make(map[string]map[rune]float64),
	}
}

// NewDefaultHMM 创建带有默认参数的 HMM 模型
func NewDefaultHMM() *HMM {
	hmm := NewHMM()

	// 初始状态概率（对数概率）
	hmm.InitProb = map[string]float64{
		StateB: -0.26268660809250016,
		StateE: -3.14e+100, // 负无穷大
		StateM: -3.14e+100,
		StateS: -1.4652633398537678,
	}

	// 状态转移概率（对数概率）
	hmm.TransProb = map[string]map[string]float64{
		StateB: {
			StateE: -0.510825623765990,
			StateM: -0.916290731874155,
		},
		StateE: {
			StateB: -0.5897149736854513,
			StateS: -0.8085250474669937,
		},
		StateM: {
			StateE: -0.33344856811948514,
			StateM: -1.2603623820268226,
		},
		StateS: {
			StateB: -0.7211965654669841,
			StateS: -0.6658631448798212,
		},
	}

	// 发射概率使用默认值（真实场景需要从语料训练）
	hmm.EmitProb = map[string]map[rune]float64{
		StateB: {},
		StateM: {},
		StateE: {},
		StateS: {},
	}

	return hmm
}

// LoadFromFile 从文件加载模型参数
func (h *HMM) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, h)
}

// SaveToFile 保存模型参数到文件
func (h *HMM) SaveToFile(filepath string) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// Segment 使用 Viterbi 算法分词
func (h *HMM) Segment(text string, offset int) []Token {
	runes := []rune(text)
	if len(runes) == 0 {
		return nil
	}

	// 只有一个字
	if len(runes) == 1 {
		return []Token{{
			Word:  text,
			Start: offset,
			End:   offset + 1,
			Type:  "hmm",
		}}
	}

	// Viterbi 算法
	states := h.viterbi(runes)

	// 根据状态序列切分
	return h.stateToTokens(runes, states, offset)
}

// viterbi Viterbi 算法求解最优状态序列
func (h *HMM) viterbi(runes []rune) []string {
	n := len(runes)
	states := []string{StateB, StateM, StateE, StateS}

	// dp[i][state] = 到位置 i 时处于 state 的最大概率
	// path[i][state] = 到位置 i 时处于 state 的最优前驱状态
	dp := make([]map[string]float64, n)
	path := make([]map[string]string, n)

	for i := 0; i < n; i++ {
		dp[i] = make(map[string]float64)
		path[i] = make(map[string]string)
	}

	// 初始化
	for _, state := range startStates {
		dp[0][state] = h.InitProb[state] + h.getEmitProb(state, runes[0])
	}

	// 递推
	for i := 1; i < n; i++ {
		for _, state := range states {
			maxProb := math.Inf(-1)
			maxPrev := ""

			for _, prev := range states {
				// 检查转移是否合法
				if !h.canTrans(prev, state) {
					continue
				}

				prob := dp[i-1][prev] + h.getTransProb(prev, state) + h.getEmitProb(state, runes[i])
				if prob > maxProb {
					maxProb = prob
					maxPrev = prev
				}
			}

			if maxProb > math.Inf(-1) {
				dp[i][state] = maxProb
				path[i][state] = maxPrev
			}
		}
	}

	// 回溯找最优状态序列
	result := make([]string, n)

	// 找结束状态
	maxProb := math.Inf(-1)
	var lastState string
	for state := range endStates {
		if prob, ok := dp[n-1][state]; ok && prob > maxProb {
			maxProb = prob
			lastState = state
		}
	}

	// 如果没有合法结束状态，选择概率最大的
	if lastState == "" {
		for _, state := range states {
			if prob, ok := dp[n-1][state]; ok && prob > maxProb {
				maxProb = prob
				lastState = state
			}
		}
	}

	// 回溯
	result[n-1] = lastState
	for i := n - 1; i > 0; i-- {
		result[i-1] = path[i][result[i]]
	}

	return result
}

func (h *HMM) canTrans(from, to string) bool {
	for _, s := range validTrans[from] {
		if s == to {
			return true
		}
	}
	return false
}

func (h *HMM) getTransProb(from, to string) float64 {
	if probs, ok := h.TransProb[from]; ok {
		if prob, ok := probs[to]; ok {
			return prob
		}
	}
	return -3.14e+100 // 负无穷大
}

func (h *HMM) getEmitProb(state string, char rune) float64 {
	if probs, ok := h.EmitProb[state]; ok {
		if prob, ok := probs[char]; ok {
			return prob
		}
	}
	// 未登录字符使用平滑值
	return -10.0
}

// stateToTokens 将状态序列转换为词语
func (h *HMM) stateToTokens(runes []rune, states []string, offset int) []Token {
	var tokens []Token
	start := 0

	for i, state := range states {
		switch state {
		case StateS:
			tokens = append(tokens, Token{
				Word:  string(runes[i]),
				Start: offset + i,
				End:   offset + i + 1,
				Type:  "hmm",
			})
			start = i + 1
		case StateE:
			tokens = append(tokens, Token{
				Word:  string(runes[start : i+1]),
				Start: offset + start,
				End:   offset + i + 1,
				Type:  "hmm",
			})
			start = i + 1
		}
	}

	// 处理未结束的词
	if start < len(runes) {
		tokens = append(tokens, Token{
			Word:  string(runes[start:]),
			Start: offset + start,
			End:   offset + len(runes),
			Type:  "hmm",
		})
	}

	return tokens
}

// Train 从标注语料训练 HMM 参数
func (h *HMM) Train(corpus [][]string) {
	// 统计计数
	initCount := make(map[string]int)
	transCount := make(map[string]map[string]int)
	emitCount := make(map[string]map[rune]int)
	stateCount := make(map[string]int)

	for _, state := range []string{StateB, StateM, StateE, StateS} {
		transCount[state] = make(map[string]int)
		emitCount[state] = make(map[rune]int)
	}

	// 遍历语料
	for _, sentence := range corpus {
		var prevState string
		for _, word := range sentence {
			runes := []rune(word)
			if len(runes) == 0 {
				continue
			}

			var states []string
			if len(runes) == 1 {
				states = []string{StateS}
			} else {
				states = make([]string, len(runes))
				states[0] = StateB
				for i := 1; i < len(runes)-1; i++ {
					states[i] = StateM
				}
				states[len(runes)-1] = StateE
			}

			for i, state := range states {
				stateCount[state]++
				emitCount[state][runes[i]]++

				if i == 0 && prevState == "" {
					initCount[state]++
				} else if i == 0 && prevState != "" {
					transCount[prevState][state]++
				} else {
					transCount[states[i-1]][state]++
				}
			}
			prevState = states[len(states)-1]
		}
	}

	// 计算概率（对数概率）
	totalInit := 0
	for _, c := range initCount {
		totalInit += c
	}
	for state, count := range initCount {
		h.InitProb[state] = math.Log(float64(count) / float64(totalInit))
	}

	for from, tos := range transCount {
		if h.TransProb[from] == nil {
			h.TransProb[from] = make(map[string]float64)
		}
		total := 0
		for _, c := range tos {
			total += c
		}
		for to, count := range tos {
			h.TransProb[from][to] = math.Log(float64(count) / float64(total))
		}
	}

	for state, chars := range emitCount {
		if h.EmitProb[state] == nil {
			h.EmitProb[state] = make(map[rune]float64)
		}
		for char, count := range chars {
			h.EmitProb[state][char] = math.Log(float64(count) / float64(stateCount[state]))
		}
	}
}
