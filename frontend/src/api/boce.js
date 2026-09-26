import {
  ExportPollutionExcel as wailsExportPollutionExcel,
  ImportTXT as wailsImportTXT,
  StartDetection as wailsStartDetection,
  ValidateTargets as wailsValidateTargets,
} from '../../wailsjs/go/main/App'
import {STATUS_BAD_DOMAIN, STATUS_BLACKLISTED, STATUS_FAILED, STATUS_FILED, STATUS_INTERCEPTED, STATUS_NORMAL, STATUS_POLLUTION, STATUS_SUSPECT_WALLED, STATUS_UNFILED, STATUS_UNREGISTERED, STATUS_UNKNOWN, STATUS_WALLED, TYPE_BLACKLIST, TYPE_ICP, TYPE_POLLUTION, TYPE_QQ, TYPE_WALL, TYPE_WECHAT} from '../lib/constants'

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
    } else if (row.status === STATUS_FAILED || row.status === STATUS_UNKNOWN) {
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

// 下面 4 个函数是检测目标的严格校验规则，与 Go 侧 boce_targets.go 逐条对应。
// 校验已收口到 Go 的 ValidateTargets；这里只在没有 Wails bridge（npm run dev）时兜底，
// 不要再在别处写第二套规则。
function parseTargetEntries(text) {
  const seen = new Set()
  const invalidTargets = []
  const validTargets = []

  text
    .split(/[\s,;，；]+/)
    .map(item => item.trim().replace(/^['"]|['"]$/g, ''))
    .filter(Boolean)
    .forEach(item => {
      const normalized = normalizeTarget(item)
      if (!normalized) {
        invalidTargets.push(item)
        return
      }
      if (!seen.has(normalized)) {
        seen.add(normalized)
        validTargets.push(normalized)
      }
    })

  return {validTargets, invalidTargets}
}

function normalizeTarget(input) {
  const raw = input.trim()
  if (!raw || /[\s@]/.test(raw)) {
    return ''
  }

  let targetURL
  try {
    if (/^https?:\/\//i.test(raw)) {
      targetURL = new URL(raw)
    } else if (/^[a-z][a-z\d+.-]*:\/\//i.test(raw)) {
      return ''
    } else {
      targetURL = new URL(`http://${raw.replace(/^\/+|\/+$/g, '')}`)
    }
  } catch {
    return ''
  }

  const host = targetURL.hostname.toLowerCase()
  if (!isValidHost(host)) {
    return ''
  }
  return host
}

function isValidHost(host) {
  if (isValidIPv4(host)) {
    return true
  }
  if (host.length > 253 || !host.includes('.') || host.includes('..')) {
    return false
  }

  const labels = host.split('.')
  const topLabel = labels.at(-1)
  return labels.every(label => (
    label.length > 0 &&
    label.length <= 63 &&
    /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/i.test(label)
  )) && topLabel.length >= 2 && /[a-z]/i.test(topLabel)
}

function isValidIPv4(host) {
  const parts = host.split('.')
  return parts.length === 4 && parts.every(part => {
    if (!/^\d{1,3}$/.test(part)) {
      return false
    }
    const value = Number(part)
    return value >= 0 && value <= 255 && String(value) === part
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
    exportPath: '',
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
    path: '',
    message: `开发预览: 已导出 ${rows.length} 条当前列表数据`,
  })
}

export function ValidateTargets(text) {
  if (hasWailsBridge()) {
    return wailsValidateTargets(text)
  }

  // 开发预览没有 Go 后端：复用上面这套前端校验逻辑，返回与 Go 同形状的对象。
  const {validTargets, invalidTargets} = parseTargetEntries(text)
  return Promise.resolve({
    valid: validTargets,
    invalid: invalidTargets,
    message: '',
  })
}
