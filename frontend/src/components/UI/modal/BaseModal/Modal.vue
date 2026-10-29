<template>
  <Teleport to="body">
    <dialog
      ref="dialog"
      v-bind="$attrs"
      :class="[
        style.overlay,
        isPopover ? style.popover : null,
        entered ? style.entered : null,
        closing ? style.leaving : null,
        dragging ? style.dragging : null,
      ]"
      :style="modalStyle"
      @cancel.self.prevent="close"
      @keydown.esc.prevent.stop="close"
      @keydown.tab="trapFocus"
      @close="onNativeClose"
      @pointerdown="backdropPressed = $event.target === $event.currentTarget"
      @click="onBackdropClick"
    >
      <div ref="panel" :class="style.modal" tabindex="-1" autofocus>
        <div
          v-if="handle"
          ref="handle"
          :class="style.handle"
          @pointerdown="startDrag"
          @pointermove="moveDrag"
          @pointerup="finishDrag"
          @pointercancel="cancelDrag"
          @lostpointercapture="cancelDrag"
        />
        <div
          :class="style.content"
          @touchstart.passive="startContentTouch"
          @touchmove="moveContentTouch"
          @touchend="finishContentTouch"
          @touchcancel="cancelDrag"
        >
          <slot :close="close" :breakpoint="currentBreakpoint" :set-breakpoint="setBreakpoint" />
        </div>
      </div>
    </dialog>
  </Teleport>
</template>

<script lang="ts">
import { defineComponent, type CSSProperties, type PropType } from 'vue'
import screenSizes from '@/styles/breakpoints.module.scss'
import style from './BaseModal.module.scss'

// Keep the page locked until the last modal closes, including its exit animation.
const scrollLocks = new Set<HTMLDialogElement>()
let previousOverflow = ''

function unlockScroll(dialog: HTMLDialogElement) {
  if (scrollLocks.delete(dialog) && !scrollLocks.size) {
    document.body.style.overflow = previousOverflow
  }
}

