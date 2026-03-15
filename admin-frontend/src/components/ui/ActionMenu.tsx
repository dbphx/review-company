import { ReactNode, useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"

interface ActionMenuProps {
  children: ReactNode
  menuClassName?: string
  placement?: "top" | "bottom"
}

export default function ActionMenu({ children, menuClassName = "", placement = "top" }: ActionMenuProps) {
  const [open, setOpen] = useState(false)
  const [menuPos, setMenuPos] = useState({ top: 0, left: 0 })
  const rootRef = useRef<HTMLDivElement | null>(null)
  const triggerRef = useRef<HTMLButtonElement | null>(null)
  const menuRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!open) return

    const onPointerDown = (event: MouseEvent | TouchEvent) => {
      const target = event.target as Node
      const clickedTrigger = rootRef.current?.contains(target)
      const clickedMenu = menuRef.current?.contains(target)
      if (!clickedTrigger && !clickedMenu) {
        setOpen(false)
      }
    }

    const onEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setOpen(false)
      }
    }

    document.addEventListener("mousedown", onPointerDown)
    document.addEventListener("touchstart", onPointerDown)
    document.addEventListener("keydown", onEscape)

    return () => {
      document.removeEventListener("mousedown", onPointerDown)
      document.removeEventListener("touchstart", onPointerDown)
      document.removeEventListener("keydown", onEscape)
    }
  }, [open])

  useEffect(() => {
    if (!open) return

    const resolvePosition = () => {
      const triggerRect = triggerRef.current?.getBoundingClientRect()
      const menuEl = menuRef.current
      if (!triggerRect || !menuEl) return

      const gap = 8
      const viewportPad = 8
      const menuWidth = menuEl.offsetWidth || 180
      const menuHeight = menuEl.offsetHeight || 220

      let left = triggerRect.right - menuWidth
      left = Math.max(viewportPad, Math.min(left, window.innerWidth - menuWidth - viewportPad))

      const topPreferred = placement === "top"
        ? triggerRect.top - menuHeight - gap
        : triggerRect.bottom + gap
      const topFallback = placement === "top"
        ? triggerRect.bottom + gap
        : triggerRect.top - menuHeight - gap

      const preferredFits = topPreferred >= viewportPad && topPreferred + menuHeight <= window.innerHeight - viewportPad
      const fallbackFits = topFallback >= viewportPad && topFallback + menuHeight <= window.innerHeight - viewportPad

      let top = topPreferred
      if (!preferredFits && fallbackFits) {
        top = topFallback
      }
      if (!preferredFits && !fallbackFits) {
        top = Math.max(viewportPad, Math.min(topPreferred, window.innerHeight - menuHeight - viewportPad))
      }

      setMenuPos({ top, left })
    }

    const raf = requestAnimationFrame(resolvePosition)
    window.addEventListener("resize", resolvePosition)
    window.addEventListener("scroll", resolvePosition, true)

    return () => {
      cancelAnimationFrame(raf)
      window.removeEventListener("resize", resolvePosition)
      window.removeEventListener("scroll", resolvePosition, true)
    }
  }, [open, placement])

  return (
    <div ref={rootRef} className="relative inline-block">
      <button
        ref={triggerRef}
        type="button"
        className="inline-flex items-center justify-center w-9 h-9 rounded-lg border border-slate-200 bg-white text-slate-700 hover:bg-slate-50"
        onClick={() => setOpen((v) => !v)}
      >
        ...
      </button>

      {open && createPortal(
        <div
          ref={menuRef}
          style={{ top: `${menuPos.top}px`, left: `${menuPos.left}px` }}
          className={`fixed z-[100] bg-white border rounded-lg shadow-lg p-1 max-h-[50vh] overflow-y-auto ${menuClassName}`}
          onClick={() => setOpen(false)}
        >
          {children}
        </div>,
        document.body
      )}
    </div>
  )
}
