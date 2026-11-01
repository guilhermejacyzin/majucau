import { describe, expect, it } from 'vitest'
import { formatBRL } from './money'

describe('formatBRL', () => {
  it('arredonda metade para cima somente na apresentação', () => {
    expect(formatBRL('100.0050')).toMatch(/^R\$\s*100,01$/)
    expect(formatBRL('100.0049')).toMatch(/^R\$\s*100,00$/)
    expect(formatBRL('-1.2350')).toMatch(/^-R\$\s*1,24$/)
  })

  it('preserva agrupamento de valores que ultrapassam a precisão segura de Number', () => {
    expect(formatBRL('999999999999999.9950')).toMatch(/^R\$\s*1\.000\.000\.000\.000\.000,00$/)
  })

  it('normaliza o sinal negativo quando o valor arredondado é zero', () => {
    expect(formatBRL('-0.0049')).toMatch(/^R\$\s*0,00$/)
    expect(formatBRL('-0.0050')).toMatch(/^-R\$\s*0,01$/)
  })

  it('não exibe texto inválido nem valores fora do contrato numeric(19,4)', () => {
    for (const value of ['', '1,25', '1e3', '1.23456', '1000000000000000', 'NaN']) {
      expect(formatBRL(value)).toBe('—')
    }
  })
})
