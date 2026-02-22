// 获取 DOM
const nodeListEl = document.getElementById('node-list');
const nodeTitleEl = document.getElementById('current-node-title');
const nodeSummaryEl = document.getElementById('current-node-summary');

// 图表实例字典
let charts = {
    ping: null,
    resources: null,
    net: null
};

// 当前状态
let activeNodeId = null;
let pollInterval = null;

// 初始化 ECharts
function initCharts() {
    // 使用 Echarts 暗黑主题初始色
    const theme = 'dark';
    const opts = { backgroundColor: 'transparent' };

    charts.ping = echarts.init(document.getElementById('chart-ping'), theme, opts);
    charts.resources = echarts.init(document.getElementById('chart-resources'), theme, opts);
    charts.net = echarts.init(document.getElementById('chart-net'), theme, opts);

    window.addEventListener('resize', () => {
        charts.ping.resize();
        charts.resources.resize();
        charts.net.resize();
    });
}

// 获取全部探针节点
async function fetchNodes() {
    try {
        const res = await fetch('/api/nodes');
        const nodes = await res.json();
        renderNodeList(nodes);
    } catch (e) {
        console.error("Failed to fetch nodes", e);
        nodeListEl.innerHTML = `<div class="loading-text" style="color:#ff3366">Connection Lost to Controller</div>`;
    }
}

// 渲染左侧探针列表
function renderNodeList(nodes) {
    if (!nodes || nodes.length === 0) {
        nodeListEl.innerHTML = `<div class="loading-text">No probes found. Waiting...</div>`;
        return;
    }

    let html = '';
    nodes.forEach(n => {
        const isActive = n.node_id === activeNodeId ? 'active' : '';
        const statusClass = n.is_online ? 'online' : 'offline';
        const lastSeen = new Date(n.last_seen).toLocaleTimeString();

        html += `
            <div class="node-card ${isActive}" onclick="selectNode('${n.node_id}')">
                <div class="node-header">
                    <span class="node-id">${n.node_id}</span>
                    <span class="node-status ${statusClass}"></span>
                </div>
                <div class="node-meta">
                    Last Seen: ${lastSeen} <br>
                    Points: ${n.history ? n.history.length : 0}
                </div>
            </div>
        `;
    });

    nodeListEl.innerHTML = html;

    // 默认选中第一个
    if (!activeNodeId && nodes.length > 0) {
        selectNode(nodes[0].node_id);
    }
}

// 点击侧边栏切换节点
function selectNode(nodeId) {
    activeNodeId = nodeId;
    nodeTitleEl.innerText = `Probe: ${nodeId}`;

    // 立即拉取一次该节点历史
    fetchNodeMetrics();
    fetchNodes(); // 刷新高亮状态

    // 建立 3 秒轮询
    if (pollInterval) clearInterval(pollInterval);
    pollInterval = setInterval(() => {
        fetchNodes();
        fetchNodeMetrics();
    }, 3000);
}

// 获取当前选中节点的高频特征历史流
async function fetchNodeMetrics() {
    if (!activeNodeId) return;
    try {
        const res = await fetch(`/api/metrics?node_id=${activeNodeId}`);
        const history = await res.json();

        if (history && history.length > 0) {
            updateDashboard(history);
        }
    } catch (e) {
        console.error(`Failed to fetch metrics for ${activeNodeId}`, e);
    }
}

// 格式化时间为 HH:mm:ss
function formatTime(ts) {
    const d = new Date(ts);
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}:${d.getSeconds().toString().padStart(2, '0')}`;
}

// 将 JSON 历史填入 ECharts 并渲染汇总看板
function updateDashboard(history) {
    if (!history || history.length === 0) return;

    // 1. 更新顶部汇总信息取最新一条
    const latest = history[history.length - 1];

    // 防错处理
    const rtt = latest.ping_avg_rtt ? latest.ping_avg_rtt.toFixed(2) : "0.00";
    const cpu = latest.cpu_load1 !== undefined ? latest.cpu_load1.toFixed(2) : "0.00";
    const mem = latest.mem_used_percent !== undefined ? latest.mem_used_percent.toFixed(1) : "0.0";
    const net = latest.net_burst !== undefined ? latest.net_burst : 0;

    nodeSummaryEl.innerHTML = `
        <div class="metric-item">
            <span class="metric-label">RTT 延迟 (ms)</span>
            <span class="metric-val">${rtt}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">CPU (Load1)</span>
            <span class="metric-val">${cpu}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">MEM Used (%)</span>
            <span class="metric-val">${mem}</span>
        </div>
        <div class="metric-item">
            <span class="metric-label">NET Microburst</span>
            <span class="metric-val">${net} evt</span>
        </div>
    `;

    // 2. 剥离时间轴与其他曲线 Y 轴
    const timeAxis = history.map(item => formatTime(item.timestamp));
    const pingData = history.map(item => item.ping_avg_rtt || 0);
    const cpuData = history.map(item => item.cpu_load1 || 0);
    const memData = history.map(item => item.mem_used_percent || 0);
    const netBurstData = history.map(item => item.net_burst || 0);

    // 通用基础配置
    const commonOpts = {
        grid: { top: 40, right: 20, bottom: 30, left: 50 },
        tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } },
        xAxis: { type: 'category', data: timeAxis, boundaryGap: false, splitLine: { show: false } },
    };

    // 绘制 Ping 图 (高光赛博蓝)
    charts.ping.setOption({
        ...commonOpts,
        yAxis: { type: 'value', splitLine: { lineStyle: { color: 'rgba(255,255,255,0.05)' } } },
        series: [{
            name: 'TCP RTT (ms)',
            type: 'line',
            smooth: true,
            symbol: 'none',
            itemStyle: { color: '#00f0ff' },
            areaStyle: {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                    { offset: 0, color: 'rgba(0, 240, 255, 0.5)' },
                    { offset: 1, color: 'rgba(0, 240, 255, 0)' }
                ])
            },
            data: pingData
        }]
    }, true);

    // 绘制 系统占用图 (CPU / MEM 双折线)
    charts.resources.setOption({
        ...commonOpts,
        legend: { data: ['CPU Load1', 'Memory %'], right: 10, top: 0, textStyle: { color: '#ccc' } },
        yAxis: { type: 'value', splitLine: { lineStyle: { color: 'rgba(255,255,255,0.05)' } } },
        series: [
            {
                name: 'CPU Load1', type: 'line', smooth: true, symbol: 'none',
                itemStyle: { color: '#ff3366' }, data: cpuData
            },
            {
                name: 'Memory %', type: 'line', smooth: true, symbol: 'none',
                itemStyle: { color: '#aa00ff' }, data: memData
            }
        ]
    }, true);

    // 绘制 突发网络特征 (柱状图)
    charts.net.setOption({
        ...commonOpts,
        yAxis: { type: 'value', splitLine: { lineStyle: { color: 'rgba(255,255,255,0.05)' } } },
        series: [{
            name: 'Microbursts',
            type: 'bar',
            itemStyle: { color: '#00ff88', borderRadius: [2, 2, 0, 0] },
            data: netBurstData
        }]
    }, true);
}


// Start application
initCharts();
fetchNodes();

// 初次启动设置轮询寻找存活节点
if (!pollInterval) {
    setInterval(fetchNodes, 5000);
}
