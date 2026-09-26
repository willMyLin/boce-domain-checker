import {
  ExportPollutionExcel as wailsExportPollutionExcel,
  GetVersion as wailsGetVersion,
  ImportTXT as wailsImportTXT,
  StartDetection as wailsStartDetection,
} from '../../wailsjs/go/main/App'
import {STATUS_BAD_DOMAIN, STATUS_BLACKLISTED, STATUS_FAILED, STATUS_FILED, STATUS_INTERCEPTED, STATUS_NORMAL, STATUS_POLLUTION, STATUS_SUSPECT_WALLED, STATUS_UNFILED, STATUS_UNREGISTERED, STATUS_WALLED, TYPE_BLACKLIST, TYPE_ICP, TYPE_POLLUTION, TYPE_QQ, TYPE_WALL, TYPE_WECHAT} from '../lib/constants'

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
        type: TYPE_POLLUTION,
        target,
        status: STATUS_NORMAL,
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableHijack) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: TYPE_QQ,
        target,
        status: index % 9 === 0 ? STATUS_INTERCEPTED : STATUS_NORMAL,
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableWechat) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: TYPE_WECHAT,
        target,
        status: index % 7 === 0 ? STATUS_INTERCEPTED : STATUS_NORMAL,
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableIcp) {
    targets.forEach((target, index) => {
      rows.push({
        id: rows.length + 1,
        type: TYPE_ICP,
        target,
        status: index % 8 === 0 ? STATUS_UNFILED : STATUS_FILED,
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
        type: TYPE_BLACKLIST,
        target,
        status: index % 10 === 0 ? STATUS_BLACKLISTED : STATUS_NORMAL,
        checkedAt: `2026-06-09 08:39:${String(40 + Math.floor(index / 2)).padStart(2, '0')}`,
        errorRemark: '',
      })
    })
  }
  if (request.enableWall) {
    targets.forEach((target, index) => {
      let status = STATUS_NORMAL
      if (index % 11 === 0) {
        status = STATUS_WALLED
      } else if (index % 7 === 0) {
        status = STATUS_SUSPECT_WALLED
      }

      rows.push({
        id: rows.length + 1,
        type: TYPE_WALL,
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
    if (row.status === STATUS_POLLUTION || row.status === STATUS_INTERCEPTED || row.status === STATUS_UNFILED || row.status === STATUS_BLACKLISTED || row.status === STATUS_WALLED || row.status === STATUS_SUSPECT_WALLED || row.status === STATUS_BAD_DOMAIN) {
      summary.pollution += 1
    } else if (row.status === STATUS_NORMAL || row.status === STATUS_FILED) {
      summary.normal += 1
    } else if (row.status === STATUS_UNREGISTERED) {
      summary.unregistered += 1
    } else if (row.status === STATUS_FAILED) {
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

export function GetVersion() {
  if (hasWailsBridge()) {
    return wailsGetVersion()
  }

  return Promise.resolve('dev')
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
