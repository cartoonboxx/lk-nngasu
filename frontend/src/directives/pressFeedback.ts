import type { ObjectDirective } from 'vue'

const cleanups = new WeakMap<HTMLElement, () => void>()

/** Immediate touch feedback without capturing the pointer or preventing native scrolling. */
export const pressFeedback: ObjectDirective<HTMLElement> = {
  mounted(element) {
    let pointer: { id: number; x: number; y: number } | null = null
    const reset = () => {
      pointer = null
      delete element.dataset.pressed
      window.removeEventListener('pointermove', move)
      window.removeEventListener('pointerup', end)
      window.removeEventListener('pointercancel', end)
      window.removeEventListener('scroll', reset, true)
      window.removeEventListener('blur', reset)
      document.removeEventListener('visibilitychange', reset)
    }
    const down = (event: PointerEvent) => {
      reset()
      if (
        event.pointerType === 'mouse' ||
        !event.isPrimary ||
        event.button !== 0 ||
        !element.matches('button, a, label') ||
        element.matches(':disabled, [aria-disabled="true"], [data-disabled="true"]') ||
        element.closest('[inert]') ||
        element.querySelector(':scope > input:disabled')
      )
        return
      pointer = { id: event.pointerId, x: event.clientX, y: event.clientY }
      element.dataset.pressed = 'true'
      window.addEventListener('pointermove', move, { passive: true })
      window.addEventListener('pointerup', end, { passive: true })
      window.addEventListener('pointercancel', end, { passive: true })
      window.addEventListener('scroll', reset, true)
      window.addEventListener('blur', reset)
      document.addEventListener('visibilitychange', reset)
    }
    const move = (event: PointerEvent) => {
      if (pointer?.id !== event.pointerId) return
      if (Math.hypot(event.clientX - pointer.x, event.clientY - pointer.y) > 10) reset()
    }
    const end = (event: PointerEvent) => {
      if (pointer?.id === event.pointerId) reset()
    }
    element.addEventListener('pointerdown', down, { passive: true })
    element.addEventListener('pointerleave', reset)
    cleanups.set(element, () => {
      reset()
      element.removeEventListener('pointerdown', down)
      element.removeEventListener('pointerleave', reset)
    })
  },
  updated(element) {
    if (
      element.matches(':disabled, [aria-disabled="true"], [data-disabled="true"]') ||
      element.querySelector(':scope > input:disabled')
    ) {
      delete element.dataset.pressed
    }
  },
  beforeUnmount(element) {
    cleanups.get(element)?.()
    cleanups.delete(element)
  },
}
