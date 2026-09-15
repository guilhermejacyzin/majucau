import { cloneElement, isValidElement, ReactElement, ReactNode, useId, useState } from 'react'
import { tooltipCatalog, type TooltipId } from '../tooltips'
type TooltipProps = { id: TooltipId; children: ReactNode }
export function Tooltip({ id, children }: TooltipProps) {
  const [visible, setVisible] = useState(false)
  const descriptionId = useId()
  const text = tooltipCatalog[id]
  const trigger = isValidElement(children) ? cloneElement(children as ReactElement<{ 'aria-describedby'?: string; tabIndex?: number }>, { 'aria-describedby': visible ? descriptionId : undefined, tabIndex: (children.props as { tabIndex?: number }).tabIndex ?? 0 }) : <span tabIndex={0} aria-describedby={visible ? descriptionId : undefined}>{children}</span>
  return <span className="tooltip" onMouseEnter={() => setVisible(true)} onMouseLeave={() => setVisible(false)} onFocusCapture={() => setVisible(true)} onBlurCapture={() => setVisible(false)} onKeyDown={(event) => { if (event.key === 'Escape') { event.preventDefault(); setVisible(false); (event.target as HTMLElement).blur() } }}>{trigger}{visible && <span id={descriptionId} className="tooltip-content" role="tooltip">{text}</span>}</span>
}