export default defineComponent({
  name: 'Modal',
  inheritAttrs: false,
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    width: {
      type: [Number, String],
      default: 540,
    },
    height: {
      type: [Number, String],
      default: 'auto',
    },
    maxWidth: {
      type: [Number, String],
      default: '100%',
    },
    maxHeight: {
      type: [Number, String],
      default: '100%',
    },
    closeOnOverlay: {
      type: Boolean,
      default: true,
    },
    handle: {
      type: Boolean,
      default: true,
    },
    swipeToClose: {
      type: Boolean,
      default: true,
    },
    breakpoints: {
      type: Array as PropType<number[]>,
      default: () => [0, 1],
    },
    initialBreakpoint: {
      type: Number,
      default: 1,
    },
    anchor: {
      type: [Object, Function] as PropType<ModalAnchor>,
      default: null,
    },
  },
  emits: {
    'update:modelValue': (value: boolean) => typeof value === 'boolean',
    'breakpoint-change': (value: number) => typeof value === 'number',
    close: () => true,
  },
  data() {
    return {
      style,
      backdropPressed: false,
      entered: false,
      closing: false,
      isMobile: false,
      dragging: false,
      currentBreakpoint: 1,
      panelHeight: 0,
      dragOffset: 0,
      dragStartY: 0,
      dragStartOffset: 0,
      dragStartTime: 0,
      pointerId: null as number | null,
      contentTouch: null as ModalContentTouch | null,
      viewportHeight: 0,
      viewportTop: 0,
      popupX: 0,
      popupY: 0,
      mediaQuery: null as MediaQueryList | null,
      resizeObserver: null as ResizeObserver | null,
      enterFrame: 0,
      closeTimer: undefined as ReturnType<typeof setTimeout> | undefined,
    }
  },
  computed: {
    isPopover(): boolean {
      return Boolean(this.anchor && !this.isMobile)
    },
    sheetBreakpoints(): number[] {
      const values = [...new Set(this.breakpoints)]
        .filter((value) => Number.isFinite(value) && value >= 0 && value <= 1)
        .sort((a, b) => a - b)
      return values.some((value) => value > 0) ? values : [0, 1]
    },
    modalStyle(): CSSProperties {
      const offset = this.dragging
        ? this.dragOffset
        : this.panelHeight * (1 - this.currentBreakpoint)
      return {
        '--modal-width': typeof this.width === 'number' ? `${this.width}px` : this.width,
        '--modal-height': typeof this.height === 'number' ? `${this.height}px` : this.height,
        '--modal-max-width':
          typeof this.maxWidth === 'number' ? `${this.maxWidth}px` : this.maxWidth,
        '--modal-max-height':
          typeof this.maxHeight === 'number' ? `${this.maxHeight}px` : this.maxHeight,
        '--sheet-offset': `${offset}px`,
        '--modal-x': `${this.popupX}px`,
        '--modal-y': `${this.popupY}px`,
        ...(this.viewportHeight
          ? {
              '--modal-viewport-height': `${this.viewportHeight}px`,
              '--modal-viewport-top': `${this.viewportTop}px`,
            }
          : {}),
      }
    },
  },
  watch: {
    modelValue() {
      this.syncDialog()
    },
    sheetBreakpoints() {
      this.setBreakpoint(this.currentBreakpoint)
    },
    async isPopover() {
      const dialog = this.$refs.dialog as HTMLDialogElement | undefined
      if (!dialog?.open || !this.modelValue) return
      await this.$nextTick()
      if (!this.modelValue || !dialog.open) return
      const focused = document.activeElement as HTMLElement | null
      dialog.close()
      unlockScroll(dialog)
      this.presentDialog()
      if (focused && dialog.contains(focused)) focused.focus({ preventScroll: true })
    },
  },
  mounted() {
    this.mediaQuery = window.matchMedia(`(max-width: ${screenSizes.phoneMax})`)
    this.isMobile = this.mediaQuery.matches
    this.mediaQuery.addEventListener('change', this.onViewportChange)
    window.visualViewport?.addEventListener('resize', this.updateViewport)
    window.visualViewport?.addEventListener('scroll', this.updateViewport)
    window.addEventListener('resize', this.updateViewport)
    window.addEventListener('scroll', this.updatePosition, true)
    document.addEventListener('pointerdown', this.onOutsidePointer)
    if (typeof ResizeObserver !== 'undefined') {
      this.resizeObserver = new ResizeObserver(this.measurePanel)
      this.resizeObserver.observe(this.$refs.panel as HTMLElement)
    }
    this.updateViewport()
    this.syncDialog()
  },
  beforeUnmount() {
    this.clearAnimation()
    this.cancelDrag()
    this.resizeObserver?.disconnect()
    this.mediaQuery?.removeEventListener('change', this.onViewportChange)
    window.visualViewport?.removeEventListener('resize', this.updateViewport)
    window.visualViewport?.removeEventListener('scroll', this.updateViewport)
    window.removeEventListener('resize', this.updateViewport)
    window.removeEventListener('scroll', this.updatePosition, true)
    document.removeEventListener('pointerdown', this.onOutsidePointer)
    const dialog = this.$refs.dialog as HTMLDialogElement
    if (dialog.open) dialog.close()
    unlockScroll(dialog)
  },
  methods: {
    clearAnimation() {
      cancelAnimationFrame(this.enterFrame)
      clearTimeout(this.closeTimer)
      this.closeTimer = undefined
    },
    syncDialog() {
      const dialog = this.$refs.dialog as HTMLDialogElement | undefined
      if (!dialog) return
      if (this.modelValue) {
        this.clearAnimation()
        this.closing = false
        if (!dialog.open) {
          this.currentBreakpoint = this.nearestBreakpoint(this.initialBreakpoint, false)
          this.entered = false
          this.presentDialog()
          this.measurePanel()
          // Commit the starting position before the opening transition.
          void dialog.offsetHeight
          this.enterFrame = requestAnimationFrame(() => {
            this.enterFrame = requestAnimationFrame(() => {
              this.entered = true
            })
          })
        } else {
          this.entered = true
        }
      } else if (dialog.open && !this.closing) {
        this.clearAnimation()
        this.cancelDrag()
        this.closing = true
        const duration = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 140
        this.closeTimer = setTimeout(() => {
          this.closeTimer = undefined
          if (!this.modelValue && dialog.open) {
            dialog.close()
            unlockScroll(dialog)
          }
        }, duration)
      }
    },
    close() {
      this.$emit('update:modelValue', false)
    },
    trapFocus(event: KeyboardEvent) {
      if (this.isPopover) return
      const panel = this.$refs.panel as HTMLElement
      const controls = Array.from(
        panel.querySelectorAll<HTMLElement>('button, a[href], input, select, textarea, [tabindex]'),
      ).filter(
        (element) =>
          element.tabIndex >= 0 &&
          !element.matches(':disabled') &&
          !element.closest('[inert]') &&
          element.getClientRects().length,
      )
      event.preventDefault()
      if (!controls.length) {
        panel.focus({ preventScroll: true })
        return
      }
      // Explicit navigation also includes buttons WebKit may skip with native Tab behavior.
      const index = controls.indexOf(document.activeElement as HTMLElement)
      const next = event.shiftKey
        ? index <= 0
          ? controls.length - 1
          : index - 1
        : (index + 1) % controls.length
      controls[next]?.focus({ preventScroll: true })
    },
    presentDialog() {
      const dialog = this.$refs.dialog as HTMLDialogElement
      if (this.isPopover) {
        dialog.show()
        this.updatePosition()
      } else {
        dialog.showModal()
        if (!scrollLocks.size) {
          previousOverflow = document.body.style.overflow
          document.body.style.overflow = 'hidden'
        }
        scrollLocks.add(dialog)
      }
    },
    updatePosition() {
      const dialog = this.$refs.dialog as HTMLDialogElement | undefined
      if (!this.isPopover || !dialog?.open || !this.anchor) return
      const element = this.getAnchor()
      if (!element) return
      const anchor = element.getBoundingClientRect()
      const right = anchor.right + 8
      const horizontal =
        right + dialog.offsetWidth <= window.innerWidth - 12
          ? right
          : anchor.left - dialog.offsetWidth - 8
      this.popupX = Math.max(12, Math.min(horizontal, window.innerWidth - dialog.offsetWidth - 12))
      this.popupY = Math.max(
        12,
        Math.min(anchor.top, window.innerHeight - dialog.offsetHeight - 12),
      )
    },
    onOutsidePointer(event: PointerEvent) {
      if (!this.isPopover || !this.modelValue || !this.closeOnOverlay) return
      const target = event.target as Node
      if (
        !(this.$refs.dialog as HTMLDialogElement).contains(target) &&
        !this.getAnchor()?.contains(target)
      )
        this.close()
    },
    getAnchor(): HTMLElement | null {
      return typeof this.anchor === 'function' ? this.anchor() : this.anchor
    },
    onBackdropClick(event: MouseEvent) {
      if (this.closeOnOverlay && this.backdropPressed && event.target === event.currentTarget)
        this.close()
      this.backdropPressed = false
    },
    onNativeClose() {
      const dialog = this.$refs.dialog as HTMLDialogElement
      if (dialog.open) return
      this.clearAnimation()
      this.cancelDrag()
      this.entered = false
      this.closing = false
      unlockScroll(dialog)
      if (this.modelValue) this.close()
      this.$emit('close')
    },
    updateViewport() {
      this.viewportHeight = window.visualViewport?.height ?? window.innerHeight
      this.viewportTop = window.visualViewport?.offsetTop ?? 0
      void this.$nextTick(this.updatePosition)
    },
    onViewportChange() {
      this.cancelDrag()
      this.isMobile = this.mediaQuery?.matches ?? false
      this.updateViewport()
    },
    measurePanel() {
      this.panelHeight = (this.$refs.panel as HTMLElement).offsetHeight
      this.updatePosition()
    },
    nearestBreakpoint(value: number, includeClosed?: boolean): number {
      const options = this.sheetBreakpoints.filter(
        (point) => (includeClosed ?? this.swipeToClose) || point > 0,
      )
      const target = Number.isFinite(value) ? value : 1
      return options.reduce(
        (nearest, point) =>
          Math.abs(point - target) < Math.abs(nearest - target) ? point : nearest,
        options[0]!,
      )
    },
    setBreakpoint(value: number) {
      const next = this.nearestBreakpoint(value)
      if (!next) {
        if (this.isMobile && this.modelValue) this.close()
        return
      }
      this.currentBreakpoint = next
      this.$emit('breakpoint-change', next)
    },
    startDrag(event: PointerEvent) {
      if (!this.isMobile || this.closing || !event.isPrimary || event.button !== 0) return
      this.beginDrag(event.clientY, event.timeStamp)
      if (!this.dragging) return
      this.pointerId = event.pointerId
      ;(event.currentTarget as HTMLElement).setPointerCapture(event.pointerId)
    },
    beginDrag(y: number, time: number) {
      this.measurePanel()
      if (!this.panelHeight) return
      this.dragStartY = y
      this.dragStartTime = time
      this.dragStartOffset = this.panelHeight * (1 - this.currentBreakpoint)
      this.dragOffset = this.dragStartOffset
      this.dragging = true
    },
    moveDrag(event: PointerEvent) {
      if (!this.dragging || event.pointerId !== this.pointerId) return
      this.updateDrag(event.clientY)
    },
    updateDrag(y: number) {
      const lowest = this.swipeToClose
        ? this.sheetBreakpoints[0]!
        : this.sheetBreakpoints.find((point) => point > 0)!
      this.dragOffset = Math.max(
        0,
        Math.min(this.panelHeight * (1 - lowest), this.dragStartOffset + y - this.dragStartY),
      )
    },
    finishDrag(event: PointerEvent) {
      if (!this.dragging || event.pointerId !== this.pointerId) return
      this.endDrag(event.clientY, event.timeStamp)
    },
    endDrag(y: number, time: number) {
      this.updateDrag(y)
      const distance = y - this.dragStartY
      const velocity = distance / Math.max(16, time - this.dragStartTime)
      const momentum = Math.max(
        -this.panelHeight * 0.2,
        Math.min(this.panelHeight * 0.2, velocity * 120),
      )
      const projected = this.dragOffset + (Math.abs(distance) > 48 ? momentum : 0)
      const next = 1 - projected / this.panelHeight
      this.cancelDrag()
      this.setBreakpoint(next)
    },
    startContentTouch(event: TouchEvent) {
      if (!this.isMobile || this.closing || this.dragging || event.touches.length !== 1) {
        this.cancelDrag()
        return
      }
      this.contentTouch = null
      // Let every scrollable ancestor (including textareas) consume its own scroll first.
      const content = event.currentTarget as HTMLElement
      let element = event.target instanceof Element ? event.target : null
      if (element?.closest('input, select, [contenteditable="true"]')) return
      while (element) {
        const overflow = getComputedStyle(element).overflowY
        if (/(auto|scroll)/.test(overflow) && element.scrollTop > 0) return
        if (element === content) break
        element = element.parentElement
      }
      const touch = event.touches[0]!
      this.contentTouch = {
        id: touch.identifier,
        x: touch.clientX,
        y: touch.clientY,
        time: event.timeStamp,
      }
    },
    moveContentTouch(event: TouchEvent) {
      const start = this.contentTouch
      if (!start) return
      const touch = Array.from(event.touches).find((touch) => touch.identifier === start.id)
      if (!touch || event.touches.length !== 1 || !event.cancelable) {
        this.cancelDrag()
        return
      }
      if (!this.dragging) {
        const dy = touch.clientY - start.y
        const dx = Math.abs(touch.clientX - start.x)
        if (Math.max(Math.abs(dy), dx) < 8) return
        if (dy <= 0 || dx >= dy) {
          this.contentTouch = null
          return
        }
        this.beginDrag(start.y, start.time)
      }
      if (this.dragging) {
        event.preventDefault()
        this.updateDrag(touch.clientY)
      }
    },
    finishContentTouch(event: TouchEvent) {
      const start = this.contentTouch
      if (!start) return
      const touch = Array.from(event.changedTouches).find((touch) => touch.identifier === start.id)
      if (touch && this.dragging) this.endDrag(touch.clientY, event.timeStamp)
      else this.cancelDrag()
    },
    cancelDrag() {
      this.contentTouch = null
      this.dragging = false
      const handle = this.$refs.handle as HTMLElement | undefined
      if (this.pointerId !== null && handle?.hasPointerCapture(this.pointerId))
        handle.releasePointerCapture(this.pointerId)
      this.pointerId = null
    },
  },
})
</script>
