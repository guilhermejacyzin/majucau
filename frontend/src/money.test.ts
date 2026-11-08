import { describe, expect, it } from 'vitest'
import { formatBRL, formatBRLDelta, formatPercentVariation } from './money'

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

describe('variações semanais monetárias', () => {
  it('calcula percentuais de entradas e de saídas por magnitude com casas exatas', () => {
    expect(formatPercentVariation('1500.0000', '1250.0000')).toBe('+20,0% vs. semana anterior')
    expect(formatPercentVariation('-600.0000', '-800.0000', true)).toBe('−25,0% vs. semana anterior')
  })

  it('não inventa percentual quando a semana anterior é zero e calcula diferença do saldo em BRL exato', () => {
    expect(formatPercentVariation('100.0000', '0.0000')).toBeUndefined()
    expect(formatBRLDelta('-76683.2000', '-83776.1500')).toMatch(/^\+R\$\s*7\.092,95$/)
  })
})
