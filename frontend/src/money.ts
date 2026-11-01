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
