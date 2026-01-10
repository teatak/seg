const API_HOST = window.location.origin;

// State for interactive segmentation
let currentChars = [];
let currentSplits = []; // boolean array, size = chars.length - 1

function log(msg, type = 'info') {
    const logs = document.getElementById('systemLogs');
    const color = type === 'error' ? 'text-red-400' : (type === 'success' ? 'text-green-400' : 'text-slate-300');
    const time = new Date().toLocaleTimeString();
    logs.innerHTML += `<div class="mb-1"><span class="opacity-50">[${time}]</span> <span class="${color}">${msg}</span></div>`;
    logs.scrollTop = logs.scrollHeight;
}

async function runSegment() {
    const text = document.getElementById('inputText').value.trim();
    if (!text) return;

    const algorithm = document.querySelector('input[name="algorithm"]:checked').value;
    const func = document.querySelector('input[name="function"]:checked').value;
    const start = performance.now();

    try {
        const res = await fetch(`${API_HOST}/segment`, {
            method: 'POST',
            body: JSON.stringify({ text, algorithm, function: func })
        });
        const data = await res.json();
        const end = performance.now();

        // Initialize interactive editor
        initInteractiveEditor(data.tokens);

        document.getElementById('latency').innerText = `后端耗时: ${(end - start).toFixed(1)}ms`;
        document.getElementById('resultArea').classList.remove('hidden');
        
        log(`Segmented "${text.substring(0, 10)}..." (${data.tokens.length} tokens)`);
    } catch (e) {
        console.error(e);
        log(`Error: ${e.message}`, 'error');
    }
}

function initInteractiveEditor(tokens) {
    currentChars = [];
    currentSplits = [];
    
    // Flatten tokens into chars and build split array
    // Example: ["AB", "C"] -> chars=["A","B","C"], splits=[false, true]
    // split[0] is between char 0 and 1
    
    for (let i = 0; i < tokens.length; i++) {
        const token = tokens[i];
        const chars = Array.from(token);
        for (let j = 0; j < chars.length; j++) {
            currentChars.push(chars[j]);
            if (j < chars.length - 1) {
                currentSplits.push(false); // Inside a token
            }
        }
        // End of token, if not the very last token
        if (i < tokens.length - 1) {
            currentSplits.push(true); // Split between tokens
        }
    }
    
    renderInteractive();
}

function renderInteractive() {
    const container = document.getElementById('interactiveContainer');
    container.innerHTML = '';
    
    const wrapper = document.createElement('div');
    wrapper.className = 'segment-container';
    
    for (let i = 0; i < currentChars.length; i++) {
        // Character
        const charSpan = document.createElement('span');
        charSpan.innerText = currentChars[i];
        charSpan.className = 'char text-slate-800';
        
        // Color coding for visual grouping (optional, maybe distinct colors for words)
        // Let's settle for simple slate-800 for now, or alternate colors per word?
        // Alternate colors might be nice.
        const wordIndex = getWordIndex(i);
        // Assign a subtle background or color to alternate words? 
        // Maybe just text color alternation: slate-900 / slate-600
        if (wordIndex % 2 === 0) {
            charSpan.classList.add('text-slate-900');
        } else {
            charSpan.classList.add('text-indigo-600'); // Alternate color
        }

        wrapper.appendChild(charSpan);

        // Gap (only if not last char)
        if (i < currentChars.length - 1) {
            const gap = document.createElement('div');
            const isSplit = currentSplits[i];
            gap.className = `gap ${isSplit ? 'active' : 'inactive'}`;
            gap.onclick = () => toggleSplit(i);
            wrapper.appendChild(gap);
        }
    }
    
    container.appendChild(wrapper);
}

function getWordIndex(charIndex) {
    // Calculate which word this char belongs to based on currentSplits
    let wIdx = 0;
    for(let i=0; i<charIndex; i++) {
        if (currentSplits[i]) wIdx++;
    }
    return wIdx;
}

function toggleSplit(index) {
    currentSplits[index] = !currentSplits[index];
    renderInteractive();
}

function getResultString() {
    let result = "";
    for (let i = 0; i < currentChars.length; i++) {
        result += currentChars[i];
        if (i < currentSplits.length && currentSplits[i]) {
            result += " "; // Split
        }
    }
    return result;
}

async function submitIntervention() {
    const wordWord = getResultString().trim();
    if (!wordWord) return;
    
    log(`Learning: "${wordWord}"...`);
    const btn = document.getElementById('btn-train');
    btn.disabled = true;
    const originalText = btn.innerHTML;
    btn.innerHTML = `Training...`;

    try {
        const res = await fetch(`${API_HOST}/feedback?word=${encodeURIComponent(wordWord)}`);
        const text = await res.text();
        
        if (res.ok) {
            log(`Success: ${text}`, 'success');
        } else {
            log(`Failed: ${text}`, 'error');
        }
    } catch (e) {
        log(`Feedback Error: ${e.message}`, 'error');
    } finally {
        btn.disabled = false;
        btn.innerHTML = originalText;
    }
}

async function triggerDiscovery() {
    if(!confirm("确定要触发自动发现流程吗？这将分析服务器日志并重新训练模型。可能需要几十秒。")) return;

    const btn = document.getElementById('btn-discover');
    btn.disabled = true;
    btn.style.opacity = "0.7";
    log("Triggering auto-discovery pipeline...");

    try {
        const res = await fetch(`${API_HOST}/trigger-discovery`);
        const text = await res.text();
        log(text, 'success');
    } catch (e) {
        log(`Discovery Error: ${e.message}`, 'error');
    } finally {
        btn.disabled = false;
        btn.style.opacity = "1";
    }
}
