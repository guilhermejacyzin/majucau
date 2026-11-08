const brlFormatter = new Intl.NumberFormat('pt-BR', {
  style: 'currency',
  currency: 'BRL',
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
})

const integerFormatter = new Intl.NumberFormat('pt-BR', {
  maximumFractionDigits: 0,
})

const numericParts = new Set(['integer', 'group', 'decimal', 'fraction'])

function parseMoneyUnits(value: string): bigint | undefined {
  const normalized = value.trim()
  if (!/^-?\d{1,15}(?:\.\d{1,4})?$/.test(normalized)) return undefined
  const negative = normalized.startsWith('-')
  const unsigned = negative ? normalized.slice(1) : normalized
  const [whole, fraction = ''] = unsigned.split('.')
  const units = BigInt(whole) * 10_000n + BigInt(fraction.padEnd(4, '0'))
  return negative ? -units : units
}

/** Exact percentage change, rounded to one decimal place for display. */
export function formatPercentVariation(current: string, previous: string, compareMagnitude = false): string | undefined {
  let currentUnits = parseMoneyUnits(current)
  let previousUnits = parseMoneyUnits(previous)
  if (currentUnits === undefined || previousUnits === undefined) return undefined
  if (compareMagnitude) {
    currentUnits = currentUnits < 0n ? -currentUnits : currentUnits
    previousUnits = previousUnits < 0n ? -previousUnits : previousUnits
  }
  if (previousUnits === 0n) return undefined

  const numerator = (currentUnits - previousUnits) * 1000n
  const denominator = previousUnits < 0n ? -previousUnits : previousUnits
  const isNegative = numerator < 0n
  const absoluteNumerator = isNegative ? -numerator : numerator
  let tenths = absoluteNumerator / denominator
  if ((absoluteNumerator % denominator) * 2n >= denominator) tenths += 1n
  if (tenths === 0n) return '0,0% vs. semana anterior'

  const whole = integerFormatter.format(tenths / 10n)
  const fraction = (tenths % 10n).toString()
  const decimalSeparator = brlFormatter.formatToParts(1).find((part) => part.type === 'decimal')?.value ?? ','
  return `${isNegative ? '−' : '+'}${whole}${decimalSeparator}${fraction}% vs. semana anterior`
}

/** Exact signed BRL difference between two numeric(19,4) strings. */
export function formatBRLDelta(current: string, previous: string): string | undefined {
  const currentUnits = parseMoneyUnits(current)
  const previousUnits = parseMoneyUnits(previous)
  if (currentUnits === undefined || previousUnits === undefined) return undefined
  const delta = currentUnits - previousUnits
  const sign = delta > 0n ? '+' : delta < 0n ? '−' : ''
  const absolute = delta < 0n ? -delta : delta
  const whole = absolute / 10_000n
  const fraction = (absolute % 10_000n).toString().padStart(4, '0').replace(/0+$/, '')
  const formatted = formatBRL(fraction ? `${whole}.${fraction}` : whole.toString())
  return `${sign}${formatted}`
}

/** Formats a numeric(19,4) string as BRL without converting it through float. */
export function formatBRL(value: string): string {
  const normalized = value.trim()
  if (!/^-?\d{1,15}(?:\.\d{1,4})?$/.test(normalized)) return '—'

  const negativeInput = normalized.startsWith('-')
  const unsigned = negativeInput ? normalized.slice(1) : normalized
  const [wholeText, fractionText = ''] = unsigned.split('.')
  let cents = BigInt(wholeText) * 100n + BigInt(fractionText.padEnd(2, '0').slice(0, 2))
  if ((fractionText[2] ?? '0') >= '5') cents += 1n

  const isNegative = negativeInput && cents > 0n
  const whole = cents / 100n
  const fraction = (cents % 100n).toString().padStart(2, '0')
  const templateParts = brlFormatter.formatToParts(isNegative ? -1 : 1)
  const firstNumericPart = templateParts.findIndex((part) => numericParts.has(part.type))
  let lastNumericPart = firstNumericPart
  for (let index = firstNumericPart + 1; index < templateParts.length; index += 1) {
    if (numericParts.has(templateParts[index].type)) lastNumericPart = index
  }

  const prefix = templateParts.slice(0, firstNumericPart).map((part) => part.value).join('')
  const suffix = templateParts.slice(lastNumericPart + 1).map((part) => part.value).join('')
  const decimalSeparator = templateParts.find((part) => part.type === 'decimal')?.value ?? ','
  return `${prefix}${integerFormatter.format(whole)}${decimalSeparator}${fraction}${suffix}`
}
