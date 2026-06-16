import {
  ExportPollutionExcel as wailsExportPollutionExcel,
  ImportTXT as wailsImportTXT,
  StartDetection as wailsStartDetection,
} from '../../wailsjs/go/main/App'

const mockTargets = [
  '80823456.com',
  '8088w.com',
  '80906.vip',
  '8090xiaoyaoju.com',
  '80969t.com',
  '8099r.com',
  '80gu.com',
  '80suncity.com',
  '80ych.top',
  '810180.com',
  '810461.com',
  '810msc.com',
  '8111777.com',
  '811k3.com',
  '8156d.com',
  '815749.org',
  '816002.com',
  '817890.com',
  '8181138.com',
  '818428.com',
  '81850zq.com',
  '81852.cc',
  '818737.cc',
  '8188q.com',
  '818988c.com',
  '818yule.net',
  '81966.com',
  '81999.cc',
  '820003.com',
  '82023.net',
  '821138.com',
  '82226.vip',
]

function hasWailsBridge() {
  return Boolean(window.go?.main?.App)
}

function makeRows(targets = mockTargets, request = {}) {
  const rows = []
  if (request.enablePollution) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: '污染',
        target,
        status: '正常',
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableHijack) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: 'QQ拦截',
        target,
        status: index % 9 === 0 ? '拦截' : '正常',
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableWechat) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: '微信拦截',
        target,
        status: index % 7 === 0 ? '拦截' : '正常',
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableIcp) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: '备案查询',
        target,
        status: index % 8 === 0 ? '未备案' : '已备案',
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
        domain: target,
        beianCode: `测试ICP备案号-${index + 1}`,
        siteName: `测试网站${index + 1}`,
      })
    })
  }
  if (request.enableBlacklist) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: '备案黑名单',
        target,
        status: index % 10 === 0 ? '黑名单' : '正常',
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableWall) {
    targets.forEach((target, index) => {
      let status = '正常'
      if (index % 11 === 0) {
        status = '被墙'
      } else if (index % 7 === 0) {
        status = '疑似被墙'
      }

      rows.push({
        id: rows.length + 1,
        type: '被墙检测',
        target,
        status,
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }

  return rows
}

function summarizeRows(rows, request) {
  if (request.enablePollution && !request.enableHijack && !request.enableWechat && !request.enableIcp && !request.enableBlacklist && !request.enableWall) {
    return {
      total: 93874,
      checked: 93874,
      pollution: 1498,
      normal: 92155,
      unregistered: 203,
      failed: 18,
    }
  }

  return rows.reduce((summary, row) => {
    summary.total += 1
    summary.checked += 1
    if (row.status === '污染' || row.status === '拦截' || row.status === '未备案' || row.status === '黑名单' || row.status === '被墙' || row.status === '疑似被墙' || row.status === '域名格式错误') {
      summary.pollution += 1
    } else if (row.status === '正常' || row.status === '已备案') {
      summary.normal += 1
    } else if (row.status === '未注册') {
      summary.unregistered += 1
    } else if (row.status === '失败') {
      summary.failed += 1
    }
    return summary
  }, {
    total: 0,
    checked: 0,
    pollution: 0,
    normal: 0,
    unregistered: 0,
    failed: 0,
  })
}

export function ImportTXT() {
  if (hasWailsBridge()) {
    return wailsImportTXT()
  }

  return Promise.resolve({
    targets: mockTargets.slice(0, 10),
    message: '开发预览: 已导入测试目标 10 条',
    canceled: false,
  })
}

export function StartDetection(request) {
  if (hasWailsBridge()) {
    return wailsStartDetection(request)
  }

  const targets = request.targets?.length ? request.targets : mockTargets
  const rows = makeRows(targets, request)
  return Promise.resolve({
    rows,
    summary: summarizeRows(rows, request),
    progress: 100,
    exportPath: 'C:/Users/xjw/Desktop/全部污染.xlsx',
    message: request.apiKey
      ? `开发预览使用测试数据，并发 ${request.concurrency}，超时 ${request.timeoutSeconds} 秒`
      : `未填写有效 API Key，返回测试数据，并发 ${request.concurrency}，超时 ${request.timeoutSeconds} 秒`,
  })
}

export function ExportPollutionExcel(rows = []) {
  if (hasWailsBridge()) {
    return wailsExportPollutionExcel(rows)
  }

  return Promise.resolve({
    path: 'C:/Users/xjw/Desktop/全部污染.xlsx',
    message: `开发预览: 已导出 ${rows.length} 条当前列表数据`,
  })
}
